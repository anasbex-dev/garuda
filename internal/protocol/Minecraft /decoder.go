package minecraft

import (
    "bytes"
    "encoding/binary"
    "fmt"
)

type PacketDecoder struct {
    buffer *bytes.Buffer
}

func NewPacketDecoder(data []byte) *PacketDecoder {
    return &PacketDecoder{
        buffer: bytes.NewBuffer(data),
    }
}

func (d *PacketDecoder) ReadByte() (byte, error) {
    return d.buffer.ReadByte()
}

func (d *PacketDecoder) ReadBytes(n int) ([]byte, error) {
    data := make([]byte, n)
    _, err := d.buffer.Read(data)
    return data, err
}

func (d *PacketDecoder) ReadBool() (bool, error) {
    b, err := d.ReadByte()
    return b != 0, err
}

func (d *PacketDecoder) ReadInt16() (int16, error) {
    var value int16
    err := binary.Read(d.buffer, binary.BigEndian, &value)
    return value, err
}

func (d *PacketDecoder) ReadUint16() (uint16, error) {
    var value uint16
    err := binary.Read(d.buffer, binary.BigEndian, &value)
    return value, err
}

func (d *PacketDecoder) ReadInt32() (int32, error) {
    var value int32
    err := binary.Read(d.buffer, binary.BigEndian, &value)
    return value, err
}

func (d *PacketDecoder) ReadUint32() (uint32, error) {
    var value uint32
    err := binary.Read(d.buffer, binary.BigEndian, &value)
    return value, err
}

func (d *PacketDecoder) ReadInt64() (int64, error) {
    var value int64
    err := binary.Read(d.buffer, binary.BigEndian, &value)
    return value, err
}

func (d *PacketDecoder) ReadUint64() (uint64, error) {
    var value uint64
    err := binary.Read(d.buffer, binary.BigEndian, &value)
    return value, err
}

func (d *PacketDecoder) ReadFloat32() (float32, error) {
    var value float32
    err := binary.Read(d.buffer, binary.BigEndian, &value)
    return value, err
}

func (d *PacketDecoder) ReadString() (string, error) {
    length, err := d.ReadUint16()
    if err != nil {
        return "", err
    }
    
    data, err := d.ReadBytes(int(length))
    if err != nil {
        return "", err
    }
    
    return string(data), nil
}

func (d *PacketDecoder) ReadVarInt() (int32, error) {
    var value uint32
    for i := 0; i < 5; i++ {
        b, err := d.ReadByte()
        if err != nil {
            return 0, err
        }
        
        value |= uint32(b&0x7F) << (7 * i)
        
        if b&0x80 == 0 {
            break
        }
    }
    
    return int32(value), nil
}

func (d *PacketDecoder) ReadVarLong() (int64, error) {
    var value uint64
    for i := 0; i < 10; i++ {
        b, err := d.ReadByte()
        if err != nil {
            return 0, err
        }
        
        value |= uint64(b&0x7F) << (7 * i)
        
        if b&0x80 == 0 {
            break
        }
    }
    
    return int64(value), nil
}

func DecodeLoginPacket(data []byte) (*LoginPacket, error) {
    if len(data) < 1 {
        return nil, fmt.Errorf("packet too short")
    }
    
    decoder := NewPacketDecoder(data)
    
    packetID, err := decoder.ReadByte()
    if err != nil {
        return nil, err
    }
    
    if packetID != IDLogin {
        return nil, fmt.Errorf("not a login packet")
    }
    
    packet := &LoginPacket{}
    
    // Protocol version
    packet.ProtocolVersion, err = decoder.ReadInt32()
    if err != nil {
        return nil, err
    }
    
    // Connection request data (length-prefixed)
    requestDataLen, err := decoder.ReadUint32()
    if err != nil {
        return nil, err
    }
    
    packet.ConnectionRequestData, err = decoder.ReadBytes(int(requestDataLen))
    if err != nil {
        return nil, err
    }
    
    // Client network version
    packet.ClientNetworkVersion, err = decoder.ReadInt32()
    if err != nil {
        return nil, err
    }
    
    return packet, nil
}

func DecodePlayStatusPacket(data []byte) (*PlayStatusPacket, error) {
    if len(data) < 5 {
        return nil, fmt.Errorf("packet too short")
    }
    
    decoder := NewPacketDecoder(data)
    
    packetID, err := decoder.ReadByte()
    if err != nil {
        return nil, err
    }
    
    if packetID != IDPlayStatus {
        return nil, fmt.Errorf("not a play status packet")
    }
    
    packet := &PlayStatusPacket{}
    packet.Status, err = decoder.ReadInt32()
    if err != nil {
        return nil, err
    }
    
    return packet, nil
}