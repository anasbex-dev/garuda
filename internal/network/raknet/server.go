package raknet

// Tambah field di RakNetServer
type RakNetServer struct {
    conn         *net.UDPConn
    sessions     map[string]*Session
    sessionsMutex sync.RWMutex
    running      bool
    address      string
    port         int
    
    // Timer untuk reliable manager tasks
    ackTicker    *time.Ticker
    timeoutTicker *time.Ticker
}

// Update handlePacket function
func (s *RakNetServer) handlePacket(addr *net.UDPAddr, data []byte) {
    if len(data) < 1 {
        return
    }

    packetID := data[0]
    
    session := s.getOrCreateSession(addr)
    session.UpdateActivity()

    switch {
    case packetID >= 0x80 && packetID <= 0x8d: // ACK ranges
        session.reliableManager.HandleAck(data)
    case packetID >= 0x8e && packetID <= 0x9d: // NACK ranges
        session.reliableManager.HandleAck(data)
    case packetID == 0x00: // Connected ping with timestamp
        // Handle game packets
        packets := session.reliableManager.ProcessReceivedDatagram(data)
        for _, packet := range packets {
            session.HandleGamePacket(packet.Data)
        }
    default:
        // Handle other packet types as before
        switch packetID {
        case 0x01: // Unconnected Ping
            s.handleUnconnectedPing(addr, data)
        case 0x05: // Open Connection Request 1
            s.handleOpenConnectionRequest1(addr, data)
        case 0x07: // Open Connection Request 2
            s.handleOpenConnectionRequest2(addr, data)
        case 0x08: // Connection Request
            s.handleConnectionRequest(session, data)
        case 0x09: // New Incoming Connection
            s.handleNewIncomingConnection(session, data)
        case 0x13: // Disconnect Notification
            s.handleDisconnect(session)
        case 0x15: // Connected Ping
            s.handleConnectedPing(session, data)
        case 0x1c: // Connected Pong
            // Just update activity
        default:
            log.Printf("Unknown packet ID: 0x%02x from %s", packetID, addr)
        }
    }
}

// Update Start function untuk start timers
func (s *RakNetServer) Start() error {
    addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.address, s.port))
    if err != nil {
        return err
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        return err
    }

    s.conn = conn
    s.running = true

    // Start timers for reliable messaging
    s.ackTicker = time.NewTicker(100 * time.Millisecond)
    s.timeoutTicker = time.NewTicker(1 * time.Second)

    log.Printf("RakNet server listening on %s:%d", s.address, s.port)
    
    go s.readLoop()
    go s.cleanupLoop()
    go s.reliableManagerLoop()

    return nil
}

func (s *RakNetServer) reliableManagerLoop() {
    for s.running {
        select {
        case <-s.ackTicker.C:
            s.sendAllAcks()
        case <-s.timeoutTicker.C:
            s.checkAllTimeouts()
        }
    }
}

func (s *RakNetServer) sendAllAcks() {
    s.sessionsMutex.RLock()
    defer s.sessionsMutex.RUnlock()
    
    for _, session := range s.sessions {
        session.reliableManager.SendAcks()
    }
}

func (s *RakNetServer) checkAllTimeouts() {
    s.sessionsMutex.RLock()
    defer s.sessionsMutex.RUnlock()
    
    for _, session := range s.sessions {
        session.reliableManager.CheckTimeouts()
    }
}

// Update Stop function
func (s *RakNetServer) Stop() {
    s.running = false
    if s.ackTicker != nil {
        s.ackTicker.Stop()
    }
    if s.timeoutTicker != nil {
        s.timeoutTicker.Stop()
    }
    if s.conn != nil {
        s.conn.Close()
    }
}