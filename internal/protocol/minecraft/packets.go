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