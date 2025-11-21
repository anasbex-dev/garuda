package raknet

import (
    "sync"
    "time"
)

type Session struct {
    address        *net.UDPAddr
    mtuSize        int
    guid           int64
    state          SessionState
    lastActivity   time.Time
    server         *RakNetServer
    reliableManager *ReliableManager
    mutex          sync.RWMutex
    
    // Game session data
    playerName     string
    playerID       int64
}

func (s *RakNetServer) createSession(addr *net.UDPAddr) *Session {
    session := &Session{
        address:      addr,
        mtuSize:      minMTUSize,
        state:        StateUnconnected,
        lastActivity: time.Now(),
        server:       s,
    }
    
    session.reliableManager = NewReliableManager(session)
    return session
}

func (s *Session) UpdateActivity() {
    s.mutex.Lock()
    s.lastActivity = time.Now()
    s.mutex.Unlock()
}

func (s *Session) GetState() SessionState {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.state
}

func (s *Session) SetState(state SessionState) {
    s.mutex.Lock()
    s.state = state
    s.mutex.Unlock()
}

func (s *Session) SendGamePacket(packetData []byte) error {
    packet := &EncapsulatedPacket{
        Reliability: ReliabilityReliable,
        Data:        packetData,
    }
    
    return s.reliableManager.SendPacket(packet, ReliabilityReliable)
}

func (s *Session) HandleGamePacket(packetData []byte) {
    // Parse Minecraft packet
    if len(packetData) < 1 {
        return
    }
    
    packetID := packetData[0]
    
    switch packetID {
    case 0x01: // Login packet
        s.handleLoginPacket(packetData)
    case 0x02: // Play status packet
        s.handlePlayStatusPacket(packetData)
    case 0x03: // Server to client handshake
        s.handleServerHandshakePacket(packetData)
    case 0x04: // Client to server handshake
        s.handleClientHandshakePacket(packetData)
    default:
        log.Printf("Unknown game packet ID: 0x%02x", packetID)
    }
}

func (s *Session) handleLoginPacket(data []byte) {
    if len(data) < 3 {
        return
    }
    
    // Parse login packet (simplified)
    protocolVersion := binary.BigEndian.Uint32(data[1:5])
    
    log.Printf("Login request: protocol=%d from %s", protocolVersion, s.address)
    
    // Send play status packet (login success)
    response := make([]byte, 3)
    response[0] = 0x02 // Play status packet
    binary.BigEndian.PutUint32(response[1:5], 0) // Success
    
    s.SendGamePacket(response)
    
    s.SetState(StateConnected)
}

func (s *Session) handlePlayStatusPacket(data []byte) {
    // Handle play status response
}

func (s *Session) handleServerHandshakePacket(data []byte) {
    // Handle server handshake
}

func (s *Session) handleClientHandshakePacket(data []byte) {
    // Handle client handshake
}