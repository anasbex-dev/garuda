package raknet

import (
    "encoding/binary"
    "fmt"
    "log"
    "net"
    "sort"
    "time"
)

type ReliableManager struct {
    session        *Session
    sendSequence   uint32
    receiveSequence uint32
    sendQueue      map[uint32]*PendingPacket
    receiveQueue   map[uint32]*EncapsulatedPacket
    ackQueue       []uint32
    nackQueue      []uint32
    mutex          sync.RWMutex
}

type PendingPacket struct {
    Packet      *EncapsulatedPacket
    SendTime    time.Time
    Retries     int
    Acknowledged bool
}

func NewReliableManager(session *Session) *ReliableManager {
    return &ReliableManager{
        session:      session,
        sendQueue:    make(map[uint32]*PendingPacket),
        receiveQueue: make(map[uint32]*EncapsulatedPacket),
        ackQueue:     make([]uint32, 0),
        nackQueue:    make([]uint32, 0),
    }
}

func (rm *ReliableManager) SendPacket(packet *EncapsulatedPacket, reliability PacketReliability) error {
    rm.mutex.Lock()
    defer rm.mutex.Unlock()

    // Set sequence numbers based on reliability
    switch reliability {
    case ReliabilityReliable:
        packet.MessageIndex = rm.sendSequence
        rm.sendSequence++
    case ReliabilityReliableOrdered:
        packet.MessageIndex = rm.sendSequence
        packet.OrderIndex = rm.sendSequence
        rm.sendSequence++
    case ReliabilityReliableSequenced:
        packet.MessageIndex = rm.sendSequence
        packet.SequenceIndex = rm.sendSequence
        rm.sendSequence++
    }

    // Create pending packet
    pending := &PendingPacket{
        Packet:   packet,
        SendTime: time.Now(),
        Retries:  0,
    }

    rm.sendQueue[packet.MessageIndex] = pending

    // Send immediately
    return rm.sendDatagram([]*EncapsulatedPacket{packet})
}

func (rm *ReliableManager) sendDatagram(packets []*EncapsulatedPacket) error {
    datagram := &Datagram{
        SequenceNumber: rm.sendSequence,
        Packets:        packets,
    }

    rm.sendSequence++

    data, err := rm.encodeDatagram(datagram)
    if err != nil {
        return err
    }

    return rm.session.server.sendRaw(rm.session.address, data)
}

func (rm *ReliableManager) encodeDatagram(datagram *Datagram) ([]byte, error) {
    // Calculate total size
    totalSize := 4 // sequence number
    for _, packet := range datagram.Packets {
        totalSize += rm.getEncapsulatedPacketSize(packet)
    }

    data := make([]byte, totalSize)
    offset := 0

    // Write sequence number
    binary.BigEndian.PutUint32(data[offset:offset+4], datagram.SequenceNumber)
    offset += 4

    // Write encapsulated packets
    for _, packet := range datagram.Packets {
        packetSize := rm.encodeEncapsulatedPacket(data[offset:], packet)
        offset += packetSize
    }

    return data[:offset], nil
}

func (rm *ReliableManager) getEncapsulatedPacketSize(packet *EncapsulatedPacket) int {
    size := 1 // flags
    if packet.Reliability == ReliabilityReliable ||
        packet.Reliability == ReliabilityReliableOrdered ||
        packet.Reliability == ReliabilityReliableSequenced {
        size += 3 // message index
    }
    if packet.Reliability == ReliabilityReliableOrdered {
        size += 4 // order index + order channel
    }
    if packet.Reliability == ReliabilityReliableSequenced {
        size += 4 // sequence index
    }
    if packet.HasSplit {
        size += 10 // split info
    }
    size += 2 // length
    size += len(packet.Data)
    return size
}

