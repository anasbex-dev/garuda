package minecraft

const (
    // Connection packets
    ID_LOGIN = 0x01
    ID_PLAY_STATUS = 0x02
    ID_DISCONNECT = 0x05
    ID_RESOURCE_PACKS_INFO = 0x06
    
    // Gameplay packets  
    ID_START_GAME = 0x0b
    ID_SET_TIME = 0x0a
    ID_TEXT = 0x09
    ID_MOVE_PLAYER = 0x13
    
    // Entity packets
    ID_ADD_PLAYER = 0x02
    ID_ADD_ENTITY = 0x0f
    ID_ADD_ITEM_ENTITY = 0x15
    ID_REMOVE_ENTITY = 0x14
    
    // Inventory packets
    ID_MOB_EQUIPMENT = 0x1f
    ID_INVENTORY_CONTENT = 0x32
    ID_INVENTORY_SLOT = 0x33
    
    // Block packets
    ID_UPDATE_BLOCK = 0x15
    ID_PLAYER_ACTION = 0x24
    ID_BLOCK_EVENT = 0x1a
    
    // World packets
    ID_LEVEL_CHUNK = 0x3a
)

type Packet interface {
    ID() byte
    Encode() ([]byte, error)
    Decode([]byte) error
}

type ItemStack struct {
    ID    uint32
    Count byte
    Data  uint16
}

type BlockCoordinates struct {
    X int32
    Y int32
    Z int32
}

// Basic packets tetap di sini
type LoginPacket struct {
    ProtocolVersion int32
    ChainData       map[string]interface{}
    ClientData      map[string]interface{}
}

func (p *LoginPacket) ID() byte { return ID_LOGIN }

func (p *LoginPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    encoder.WriteVarInt(p.ProtocolVersion)
    
    chainDataJSON := `{"chain":[]}`
    encoder.WriteString(chainDataJSON)
    
    clientDataJSON := `{"ClientData":{}}`
    encoder.WriteString(clientDataJSON)
    
    return encoder.Bytes(), nil
}

func (p *LoginPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    p.ProtocolVersion = decoder.ReadVarInt()
    
    chainDataStr := decoder.ReadString()
    clientDataStr := decoder.ReadString()
    
    p.ChainData = make(map[string]interface{})
    p.ClientData = make(map[string]interface{})
    
    return nil
}

type PlayStatusPacket struct {
    Status int32
}

func (p *PlayStatusPacket) ID() byte { return ID_PLAY_STATUS }

func (p *PlayStatusPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    encoder.WriteVarInt(p.Status)
    return encoder.Bytes(), nil
}

func (p *PlayStatusPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    p.Status = decoder.ReadVarInt()
    return nil
}

type DisconnectPacket struct {
    HideDisconnectionScreen bool
    Message string
}

func (p *DisconnectPacket) ID() byte { return ID_DISCONNECT }

func (p *DisconnectPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    encoder.WriteBool(p.HideDisconnectionScreen)
    if !p.HideDisconnectionScreen {
        encoder.WriteString(p.Message)
    }
    return encoder.Bytes(), nil
}

func (p *DisconnectPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    p.HideDisconnectionScreen = decoder.ReadBool()
    if !p.HideDisconnectionScreen {
        p.Message = decoder.ReadString()
    }
    return nil
}

type TextPacket struct {
    TextType byte
    Message  string
}

func (p *TextPacket) ID() byte { return ID_TEXT }

func (p *TextPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    encoder.WriteByte(p.TextType)
    
    switch p.TextType {
    case 1, 2:
        encoder.WriteString("")
        encoder.WriteString(p.Message)
        encoder.WriteString("")
        encoder.WriteString("")
    case 5:
        encoder.WriteString(p.Message)
    default:
        encoder.WriteBool(false)
        encoder.WriteString(p.Message)
        encoder.WriteString("")
        encoder.WriteString("")
        encoder.WriteString("")
    }
    
    return encoder.Bytes(), nil
}

func (p *TextPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    p.TextType = decoder.stream.ReadByte()
    
    switch p.TextType {
    case 1, 2:
        _ = decoder.ReadString()
        p.Message = decoder.ReadString()
    case 5:
        p.Message = decoder.ReadString()
    default:
        _ = decoder.ReadBool()
        p.Message = decoder.ReadString()
    }
    
    return nil
}

