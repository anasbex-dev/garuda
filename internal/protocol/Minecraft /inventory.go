package minecraft

const (
    IDInventoryTransaction = 0x1e
    IDMobEquipment = 0x1f
    IDMobArmorEquipment = 0x20
    IDContainerOpen = 0x2e
    IDContainerClose = 0x2f
    IDContainerSetData = 0x31
)

type InventoryTransactionPacket struct {
    TransactionType uint32
    Actions         []InventoryAction
    TransactionData TransactionData
}

type InventoryAction struct {
    SourceType  uint32
    WindowID    uint32
    SourceFlags uint32
    InventorySlot uint32
    OldItem     *ItemStack
    NewItem     *ItemStack
}

type TransactionData struct {
    RequestID        int32
    RequestChanged   []int32
    TransactionType  int32
}

type MobEquipmentPacket struct {
    RuntimeID uint64
    Item      *ItemStack
    Slot      byte
    SelectedSlot byte
    WindowID  byte
}

type MobArmorEquipmentPacket struct {
    RuntimeID uint64
    Helmet    *ItemStack
    Chestplate *ItemStack
    Leggings  *ItemStack
    Boots     *ItemStack
}

type ContainerOpenPacket struct {
    WindowID byte
    Type     byte
    Position BlockPos
    RuntimeID uint64
}

type ContainerClosePacket struct {
    WindowID byte
}

type ContainerSetDataPacket struct {
    WindowID byte
    Property int32
    Value    int32
}

func DecodeInventoryTransactionPacket(data []byte) (*InventoryTransactionPacket, error) {
    decoder := NewPacketDecoder(data)
    
    packetID, err := decoder.ReadByte()
    if err != nil {
        return nil, err
    }
    
    if packetID != IDInventoryTransaction {
        return nil, fmt.Errorf("not an inventory transaction packet")
    }
    
    packet := &InventoryTransactionPacket{}
    
    packet.TransactionType, err = decoder.ReadUint32()
    if err != nil {
        return nil, err
    }
    
    // Read actions
    actionCount, err := decoder.ReadUint32()
    if err != nil {
        return nil, err
    }
    
    packet.Actions = make([]InventoryAction, actionCount)
    for i := uint32(0); i < actionCount; i++ {
        action := &InventoryAction{}
        
        action.SourceType, err = decoder.ReadUint32()
        if err != nil {
            return nil, err
        }
        
        action.WindowID, err = decoder.ReadUint32()
        if err != nil {
            return nil, err
        }
        
        action.SourceFlags, err = decoder.ReadUint32()
        if err != nil {
            return nil, err
        }
        
        action.InventorySlot, err = decoder.ReadUint32()
        if err != nil {
            return nil, err
        }
        
        // Old item
        oldItem, err := decodeItemStack(decoder)
        if err != nil {
            return nil, err
        }
        action.OldItem = oldItem
        
        // New item
        newItem, err := decodeItemStack(decoder)
        if err != nil {
            return nil, err
        }
        action.NewItem = newItem
        
        packet.Actions[i] = *action
    }
    
    // Transaction data based on type
    switch packet.TransactionType {
    case 0: // Normal
        // No additional data
    case 1: // Mismatch
        packet.TransactionData.RequestID, err = decoder.ReadInt32()
        if err != nil {
            return nil, err
        }
        
        changedCount, err := decoder.ReadUint32()
        if err != nil {
            return nil, err
        }
        
        packet.TransactionData.RequestChanged = make([]int32, changedCount)
        for i := uint32(0); i < changedCount; i++ {
            packet.TransactionData.RequestChanged[i], err = decoder.ReadInt32()
            if err != nil {
                return nil, err
            }
        }
    }
    
    return packet, nil
}

func decodeItemStack(decoder *PacketDecoder) (*ItemStack, error) {
    id, err := decoder.ReadUint16()
    if err != nil {
        return nil, err
    }
    
    if id == 0 {
        return &ItemStack{ID: 0, Count: 0}, nil
    }
    
    count, err := decoder.ReadByte()
    if err != nil {
        return nil, err
    }
    
    damage, err := decoder.ReadUint16()
    if err != nil {
        return nil, err
    }
    
    // Skip NBT for now
    nbtLen, err := decoder.ReadUint16()
    if err != nil {
        return nil, err
    }
    
    if nbtLen > 0 {
        _, err = decoder.ReadBytes(int(nbtLen))
        if err != nil {
            return nil, err
        }
    }
    
    return &ItemStack{
        ID:     id,
        Count:  count,
        Damage: damage,
    }, nil
}

func EncodeMobEquipmentPacket(packet *MobEquipmentPacket) ([]byte, error) {
    encoder := NewPacketEncoder()
    
    if err := encoder.WriteByte(IDMobEquipment); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteUint64(packet.RuntimeID); err != nil {
        return nil, err
    }
    
    if err := encodeItemStack(encoder, packet.Item); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteByte(packet.Slot); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteByte(packet.SelectedSlot); err != nil {
        return nil, err
    }
    
    if err := encoder.WriteByte(packet.WindowID); err != nil {
        return nil, err
    }
    
    return encoder.Bytes(), nil
}

func encodeItemStack(encoder *PacketEncoder, item *ItemStack) error {
    if item == nil || item.ID == 0 {
        return encoder.WriteUint16(0)
    }
    
    if err := encoder.WriteUint16(item.ID); err != nil {
        return err
    }
    
    if err := encoder.WriteByte(item.Count); err != nil {
        return err
    }
    
    if err := encoder.WriteUint16(item.Damage); err != nil {
        return err
    }
    
    // Empty NBT
    if err := encoder.WriteUint16(0); err != nil {
        return err
    }
    
    return nil
}