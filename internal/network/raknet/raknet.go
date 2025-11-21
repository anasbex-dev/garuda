package raknet

import (
    "bytes"
    "encoding/binary"
    "garuda/pkg/utils"
)

var (
    Magic = [16]byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}
)

func ReadPacket(data []byte) (*Packet, error) {
    if len(data) < 1 {
        return nil, nil
    }
    
    packetID := data[0]
    packetData := data[1:]
    
    return &Packet{
        ID:   packetID,
        Data: packetData,
    }, nil
}

func WritePacket(packet *Packet) []byte {
    return append([]byte{packet.ID}, packet.Data...)
}

func DecodeUnconnectedPing(data []byte) (*UnconnectedPing, error) {
    stream := utils.NewBinaryStream(data)
    
    pingID := int64(stream.ReadUint64())
    magic := [16]byte{}
    copy(magic[:], stream.ReadBytes(16))
    clientGUID := int64(stream.ReadUint64())
    
    return &UnconnectedPing{
        PingID:    pingID,
        Magic:     magic,
        ClientGUID: clientGUID,
    }, nil
}

func EncodeUnconnectedPong(pong *UnconnectedPong) []byte {
    stream := utils.NewBinaryStream(nil)
    stream.WriteByte(ID_UNCONNECTED_PONG)
    stream.WriteUint64(uint64(pong.PingID))
    stream.WriteUint64(uint64(pong.ServerGUID))
    stream.WriteBytes(Magic[:])
    stream.WriteUint16(uint16(len(pong.MOTD)))
    stream.WriteBytes([]byte(pong.MOTD))
    
    return stream.Bytes()
}

func DecodeOpenConnectionRequest1(data []byte) (*OpenConnectionRequest1, error) {
    stream := utils.NewBinaryStream(data[1:]) // Skip packet ID
    
    magic := [16]byte{}
    copy(magic[:], stream.ReadBytes(16))
    protocol := stream.ReadByte()
    mtuSize := len(data) + 18 // +18 for IP header
    
    return &OpenConnectionRequest1{
        Magic:    magic,
        Protocol: protocol,
        MTUSize:  mtuSize,
    }, nil
}

func EncodeOpenConnectionReply1(reply *OpenConnectionReply1) []byte {
    stream := utils.NewBinaryStream(nil)
    stream.WriteByte(ID_OPEN_CONNECTION_REPLY)
    stream.WriteBytes(Magic[:])
    stream.WriteUint64(uint64(reply.ServerGUID))
    
    // Use security (always false for now)
    if reply.UseSecurity {
        stream.WriteByte(1)
    } else {
        stream.WriteByte(0)
    }
    
    stream.WriteUint16(uint16(reply.MTUSize))
    
    return stream.Bytes()
}

func VerifyMagic(magic [16]byte) bool {
    return bytes.Equal(magic[:], Magic[:])
}