type SetTimePacket struct {
    Time int32
}

func (p *SetTimePacket) ID() byte { return ID_SET_TIME }

func (p *SetTimePacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    encoder.WriteVarInt(p.Time)
    return encoder.Bytes(), nil
}

func (p *SetTimePacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    p.Time = decoder.ReadVarInt()
    return nil
}

type StartGamePacket struct {
    EntityID       int64
    RuntimeID      uint64
    PlayerGamemode int32
    Position       [3]float32
    WorldName      string
}

func (p *StartGamePacket) ID() byte { return ID_START_GAME }

func (p *StartGamePacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    
    encoder.WriteLong(p.EntityID)
    encoder.WriteUnsignedLong(p.RuntimeID)
    encoder.WriteVarInt(p.PlayerGamemode)
    
    // Position
    encoder.WriteFloat32(p.Position[0])
    encoder.WriteFloat32(p.Position[1])
    encoder.WriteFloat32(p.Position[2])
    
    // Default rotation
    encoder.WriteFloat32(0) // Pitch
    encoder.WriteFloat32(0) // Yaw
    
    // Seed & dimension
    encoder.WriteVarInt(0) // Seed
    encoder.WriteVarInt(0) // Dimension
    
    // World name
    encoder.WriteString(p.WorldName)
    
    return encoder.Bytes(), nil
}

func (p *StartGamePacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    
    p.EntityID = decoder.ReadLong()
    p.RuntimeID = decoder.ReadUnsignedLong()
    p.PlayerGamemode = decoder.ReadVarInt()
    
    p.Position[0] = decoder.ReadFloat32()
    p.Position[1] = decoder.ReadFloat32()
    p.Position[2] = decoder.ReadFloat32()
    
    _ = decoder.ReadFloat32() // Pitch
    _ = decoder.ReadFloat32() // Yaw
    
    _ = decoder.ReadVarInt() // Seed
    _ = decoder.ReadVarInt() // Dimension
    
    p.WorldName = decoder.ReadString()
    
    return nil
}

type MovePlayerPacket struct {
    EntityID uint64
    Position [3]float32
    Pitch    float32
    Yaw      float32
    HeadYaw  float32
    Mode     byte
    OnGround bool
}

func (p *MovePlayerPacket) ID() byte { return ID_MOVE_PLAYER }

func (p *MovePlayerPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    
    encoder.WriteUnsignedLong(p.EntityID)
    encoder.WriteFloat32(p.Position[0])
    encoder.WriteFloat32(p.Position[1])
    encoder.WriteFloat32(p.Position[2])
    encoder.WriteFloat32(p.Pitch)
    encoder.WriteFloat32(p.Yaw)
    encoder.WriteFloat32(p.HeadYaw)
    encoder.WriteByte(p.Mode)
    encoder.WriteBool(p.OnGround)
    
    return encoder.Bytes(), nil
}

func (p *MovePlayerPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    
    p.EntityID = decoder.ReadUnsignedLong()
    p.Position[0] = decoder.ReadFloat32()
    p.Position[1] = decoder.ReadFloat32()
    p.Position[2] = decoder.ReadFloat32()
    p.Pitch = decoder.ReadFloat32()
    p.Yaw = decoder.ReadFloat32()
    p.HeadYaw = decoder.ReadFloat32()
    p.Mode = decoder.stream.ReadByte()
    p.OnGround = decoder.ReadBool()
    
    return nil
}

type AddEntityPacket struct {
    EntityID   int64
    RuntimeID  uint64
    EntityType string
    Position   [3]float32
    Velocity   [3]float32
    Pitch      float32
    Yaw        float32
}

func (p *AddEntityPacket) ID() byte { return ID_ADD_ENTITY }

