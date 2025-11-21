package minecraft

const (
    IDAnimate = 0x2c
    IDHurtArmor = 0x3d
)

type AnimatePacket struct {
    Action     byte
    RuntimeID  uint64
    RowdingTime float32
}

func DecodeAnimatePacket(data []byte) (*AnimatePacket, error) {
    if len(data) < 2 {
        return nil, fmt.Errorf("packet too short")
    }
    
    decoder := NewPacketDecoder(data)
    
    packetID, err := decoder.ReadByte()
    if err != nil {
        return nil, err
    }
    
    if packetID != IDAnimate {
        return nil, fmt.Errorf("not an animate packet")
    }
    
    packet := &AnimatePacket{}
    
    packet.Action, err = decoder.ReadByte()
    if err != nil {
        return nil, err
    }
    
    packet.RuntimeID, err = decoder.ReadVarLong()
    if err != nil {
        return nil, err
    }
    
    if packet.Action == 3 { // Critical hit
        packet.RowdingTime, err = decoder.ReadFloat32()
        if err != nil {
            return nil, err
        }
    }
    
    return packet, nil
}