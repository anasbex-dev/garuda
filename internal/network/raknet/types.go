package raknet

type PacketReliability int

const (
    ReliabilityUnreliable PacketReliability = iota
    ReliabilityUnreliableSequenced
    ReliabilityReliable
    ReliabilityReliableOrdered
    ReliabilityReliableSequenced
)

type EncapsulatedPacket struct {
    Reliability    PacketReliability
    HasSplit       bool
    MessageIndex   uint32
    SequenceIndex  uint32
    OrderIndex     uint32
    OrderChannel   byte
    SplitCount     uint32
    SplitID        uint16
    SplitIndex     uint32
    Data           []byte
}

type Datagram struct {
    SequenceNumber uint32
    Packets        []*EncapsulatedPacket
}