func (p *AddEntityPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    
    encoder.WriteVarLong(p.EntityID)
    encoder.WriteUnsignedLong(p.RuntimeID)
    encoder.WriteString(p.EntityType)
    
    // Position
    encoder.WriteFloat32(p.Position[0])
    encoder.WriteFloat32(p.Position[1])
    encoder.WriteFloat32(p.Position[2])
    
    // Velocity
    encoder.WriteFloat32(p.Velocity[0])
    encoder.WriteFloat32(p.Velocity[1])
    encoder.WriteFloat32(p.Velocity[2])
    
    // Rotation
    encoder.WriteFloat32(p.Pitch)
    encoder.WriteFloat32(p.Yaw)
    
    // Attributes (empty for now)
    encoder.WriteVarInt(0)
    
    // Metadata
    encoder.WriteVarInt(0)
    
    return encoder.Bytes(), nil
}

func (p *AddEntityPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    
    p.EntityID = decoder.ReadVarLong()
    p.RuntimeID = decoder.ReadUnsignedLong()
    p.EntityType = decoder.ReadString()
    
    p.Position[0] = decoder.ReadFloat32()
    p.Position[1] = decoder.ReadFloat32()
    p.Position[2] = decoder.ReadFloat32()
    
    p.Velocity[0] = decoder.ReadFloat32()
    p.Velocity[1] = decoder.ReadFloat32()
    p.Velocity[2] = decoder.ReadFloat32()
    
    p.Pitch = decoder.ReadFloat32()
    p.Yaw = decoder.ReadFloat32()
    
    _ = decoder.ReadVarInt() // Skip attributes
    _ = decoder.ReadVarInt() // Skip metadata
    
    return nil
}

type AddItemEntityPacket struct {
    EntityID  int64
    RuntimeID uint64
    Item      ItemStack
    Position  [3]float32
    Velocity  [3]float32
}

func (p *AddItemEntityPacket) ID() byte { return ID_ADD_ITEM_ENTITY }