func (rm *ReliableManager) encodeEncapsulatedPacket(buffer []byte, packet *EncapsulatedPacket) int {
    offset := 0

    // Flags
    flags := byte(packet.Reliability) << 5
    if packet.HasSplit {
        flags |= 0x10
    }
    buffer[offset] = flags
    offset++

    // Length
    length := uint16(len(packet.Data) << 3)
    binary.BigEndian.PutUint16(buffer[offset:offset+2], length)
    offset += 2

    // Reliable message index
    if packet.Reliability == ReliabilityReliable ||
        packet.Reliability == ReliabilityReliableOrdered ||
        packet.Reliability == ReliabilityReliableSequenced {
        // 24-bit message index
        buffer[offset] = byte(packet.MessageIndex >> 16)
        buffer[offset+1] = byte(packet.MessageIndex >> 8)
        buffer[offset+2] = byte(packet.MessageIndex)
        offset += 3
    }

    // Sequencing
    if packet.Reliability == ReliabilityReliableOrdered {
        binary.BigEndian.PutUint32(buffer[offset:offset+4], packet.OrderIndex)
        offset += 4
        buffer[offset] = packet.OrderChannel
        offset++
    }

    if packet.Reliability == ReliabilityReliableSequenced {
        binary.BigEndian.PutUint32(buffer[offset:offset+4], packet.SequenceIndex)
        offset += 4
    }

    // Split information
    if packet.HasSplit {
        binary.BigEndian.PutUint32(buffer[offset:offset+4], packet.SplitCount)
        offset += 4
        binary.BigEndian.PutUint16(buffer[offset:offset+2], packet.SplitID)
        offset += 2
        binary.BigEndian.PutUint32(buffer[offset:offset+4], packet.SplitIndex)
        offset += 4
    }

    // Data
    copy(buffer[offset:], packet.Data)
    offset += len(packet.Data)

    return offset
}

func (rm *ReliableManager) HandleAck(ackPacket []byte) {
    rm.mutex.Lock()
    defer rm.mutex.Unlock()

    if len(ackPacket) < 3 {
        return
    }

    packetID := ackPacket[0]
    if packetID != 0xc0 && packetID != 0xa0 {
        return
    }

    count := binary.BigEndian.Uint16(ackPacket[1:3])
    if len(ackPacket) < 3+int(count)*4 {
        return
    }

    offset := 3
    for i := 0; i < int(count); i++ {
        sequence := binary.BigEndian.Uint32(ackPacket[offset : offset+4])
        offset += 4

        if packetID == 0xc0 { // ACK
            rm.handleSingleAck(sequence)
        } else { // NACK
            rm.handleSingleNack(sequence)
        }
    }
}

func (rm *ReliableManager) handleSingleAck(sequence uint32) {
    if pending, exists := rm.sendQueue[sequence]; exists {
        pending.Acknowledged = true
        delete(rm.sendQueue, sequence)
    }
}

func (rm *ReliableManager) handleSingleNack(sequence uint32) {
    if pending, exists := rm.sendQueue[sequence]; exists {
        // Resend the packet
        packets := []*EncapsulatedPacket{pending.Packet}
        rm.sendDatagram(packets)
        pending.Retries++
        pending.SendTime = time.Now()
    }
}

func (rm *ReliableManager) ProcessReceivedDatagram(data []byte) []*EncapsulatedPacket {
    if len(data) < 4 {
        return nil
    }

    sequence := binary.BigEndian.Uint32(data[0:4])
    rm.mutex.Lock()
    rm.ackQueue = append(rm.ackQueue, sequence)
    rm.mutex.Unlock()

    packets := rm.decodeDatagram(data[4:])
    
    // Process reliable packets
    var reliablePackets []*EncapsulatedPacket
    for _, packet := range packets {
        if packet.Reliability == ReliabilityReliable ||
            packet.Reliability == ReliabilityReliableOrdered ||
            packet.Reliability == ReliabilityReliableSequenced {
            reliablePackets = append(reliablePackets, packet)
        }
    }

    return packets
}

