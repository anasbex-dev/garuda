package raknet

import (
    "encoding/binary"
    "garuda/internal/protocol/minecraft"
    "garuda/pkg/utils"
    "net"
    "time"
)

type Session struct {
    addr          *net.UDPAddr
    server        *Server
    logger        *utils.Logger
    connected     bool
    mtuSize       int
    lastActivity  time.Time
    packetChan    chan []byte
    closeChan     chan bool
    player        *server.Player
    compression   bool
    encryption    bool
    crypto      *utils.CryptoHandler
    encrypted   bool
}

func NewSession(addr *net.UDPAddr, server *Server, logger *utils.Logger) *Session {
    crypto, _ := utils.NewCryptoHandler()
    session := &Session{
        addr:         addr,
        server:       server,
        logger:       logger,
        mtuSize:      1400,
        lastActivity: time.Now(),
        packetChan:   make(chan []byte, 100),
        closeChan:    make(chan bool, 1),
        crypto:      crypto,
        encrypted:   false,
    }
    
    go session.handleMinecraftPackets()
    return session
}

func (s *Session) GetPublicKey() []byte {
    if s.crypto != nil {
        return s.crypto.GetPublicKey()
    }
    return nil
}

// Method untuk enable encryption
func (s *Session) EnableEncryption(sharedSecret []byte) error {
    if s.crypto == nil {
        return fmt.Errorf("crypto handler not initialized")
    }
    
    err := s.crypto.SetupEncryption(sharedSecret)
    if err != nil {
        return err
    }
    
    s.encrypted = true
    s.logger.Debug("Encryption enabled for session %s", s.addr.String())
    return nil
}

func (s *Session) processMinecraftPacket(data []byte) {
    if len(data) < 1 {
        return
    }
    
    packetID := data[0]
    
    switch packetID {
    case minecraft.ID_LOGIN:
        loginPacket := &minecraft.LoginPacket{}
        if err := loginPacket.Decode(data); err == nil {
            s.logger.Info("Processing login from %s", s.addr.String())
            // Pass ke server untuk handle login
            s.server.HandleLogin(s, loginPacket)
        }
    case minecraft.ID_TEXT:
        textPacket := &minecraft.TextPacket{}
        if err := textPacket.Decode(data); err == nil {
            s.logger.Debug("Chat message from %s: %s", s.addr.String(), textPacket.Message)
            s.server.HandleChatMessage(s, textPacket)
        }
    case minecraft.ID_MOVE_PLAYER:
        movePacket := &minecraft.MovePlayerPacket{}
        if err := movePacket.Decode(data); err == nil {
            s.server.HandleMovePlayer(s, movePacket)
        }
    case minecraft.ID_PLAYER_ACTION:
        actionPacket := &minecraft.PlayerActionPacket{}
        if err := actionPacket.Decode(data); err == nil {
            s.server.HandlePlayerAttack(s, actionPacket)
        }
    case minecraft.ID_INVENTORY_SLOT:
        slotPacket := &minecraft.InventorySlotPacket{}
        if err := slotPacket.Decode(data); err == nil {
            s.server.HandleInventoryClick(s, slotPacket)
        }
    default:
        s.logger.Debug("Received Minecraft packet 0x%02x from %s", packetID, s.addr.String())
    }
}

func (s *Session) handleMinecraftWrapper(data []byte) {
    s.packetChan <- data
}

func (s *Session) handleMinecraftPackets() {
    for {
        select {
        case packetData := <-s.packetChan:
            s.processMinecraftPacket(packetData)
        case <-s.closeChan:
            return
        }
    }
}

func (s *Session) processMinecraftPacket(data []byte) {
    if len(data) < 1 {
        return
    }
    
    packetID := data[0]
    
    switch packetID {
    case minecraft.ID_LOGIN:
        loginPacket := &minecraft.LoginPacket{}
        if err := loginPacket.Decode(data); err == nil {
            s.logger.Info("Processing login from %s", s.addr.String())
        }
    case minecraft.ID_TEXT:
        textPacket := &minecraft.TextPacket{}
        if err := textPacket.Decode(data); err == nil {
            s.logger.Debug("Chat message from %s: %s", s.addr.String(), textPacket.Message)
        }
    case minecraft.ID_MOVE_PLAYER:
        movePacket := &minecraft.MovePlayerPacket{}
        if err := movePacket.Decode(data); err == nil {
            s.logger.Debug("Move player from %s", s.addr.String())
        }
    default:
        s.logger.Debug("Received Minecraft packet 0x%02x from %s", packetID, s.addr.String())
    }
}

func (s *Session) handleOpenConnectionRequest2(data []byte) {
    stream := utils.NewBinaryStream(nil)
    stream.WriteByte(ID_OPEN_CONNECTION_REPLY_2)
    stream.WriteBytes(Magic[:])
    stream.WriteUint64(uint64(s.server.serverGUID))
    
    stream.WriteByte(byte(4))
    stream.WriteBytes(s.addr.IP.To4())
    stream.WriteUint16(uint16(s.addr.Port))
    stream.WriteUint16(uint16(s.mtuSize))
    stream.WriteByte(0)
    
    s.server.sendResponse(stream.Bytes(), s.addr)
    s.logger.Debug("Sent OpenConnectionReply2 to %s", s.addr.String())
}

func (s *Session) handleConnectionRequest(data []byte) {
    stream := utils.NewBinaryStream(nil)
    stream.WriteByte(ID_CONNECTION_REQUEST_ACCEPTED)
    
    stream.WriteByte(byte(4))
    stream.WriteBytes(s.addr.IP.To4())
    stream.WriteUint16(uint16(s.addr.Port))
    
    for i := 0; i < 10; i++ {
        stream.WriteByte(byte(4))
        stream.WriteBytes([]byte{127, 0, 0, 1})
        stream.WriteUint16(19132)
    }
    
    currentTime := uint64(time.Now().UnixNano() / int64(time.Millisecond))
    stream.WriteUint64(currentTime)
    stream.WriteUint64(currentTime)
    
    s.server.sendResponse(stream.Bytes(), s.addr)
    s.logger.Debug("Sent ConnectionRequestAccepted to %s", s.addr.String())
}

func (s *Session) handleNewIncomingConnection(data []byte) {
    s.connected = true
    s.logger.Info("Client %s fully connected", s.addr.String())
}

func (s *Session) SendMinecraftPacket(packetData []byte) {
    if !s.connected {
        return
    }
    
    s.server.sendResponse(packetData, s.addr)
}

func (s *Session) Close() {
    if s.connected {
        stream := utils.NewBinaryStream(nil)
        stream.WriteByte(ID_DISCONNECTION_NOTICE)
        s.server.sendResponse(stream.Bytes(), s.addr)
    }
    s.connected = false
    s.closeChan <- true
}

func (s *Session) IsConnected() bool {
    return s.connected
}

func (s *Session) GetAddress() string {
    return s.addr.String()
}