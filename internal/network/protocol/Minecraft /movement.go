package minecraft

const (
    IDMovePlayer = 0x13
    IDPlayerAction = 0x24
    IDBlockEvent = 0x1a
    IDUpdateBlock = 0x15
)

type MovePlayerPacket struct {
    RuntimeID uint64
    Position Vector3
    Rotation Vector2
    Mode byte
    OnGround bool
    Tick uint64
}

type PlayerActionPacket struct {
    RuntimeID uint64
    Action int32
    Position BlockPos
    Face int32
}

type UpdateBlockPacket struct {
    Position BlockPos
    BlockID uint32
    Flags uint32
    Layer uint32
}

type BlockEventPacket struct {
    Position BlockPos
    EventType int32
    EventData int32
}

func DecodeMovePlayerPacket(data []byte) (*MovePlayerPacket, error) {
    if len(data) < 1 {
        return nil, fmt.Errorf("packet too short")
    }
    
    decoder := NewPacketDecoder(data)
    
    packetID, err := decoder.ReadByte()
    if err != nil {
        return nil, err
    }
    
    if packetID != IDMovePlayer {
        return nil, fmt.Errorf("not a move player packet")
    }
    
    packet := &MovePlayerPacket{}
    
    packet.RuntimeID, err = decoder.ReadUint64()
    if err != nil {
        return nil, err
    }
    
    // Position
    packet.Position.X, err = decoder.ReadFloat32()
    if err != nil {
        return nil, err
    }
    packet.Position.Y, err = decoder.ReadFloat32()
    if err != nil {
        return nil, err
    }
    packet.Position.Z, err = decoder.ReadFloat32()
    if err != nil {
        return nil, err
    }
    
    // Rotation (Pitch, Yaw)
    packet.Rotation.X, err = decoder.ReadFloat32()
    if err != nil {
        return nil, err
    }
    packet.Rotation.Y, err = decoder.ReadFloat32()
    if err != nil {
        return nil, err
    }
    
    packet.Mode, err = decoder.ReadByte()
    if err != nil {
        return nil, err
    }
    
    packet.OnGround, err = decoder.ReadBool()
    if err != nil {
        return nil, err
    }
    
    packet.Tick, err = decoder.ReadUint64()
    if err != nil {
        return nil, err
    }
    
    return packet, nil
}

func DecodePlayerActionPacket(data []byte) (*PlayerActionPacket, error) {
    if len(data) < 1 {
        return nil, fmt.Errorf("packet too short")
    }
    
    decoder := NewPacketDecoder(data)
    
    packetID, err := decoder.ReadByte()
    if err != nil {
        return nil, err
    }
    
    if packetID != IDPlayerAction {
        return nil, fmt.Errorf("not a player action packet")
    }
    
    packet := &PlayerActionPacket{}
    
    packet.RuntimeID, err = decoder.ReadUint64()
    if err != nil {
        return nil, err
    }
    
    packet.Action, err = decoder.ReadVarInt()
    if err != nil {
        return nil, err
    }
    
    // Position
    packet.Position.X, err = decoder.ReadVarInt()
    if err != nil {
        return nil, err
    }
    packet.Position.Y, err = decoder.ReadVarInt()
    if err != nil {
        return nil, err
    }
    packet.Position.Z, err = decoder.ReadVarInt()
    if err != nil {
        return nil, err
    }
    
    packet.Face, err = decoder.ReadVarInt()
    if err != nil {
        return nil, err
    }
    
    return packet, nil
}

func EncodeMovePlayerPacket(packet *MovePlayerPacket) ([]byte, error) {
    encoder := NewPacketEncoder()
    
    if err := encoder.WriteByte(IDMovePlayer); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteUint64(packet.RuntimeID); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteFloat32(packet.Position.X); err != nil {
        return nil, err
    }
    if err := encoder.WriteFloat32(packet.Position.Y); err != nil {
        return nil, err
    }
    if err := encoder.WriteFloat32(packet.Position.Z); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteFloat32(packet.Rotation.X); err != nil {
        return nil, err
    }
    if err := encoder.WriteFloat32(packet.Rotation.Y); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteByte(packet.Mode); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteBool(packet.OnGround); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteUint64(packet.Tick); err != nil {
        return nil, err
    }
    
    return encoder.Bytes(), nil
}

func EncodeUpdateBlockPacket(packet *UpdateBlockPacket) ([]byte, error) {
    encoder := NewPacketEncoder()
    
    if err := encoder.WriteByte(IDUpdateBlock); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteVarInt(packet.Position.X); err != nil {
        return nil, err
    }
    if err := encoder.WriteVarInt(packet.Position.Y); err != nil {
        return nil, err
    }
    if err := encoder.WriteVarInt(packet.Position.Z); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteUint32(packet.BlockID); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteUint32(packet.Flags); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteUint32(packet.Layer); err != nil {
        return nil, err
    }
    
    return encoder.Bytes(), nil
}