func (rm *ReliableManager) decodeDatagram(data []byte) []*EncapsulatedPacket {
    var packets []*EncapsulatedPacket
    offset := 0

    for offset < len(data) {
        packet, bytesRead := rm.decodeEncapsulatedPacket(data[offset:])
        if packet == nil {
            break
        }
        packets = append(packets, packet)
        offset += bytesRead
    }

    return packets
}

func (rm *ReliableManager) decodeEncapsulatedPacket(data []byte) (*EncapsulatedPacket, int) {
    if len(data) < 3 {
        return nil, 0
    }

    packet := &EncapsulatedPacket{}
    offset := 0

    // Flags
    flags := data[offset]
    packet.Reliability = PacketReliability((flags & 0xE0) >> 5)
    packet.HasSplit = (flags & 0x10) != 0
    offset++

    // Length
    length := binary.BigEndian.Uint16(data[offset:offset+2]) >> 3
    offset += 2

    // Message index (24-bit)
    if packet.Reliability == ReliabilityReliable ||
        packet.Reliability == ReliabilityReliableOrdered ||
        packet.Reliability == ReliabilityReliableSequenced {
        packet.MessageIndex = uint32(data[offset])<<16 | uint32(data[offset+1])<<8 | uint32(data[offset+2])
        offset += 3
    }

    // Order index
    if packet.Reliability == ReliabilityReliableOrdered {
        packet.OrderIndex = binary.BigEndian.Uint32(data[offset : offset+4])
        offset += 4
        packet.OrderChannel = data[offset]
        offset++
    }

    // Sequence index
    if packet.Reliability == ReliabilityReliableSequenced {
        packet.SequenceIndex = binary.BigEndian.Uint32(data[offset : offset+4])
        offset += 4
    }

    // Split information
    if packet.HasSplit {
        packet.SplitCount = binary.BigEndian.Uint32(data[offset : offset+4])
        offset += 4
        packet.SplitID = binary.BigEndian.Uint16(data[offset : offset+2])
        offset += 2
        packet.SplitIndex = binary.BigEndian.Uint32(data[offset : offset+4])
        offset += 4
    }

    // Data
    if int(length) > len(data)-offset {
        return nil, 0
    }
    packet.Data = make([]byte, length)
    copy(packet.Data, data[offset:offset+int(length)])
    offset += int(length)

    return packet, offset
}

func (rm *ReliableManager) SendAcks() {
    rm.mutex.Lock()
    defer rm.mutex.Unlock()

    if len(rm.ackQueue) == 0 {
        return
    }

    // Sort and remove duplicates
    sort.Slice(rm.ackQueue, func(i, j int) bool {
        return rm.ackQueue[i] < rm.ackQueue[j]
    })

    uniqueAcks := make([]uint32, 0)
    for i, ack := range rm.ackQueue {
        if i == 0 || ack != rm.ackQueue[i-1] {
            uniqueAcks = append(uniqueAcks, ack)
        }
    }

    // Create ACK packet
    packet := make([]byte, 3+len(uniqueAcks)*4)
    packet[0] = 0xc0 // ACK packet ID
    binary.BigEndian.PutUint16(packet[1:3], uint16(len(uniqueAcks)))

    offset := 3
    for _, ack := range uniqueAcks {
        binary.BigEndian.PutUint32(packet[offset:offset+4], ack)
        offset += 4
    }

    rm.session.server.sendRaw(rm.session.address, packet)
    rm.ackQueue = rm.ackQueue[:0]
}

func (rm *ReliableManager) CheckTimeouts() {
    rm.mutex.Lock()
    defer rm.mutex.Unlock()

    now := time.Now()
    timeout := 5 * time.Second

    for seq, pending := range rm.sendQueue {
        if now.Sub(pending.SendTime) > timeout && pending.Retries < 3 {
            // Resend packet
            packets := []*EncapsulatedPacket{pending.Packet}
            rm.sendDatagram(packets)
            pending.Retries++
            pending.SendTime = now
        } else if pending.Retries >= 3 {
            // Give up on this packet
            delete(rm.sendQueue, seq)
            log.Printf("Packet %d timed out after %d retries", seq, pending.Retries)
        }
    }
}