func (p *AddItemEntityPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    
    encoder.WriteVarLong(p.EntityID)
    encoder.WriteUnsignedLong(p.RuntimeID)
    
    // Item stack
    encoder.WriteVarInt(int32(p.Item.ID))
    encoder.WriteByte(p.Item.Count)
    encoder.WriteShort(p.Item.Data)
    
    // Position
    encoder.WriteFloat32(p.Position[0])
    encoder.WriteFloat32(p.Position[1])
    encoder.WriteFloat32(p.Position[2])
    
    // Velocity
    encoder.WriteFloat32(p.Velocity[0])
    encoder.WriteFloat32(p.Velocity[1])
    encoder.WriteFloat32(p.Velocity[2())
    
    // Metadata
    encoder.WriteVarInt(0)
    
    return encoder.Bytes(), nil
}

func (p *AddItemEntityPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    
    p.EntityID = decoder.ReadVarLong()
    p.RuntimeID = decoder.ReadUnsignedLong()
    
    p.Item.ID = uint32(decoder.ReadVarInt())
    p.Item.Count = decoder.stream.ReadByte()
    p.Item.Data = decoder.ReadShort()
    
    p.Position[0] = decoder.ReadFloat32()
    p.Position[1] = decoder.ReadFloat32()
    p.Position[2] = decoder.ReadFloat32()
    
    p.Velocity[0] = decoder.ReadFloat32()
    p.Velocity[1] = decoder.ReadFloat32()
    p.Velocity[2] = decoder.ReadFloat32()
    
    _ = decoder.ReadVarInt() // Skip metadata
    
    return nil
}

type RemoveEntityPacket struct {
    EntityID int64
}

func (p *RemoveEntityPacket) ID() byte { return ID_REMOVE_ENTITY }

func (p *RemoveEntityPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    encoder.WriteVarLong(p.EntityID)
    return encoder.Bytes(), nil
}

func (p *RemoveEntityPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    p.EntityID = decoder.ReadVarLong()
    return nil
}

type InventoryContentPacket struct {
    WindowID byte
    Items    []ItemStack
}

func (p *InventoryContentPacket) ID() byte { return ID_INVENTORY_CONTENT }

func (p *InventoryContentPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    
    encoder.WriteByte(p.WindowID)
    encoder.WriteVarInt(int32(len(p.Items)))
    
    for _, item := range p.Items {
        encoder.WriteVarInt(int32(item.ID))
        encoder.WriteByte(item.Count)
        encoder.WriteShort(item.Data)
    }
    
    return encoder.Bytes(), nil
}

func (p *InventoryContentPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    
    p.WindowID = decoder.stream.ReadByte()
    count := decoder.ReadVarInt()
    
    p.Items = make([]ItemStack, count)
    for i := 0; i < int(count); i++ {
        p.Items[i].ID = uint32(decoder.ReadVarInt())
        p.Items[i].Count = decoder.stream.ReadByte()
        p.Items[i].Data = decoder.ReadShort()
    }
    
    return nil
}

type InventorySlotPacket struct {
    WindowID byte
    Slot     uint32
    Item     ItemStack
}

func (p *InventorySlotPacket) ID() byte { return ID_INVENTORY_SLOT }

func (p *InventorySlotPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    
    encoder.WriteByte(p.WindowID)
    encoder.WriteVarInt(int32(p.Slot))
    encoder.WriteVarInt(int32(p.Item.ID))
    encoder.WriteByte(p.Item.Count)
    encoder.WriteShort(p.Item.Data)
    
    return encoder.Bytes(), nil
}

func (p *InventorySlotPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    
    p.WindowID = decoder.stream.ReadByte()
    p.Slot = uint32(decoder.ReadVarInt())
    p.Item.ID = uint32(decoder.ReadVarInt())
    p.Item.Count = decoder.stream.ReadByte()
    p.Item.Data = decoder.ReadShort()
    
    return nil
}

type PlayerActionPacket struct {
    EntityID  int64
    Action    int32
    Position  BlockCoordinates
    Face      int32
}

func (p *PlayerActionPacket) ID() byte { return ID_PLAYER_ACTION }

func (p *PlayerActionPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    
    encoder.WriteVarLong(p.EntityID)
    encoder.WriteVarInt(p.Action)
    encoder.WriteBlockCoordinates(p.Position)
    encoder.WriteVarInt(p.Face)
    
    return encoder.Bytes(), nil
}

func (p *PlayerActionPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    
    p.EntityID = decoder.ReadVarLong()
    p.Action = decoder.ReadVarInt()
    p.Position = decoder.ReadBlockCoordinates()
    p.Face = decoder.ReadVarInt()
    
    return nil
}

type UpdateBlockPacket struct {
    Position BlockCoordinates
    BlockID  uint32
    Flags    uint32
}

func (p *UpdateBlockPacket) ID() byte { return ID_UPDATE_BLOCK }

func (p *UpdateBlockPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    
    encoder.WriteBlockCoordinates(p.Position)
    encoder.WriteVarInt(int32(p.BlockID))
    encoder.WriteVarInt(int32(p.Flags))
    
    return encoder.Bytes(), nil
}

func (p *UpdateBlockPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    
    p.Position = decoder.ReadBlockCoordinates()
    p.BlockID = uint32(decoder.ReadVarInt())
    p.Flags = uint32(decoder.ReadVarInt())
    
    return nil
}

// LevelChunkPacket untuk mengirim chunk data
type LevelChunkPacket struct {
    ChunkX        int32
    ChunkZ        int32
    SubChunkCount uint32
    Data          []byte
}

func (p *LevelChunkPacket) ID() byte { return ID_LEVEL_CHUNK }

func (p *LevelChunkPacket) Encode() ([]byte, error) {
    encoder := NewEncoder()
    
    encoder.WriteVarInt(p.ChunkX)
    encoder.WriteVarInt(p.ChunkZ)
    encoder.WriteVarInt(int32(p.SubChunkCount))
    encoder.WriteBool(false) // Cache enabled
    encoder.WriteUnsignedVarInt(uint64(len(p.Data)))
    encoder.WriteBytes(p.Data)
    
    return encoder.Bytes(), nil
}

func (p *LevelChunkPacket) Decode(data []byte) error {
    decoder := NewDecoder(data)
    
    p.ChunkX = decoder.ReadVarInt()
    p.ChunkZ = decoder.ReadVarInt()
    p.SubChunkCount = uint32(decoder.ReadVarInt())
    _ = decoder.ReadBool() // Cache enabled
    dataLen := decoder.ReadUnsignedVarInt()
    p.Data = decoder.ReadBytes(int(dataLen))
    
    return nil
}