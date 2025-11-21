package minecraft

import (
    "bytes"
    "encoding/binary"
)

type PacketEncoder struct {
    buffer *bytes.Buffer
}

func NewPacketEncoder() *PacketEncoder {
    return &PacketEncoder{
        buffer: bytes.NewBuffer(nil),
    }
}

func (e *PacketEncoder) WriteByte(value byte) error {
    return e.buffer.WriteByte(value)
}

func (e *PacketEncoder) WriteBytes(data []byte) error {
    _, err := e.buffer.Write(data)
    return err
}

func (e *PacketEncoder) WriteBool(value bool) error {
    if value {
        return e.WriteByte(1)
    }
    return e.WriteByte(0)
}

func (e *PacketEncoder) WriteInt16(value int16) error {
    return binary.Write(e.buffer, binary.BigEndian, value)
}

func (e *PacketEncoder) WriteUint16(value uint16) error {
    return binary.Write(e.buffer, binary.BigEndian, value)
}

func (e *Encoder) WriteInt32(value int32) error {
    return binary.Write(e.buffer, binary.BigEndian, value)
}

func (e *Encoder) WriteUint32(value uint32) error {
    return binary.Write(e.buffer, binary.BigEndian, value)
}

func (e *Encoder) WriteInt64(value int64) error {
    return binary.Write(e.buffer, binary.BigEndian, value)
}

func (e *Encoder) WriteUint64(value uint64) error {
    return binary.Write(e.buffer, binary.BigEndian, value)
}

func (e *Encoder) WriteFloat32(value float32) error {
    return binary.Write(e.buffer, binary.BigEndian, value)
}

func (e *Encoder) WriteString(value string) error {
    data := []byte(value)
    if err := e.WriteUint16(uint16(len(data))); err != nil {
        return err
    }
    return e.WriteBytes(data)
}

func (e *Encoder) WriteVarInt(value int32) error {
    v := uint32(value)
    for {
        b := byte(v & 0x7F)
        v >>= 7
        if v != 0 {
            b |= 0x80
        }
        if err := e.WriteByte(b); err != nil {
            return err
        }
        if v == 0 {
            break
        }
    }
    return nil
}

func (e *Encoder) Bytes() []byte {
    return e.buffer.Bytes()
}

func EncodePlayStatusPacket(status int32) ([]byte, error) {
    encoder := NewPacketEncoder()
    
    // Packet ID
    if err := encoder.WriteByte(IDPlayStatus); err != nil {
        return nil, err
    }
    
    // Status
    if err := encoder.WriteInt32(status); err != nil {
        return nil, err
    }
    
    return encoder.Bytes(), nil
}

func EncodeDisconnectPacket(hideScreen bool, message string) ([]byte, error) {
    encoder := NewPacketEncoder()
    
    // Packet ID
    if err := encoder.WriteByte(IDDisconnect); err != nil {
        return nil, err
    }
    
    // Hide disconnection screen
    if err := encoder.WriteBool(hideScreen); err != nil {
        return nil, err
    }
    
    // Message
    if err := encoder.WriteString(message); err != nil {
        return nil, err
    }
    
    return encoder.Bytes(), nil
}

