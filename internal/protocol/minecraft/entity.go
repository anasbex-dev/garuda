package minecraft

const (
    IDAddActor = 0x0d
    IDRemoveActor = 0x0e
    IDAddItemActor = 0x0f
    IDSetActorMotion = 0x1f
    IDUpdateAttributes = 0x1d
)

type AddActorPacket struct {
    RuntimeID uint64
    Type      string
    Position  Vector3
    Motion    Vector3
    Rotation  Vector2
    Attributes []EntityAttribute
    EntityData map[string]interface{}
}

type RemoveActorPacket struct {
    RuntimeID uint64
}

type AddItemActorPacket struct {
    RuntimeID uint64
    Item      *ItemStack
    Position  Vector3
    Motion    Vector3
    EntityData map[string]interface{}
}

type SetActorMotionPacket struct {
    RuntimeID uint64
    Motion    Vector3
}

type EntityAttribute struct {
    Name      string
    MinValue  float32
    MaxValue  float32
    Value     float32
    Default   float32
}

func EncodeAddActorPacket(packet *AddActorPacket) ([]byte, error) {
    encoder := NewPacketEncoder()
    
    if err := encoder.WriteByte(IDAddActor); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteVarLong(int64(packet.RuntimeID)); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteString(packet.Type); err != nil {
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
    
    if err := encoder.WriteFloat32(packet.Motion.X); err != nil {
        return nil, err
    }
    if err := encoder.WriteFloat32(packet.Motion.Y); err != nil {
        return nil, err
    }
    if err := encoder.WriteFloat32(packet.Motion.Z); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteFloat32(packet.Rotation.X); err != nil {
        return nil, err
    }
    if err := encoder.WriteFloat32(packet.Rotation.Y); err != nil {
        return nil, err
    }
    
    // Attributes
    if err := encoder.WriteUint32(uint32(len(packet.Attributes))); err != nil {
        return nil, err
    }
    
    for _, attr := range packet.Attributes {
        if err := encoder.WriteString(attr.Name); err != nil {
            return nil, err
        }
        if err := encoder.WriteFloat32(attr.MinValue); err != nil {
            return nil, err
        }
        if err := encoder.WriteFloat32(attr.MaxValue); err != nil {
            return nil, err
        }
        if err := encoder.WriteFloat32(attr.Value); err != nil {
            return nil, err
        }
        if err := encoder.WriteFloat32(attr.Default); err != nil {
            return nil, err
        }
    }
    
    // Entity data (simplified - just write empty for now)
    if err := encoder.WriteUint32(0); err != nil {
        return nil, err
    }
    
    return encoder.Bytes(), nil
}

func EncodeRemoveActorPacket(packet *RemoveActorPacket) ([]byte, error) {
    encoder := NewPacketEncoder()
    
    if err := encoder.WriteByte(IDRemoveActor); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteVarLong(int64(packet.RuntimeID)); err != nil {
        return nil, err
    }
    
    return encoder.Bytes(), nil
}