package raknet

import (
    "fmt"
    "garuda/pkg/utils"
    "net"
    "sync"
    "time"
)

type Server struct {
    address     string
    listener    *net.UDPConn
    logger      *utils.Logger
    sessions    map[string]*Session
    sessionMutex sync.RWMutex
    running     bool
    serverGUID  int64
}

func NewServer(address string, logger *utils.Logger) *Server {
    return &Server{
        address:    address,
        logger:     logger,
        sessions:   make(map[string]*Session),
        serverGUID: generateGUID(),
        running:    false,
    }
}

func generateGUID() int64 {
    return time.Now().UnixNano()
}

func (s *Server) Start() error {
    udpAddr, err := net.ResolveUDPAddr("udp", s.address)
    if err != nil {
        return fmt.Errorf("failed to resolve address: %v", err)
    }

    s.listener, err = net.ListenUDP("udp", udpAddr)
    if err != nil {
        return fmt.Errorf("failed to listen: %v", err)
    }

    s.running = true
    s.logger.Info("RakNet server started on %s", s.address)
    s.logger.Info("Server GUID: %d", s.serverGUID)

    go s.handlePackets()

    return nil
}

func (s *Server) handlePackets() {
    buffer := make([]byte, 1500) // Standard MTU

    for s.running {
        n, addr, err := s.listener.ReadFromUDP(buffer)
        if err != nil {
            if s.running {
                s.logger.Error("Error reading from UDP: %v", err)
            }
            continue
        }

        data := make([]byte, n)
        copy(data, buffer[:n])

        go s.handlePacket(data, addr)
    }
}

func (s *Server) handlePacket(data []byte, addr *net.UDPAddr) {
    packet, err := ReadPacket(data)
    if err != nil {
        s.logger.Debug("Failed to read packet from %s: %v", addr.String(), err)
        return
    }

    clientAddr := addr.String()

    switch packet.ID {
    case ID_UNCONNECTED_PING:
        s.handleUnconnectedPing(packet.Data, addr)
    case ID_OPEN_CONNECTION_REQUEST_1:
        s.handleOpenConnectionRequest1(packet.Data, addr)
    case ID_OPEN_CONNECTION_REQUEST_2:
        s.handleOpenConnectionRequest2(packet.Data, addr)
    case ID_CONNECTION_REQUEST:
        s.handleConnectionRequest(packet.Data, clientAddr)
    default:
        // Handle connected packets through session
        s.sessionMutex.RLock()
        session, exists := s.sessions[clientAddr]
        s.sessionMutex.RUnlock()

        if exists {
            session.handlePacket(packet)
        } else {
            s.logger.Debug("Unknown packet ID 0x%02x from %s", packet.ID, clientAddr)
        }
    }
}

func (s *Server) handleUnconnectedPing(data []byte, addr *net.UDPAddr) {
    ping, err := DecodeUnconnectedPing(data)
    if err != nil {
        s.logger.Debug("Failed to decode unconnected ping: %v", err)
        return
    }

    if !VerifyMagic(ping.Magic) {
        s.logger.Debug("Invalid magic from %s", addr.String())
        return
    }

    // USE PROTOCOL MANAGER FOR MOTD
    motd := s.protocolManager.GetMOTD(s.config.Server.MOTD, len(s.players), s.config.Server.MaxPlayers)
    pong := &UnconnectedPong{
        PingID:    ping.PingID,
        ServerGUID: s.serverGUID,
        Magic:     Magic,
        MOTD:      motd,
    }

    response := EncodeUnconnectedPong(pong)
    s.sendResponse(response, addr)
    s.logger.Debug("Sent pong to %s", addr.String())
}

func (s *Server) handleOpenConnectionRequest1(data []byte, addr *net.UDPAddr) {
    request, err := DecodeOpenConnectionRequest1(data)
    if err != nil {
        s.logger.Debug("Failed to decode open connection request 1: %v", err)
        return
    }

    if !VerifyMagic(request.Magic) {
        s.logger.Debug("Invalid magic in OpenConnectionRequest1 from %s", addr.String())
        return
    }

    // Check protocol version
    if request.Protocol != 10 { // Bedrock protocol version
        s.logger.Debug("Incompatible protocol version %d from %s", request.Protocol, addr.String())
        // Send incompatible protocol packet
        return
    }

    reply := &OpenConnectionReply1{
        Magic:      Magic,
        ServerGUID: s.serverGUID,
        UseSecurity: false,
        MTUSize:    int16(request.MTUSize),
    }

    response := EncodeOpenConnectionReply1(reply)
    s.sendResponse(response, addr)
    s.logger.Debug("Sent OpenConnectionReply1 to %s", addr.String())
}

func (s *Server) handleOpenConnectionRequest2(data []byte, addr *net.UDPAddr) {
    // For now, just accept the connection
    session := NewSession(addr, s, s.logger)
    
    s.sessionMutex.Lock()
    s.sessions[addr.String()] = session
    s.sessionMutex.Unlock()

    session.handleOpenConnectionRequest2(data)
    s.logger.Info("New connection from %s", addr.String())
}

func (s *Server) handleConnectionRequest(data []byte, clientAddr string) {
    s.sessionMutex.RLock()
    session, exists := s.sessions[clientAddr]
    s.sessionMutex.RUnlock()

    if exists {
        session.handleConnectionRequest(data)
    }
}

func (s *Server) sendResponse(data []byte, addr *net.UDPAddr) {
    _, err := s.listener.WriteToUDP(data, addr)
    if err != nil {
        s.logger.Error("Failed to send response to %s: %v", addr.String(), err)
    }
}

func (s *Server) Close() {
    s.running = false
    if s.listener != nil {
        s.listener.Close()
    }
    
    // Close all sessions
    s.sessionMutex.Lock()
    for _, session := range s.sessions {
        session.Close()
    }
    s.sessions = make(map[string]*Session)
    s.sessionMutex.Unlock()
    
    s.logger.Info("RakNet server stopped")
}