func EncodeStartGamePacket(packet *StartGamePacket) ([]byte, error) {
    encoder := NewPacketEncoder()
    
    // Packet ID
    if err := encoder.WriteByte(IDStartGame); err != nil {
        return nil, err
    }
    
    // Entity ID
    if err := encoder.WriteInt64(packet.EntityID); err != nil {
        return nil, err
    }
    
    // Runtime Entity ID
    if err := encoder.WriteUint64(packet.RuntimeEntityID); err != nil {
        return nil, err
    }
    
    // Player Game Type
    if err := encoder.WriteInt32(packet.PlayerGameType); err != nil {
        return nil, err
    }
    
    // Player Position
    if err := encoder.WriteFloat32(packet.PlayerPosition.X); err != nil {
        return nil, err
    }
    if err := encoder.WriteFloat32(packet.PlayerPosition.Y); err != nil {
        return nil, err
    }
    if err := encoder.WriteFloat32(packet.PlayerPosition.Z); err != nil {
        return nil, err
    }
    
    // Rotation
    if err := encoder.WriteFloat32(packet.Rotation.X); err != nil {
        return nil, err
    }
    if err := encoder.WriteFloat32(packet.Rotation.Y); err != nil {
        return nil, err
    }
    
    // Seed
    if err := encoder.WriteInt32(packet.Seed); err != nil {
        return nil, err
    }
    
    // Biome Type
    if err := encoder.WriteInt16(packet.BiomeType); err != nil {
        return nil, err
    }
    
    // Biome Name
    if err := encoder.WriteString(packet.BiomeName); err != nil {
        return nil, err
    }
    
    // Dimension
    if err := encoder.WriteInt32(packet.Dimension); err != nil {
        return nil, err
    }
    
    // Generator
    if err := encoder.WriteInt32(packet.Generator); err != nil {
        return nil, err
    }
    
    // World Game Mode
    if err := encoder.WriteInt32(packet.WorldGameMode); err != nil {
        return nil, err
    }
    
    // Difficulty
    if err := encoder.WriteInt32(packet.Difficulty); err != nil {
        return nil, err
    }
    
    // Spawn Position
    if err := encoder.WriteInt32(packet.SpawnPosition.X); err != nil {
        return nil, err
    }
    if err := encoder.WriteInt32(packet.SpawnPosition.Y); err != nil {
        return nil, err
    }
    if err := encoder.WriteInt32(packet.SpawnPosition.Z); err != nil {
        return nil, err
    }
    
    // Achievements Disabled
    if err := encoder.WriteBool(packet.AchievementsDisabled); err != nil {
        return nil, err
    }
    
    // Time
    if err := encoder.WriteInt32(packet.Time); err != nil {
        return nil, err
    }
    
    // Edu Mode
    if err := encoder.WriteBool(packet.EduMode); err != nil {
        return nil, err
    }
    
    // Rain Level
    if err := encoder.WriteFloat32(packet.RainLevel); err != nil {
        return nil, err
    }
    
    // Lightning Level
    if err := encoder.WriteFloat32(packet.LightningLevel); err != nil {
        return nil, err
    }
    
    // Commands Enabled
    if err := encoder.WriteBool(packet.CommandsEnabled); err != nil {
        return nil, err
    }
    
    // Texture Packs Required
    if err := encoder.WriteBool(packet.TexturePacksRequired); err != nil {
        return nil, err
    }
    
    // Game Rules
    if err := encoder.WriteUint32(uint32(len(packet.GameRules))); err != nil {
        return nil, err
    }
    for _, rule := range packet.GameRules {
        if err := encoder.WriteString(rule.Name); err != nil {
            return nil, err
        }
        // Note: Game rule value encoding would need type-specific handling
    }
    
    // Experiments
    if err := encoder.WriteUint32(uint32(len(packet.Experiments))); err != nil {
        return nil, err
    }
    for _, exp := range packet.Experiments {
        if err := encoder.WriteString(exp.Name); err != nil {
            return nil, err
        }
        if err := encoder.WriteBool(exp.Enabled); err != nil {
            return nil, err
        }
    }
    
    // Experiments Previously Used
    if err := encoder.WriteBool(packet.ExperimentsPreviouslyUsed); err != nil {
        return nil, err
    }
    
    // Bonus Chest Enabled
    if err := encoder.WriteBool(packet.BonusChestEnabled); err != nil {
        return nil, err
    }
    
    // Start With Map Enabled
    if err := encoder.WriteBool(packet.StartWithMapEnabled); err != nil {
        return nil, err
    }
    
    // Player Permissions
    if err := encoder.WriteInt32(packet.PlayerPermissions); err != nil {
        return nil, err
    }
    
    // Chunk Radius
    if err := encoder.WriteInt32(packet.ChunkRadius); err != nil {
        return nil, err
    }
    
    // Server Chunk Tick Range
    if err := encoder.WriteInt32(packet.ServerChunkTickRange); err != nil {
        return nil, err
    }
    
    // Broadcast To LAN
    if err := encoder.WriteBool(packet.BroadcastToLAN); err != nil {
        return nil, err
    }
    
    // XBL Broadcast Mode
    if err := encoder.WriteInt32(packet.XBLBroadcastMode); err != nil {
        return nil, err
    }
    
    // Platform Broadcast Mode
    if err := encoder.WriteInt32(packet.PlatformBroadcastMode); err != nil {
        return nil, err
    }
    
    // XBL Broadcast Intent
    if err := encoder.WriteBool(packet.XBLBroadcastIntent); err != nil {
        return nil, err
    }
    
    // Platform Broadcast Intent
    if err := encoder.WriteBool(packet.PlatformBroadcastIntent); err != nil {
        return nil, err
    }
    
    // Commands Enabled On First Join
    if err := encoder.WriteBool(packet.CommandsEnabledOnFirstJoin); err != nil {
        return nil, err
    }
    
    return encoder.Bytes(), nil
}