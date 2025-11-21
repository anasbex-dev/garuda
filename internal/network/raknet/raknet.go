package raknet

import (
    "crypto/rand"
    "encoding/binary"
    "errors"
    "fmt"
    "log"
    "math"
    "net"
    "sync"
    "sync/atomic"
    "time"

    "github.com/anabex-dev/garuda/internal/world"
    "github.com/anabex-dev/garuda/pkg/plugin"
)

// RakNetServer represents the main RakNet server implementation
type RakNetServer struct {
    conn          *net.UDPConn
    sessions      map[string]*Session
    sessionsMutex sync.RWMutex
    running       atomic.Bool
    address       string
    port          int
    world         *world.World
    pluginManager *plugin.PluginManager
    serverAPI     ServerAPI
    
    // Network configuration
    mtuSize              int
    maxPlayers          int
    guid                uint64
    protocolVersion     byte
    compressionThreshold int
    
    // Timing and statistics
    startTime           time.Time
    statistics          *ServerStatistics
    statisticsMutex     sync.RWMutex
    
    // Channels for communication
    shutdownChan        chan struct{}
    sessionCloseChan    chan *Session
    broadcastChan       chan *BroadcastMessage
    
    // Connection management
    connectionQueue     chan *net.UDPAddr
    maxConnections      int
    currentConnections  atomic.Int32
    
    // Security
    securityCookie      uint32
    enableEncryption    bool
}

// ServerStatistics holds server-wide statistics
type ServerStatistics struct {
    TotalConnections    uint64
    CurrentConnections  uint32
    TotalPacketsSent    uint64
    TotalPacketsReceived uint64
    BytesSent           uint64
    BytesReceived       uint64
    PacketLoss          float32
    Uptime              time.Duration
}

// BroadcastMessage represents a message to broadcast to multiple sessions
type BroadcastMessage struct {
    Data    []byte
    Filter  func(*Session) bool
    Reliable bool
}

// ServerAPI defines the interface for server operations
type ServerAPI interface {
    BroadcastMessage(message string)
    GetPlayer(name string) *world.Player
    GetOnlinePlayers() []*world.Player
    ExecuteCommand(command string) bool
    GetWorld() *world.World
    GetServerStats() map[string]interface{}
}

// NewServer creates a new RakNet server instance
func NewServer(address string, port int) *RakNetServer {
    // Generate server GUID
    guid := generateGUID()
    
    // Generate security cookie
    var securityCookie uint32
    binary.Read(rand.Reader, binary.BigEndian, &securityCookie)
    
    server := &RakNetServer{
        address:     address,
        port:        port,
        sessions:    make(map[string]*Session),
        guid:        guid,
        protocolVersion: 11, // Minecraft Bedrock protocol version
        mtuSize:     maxMTUSize,
        maxPlayers:  20,
        compressionThreshold: 256,
        
        statistics: &ServerStatistics{},
        shutdownChan: make(chan struct{}),
        sessionCloseChan: make(chan *Session, 100),
        broadcastChan: make(chan *BroadcastMessage, 100),
        connectionQueue: make(chan *net.UDPAddr, 1000),
        maxConnections: 1000,
        
        securityCookie: securityCookie,
        enableEncryption: false,
    }
    
    server.running.Store(false)
    server.currentConnections.Store(0)
    
    log.Printf("RakNet server initialized with GUID: 0x%x", guid)
    return server
}

// Start starts the RakNet server
func (s *RakNetServer) Start() error {
    if s.running.Load() {
        return errors.New("server is already running")
    }
    
    addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", s.address, s.port))
    if err != nil {
        return fmt.Errorf("failed to resolve address: %v", err)
    }

    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        return fmt.Errorf("failed to listen on UDP: %v", err)
    }

    s.conn = conn
    s.running.Store(true)
    s.startTime = time.Now()
    
    log.Printf("RakNet server started on %s:%d", s.address, s.port)
    log.Printf("Server GUID: 0x%x", s.guid)
    log.Printf("Protocol Version: %d", s.protocolVersion)
    log.Printf("MTU Size: %d", s.mtuSize)
    
    // Start background workers
    go s.readLoop()
    go s.sessionManagerLoop()
    go s.broadcastLoop()
    go s.connectionManagerLoop()
    go s.statisticsLoop()
    go s.cleanupLoop()
    
    // Start reliability managers for timing
    go s.reliabilityManagerLoop()
    
    log.Printf("All server workers started successfully")
    return nil
}

// readLoop handles incoming UDP packets
func (s *RakNetServer) readLoop() {
    buffer := make([]byte, s.mtuSize)
    
    for s.running.Load() {
        n, addr, err := s.conn.ReadFromUDP(buffer)
        if err != nil {
            if s.running.Load() {
                log.Printf("Read error: %v", err)
            }
            continue
        }

        if n < 1 {
            continue
        }

        packet := make([]byte, n)
        copy(packet, buffer[:n])
        
        // Handle packet in separate goroutine for concurrency
        go s.handlePacket(addr, packet)
    }
}

// handlePacket processes an incoming packet
func (s *RakNetServer) handlePacket(addr *net.UDPAddr, data []byte) {
    if len(data) < 1 {
        return
    }

    packetID := data[0]
    
    // Update statistics
    s.updateStatistics(0, uint64(len(data)))
    
    // Get or create session
    session := s.getOrCreateSession(addr)
    if session == nil {
        return
    }
    
    session.UpdateActivity()

    switch {
    case packetID >= 0x80 && packetID <= 0x8d: // ACK ranges
        session.reliableManager.HandleAck(data)
        return
    case packetID >= 0x8e && packetID <= 0x9d: // NACK ranges
        session.reliableManager.HandleAck(data)
        return
    case packetID == 0x00: // Connected ping with timestamp
        // Handle game packets through reliable manager
        packets := session.reliableManager.ProcessReceivedDatagram(data)
        for _, packet := range packets {
            session.HandleGamePacket(packet.Data)
        }
        return
    }

    // Handle connection-level packets
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
        // Just update activity, no response needed
    default:
        log.Printf("Unknown packet ID: 0x%02x from %s", packetID, addr)
    }
}

// sessionManagerLoop manages session lifecycle
func (s *RakNetServer) sessionManagerLoop() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.shutdownChan:
            return
        case session := <-s.sessionCloseChan:
            s.removeSession(session)
        case <-ticker.C:
            s.cleanupInactiveSessions()
        }
    }
}

// broadcastLoop handles broadcasting messages to multiple sessions
func (s *RakNetServer) broadcastLoop() {
    for {
        select {
        case <-s.shutdownChan:
            return
        case msg := <-s.broadcastChan:
            s.broadcastToSessions(msg)
        }
    }
}

// connectionManagerLoop manages incoming connection queue
func (s *RakNetServer) connectionManagerLoop() {
    for {
        select {
        case <-s.shutdownChan:
            return
        case addr := <-s.connectionQueue:
            s.handleQueuedConnection(addr)
        }
    }
}

// statisticsLoop updates and logs server statistics
func (s *RakNetServer) statisticsLoop() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.shutdownChan:
            return
        case <-ticker.C:
            s.logStatistics()
        }
    }
}

// reliabilityManagerLoop handles timing for reliability systems
func (s *RakNetServer) reliabilityManagerLoop() {
    ackTicker := time.NewTicker(100 * time.Millisecond)
    timeoutTicker := time.NewTicker(1 * time.Second)
    defer ackTicker.Stop()
    defer timeoutTicker.Stop()
    
    for {
        select {
        case <-s.shutdownChan:
            return
        case <-ackTicker.C:
            s.sendAllAcks()
        case <-timeoutTicker.C:
            s.checkAllTimeouts()
        }
    }
}

// cleanupLoop performs periodic cleanup tasks
func (s *RakNetServer) cleanupLoop() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-s.shutdownChan:
            return
        case <-ticker.C:
            s.performCleanup()
        }
    }
}

// getOrCreateSession gets an existing session or creates a new one
func (s *RakNetServer) getOrCreateSession(addr *net.UDPAddr) *Session {
    key := addr.String()
    
    s.sessionsMutex.RLock()
    session, exists := s.sessions[key]
    s.sessionsMutex.RUnlock()
    
    if exists {
        return session
    }
    
    // Check connection limits
    if int(s.currentConnections.Load()) >= s.maxPlayers {
        log.Printf("Connection limit reached, rejecting %s", addr)
        return nil
    }
    
    // Create new session
    session = s.createSession(addr)
    
    s.sessionsMutex.Lock()
    s.sessions[key] = session
    s.sessionsMutex.Unlock()
    
    s.currentConnections.Add(1)
    s.statisticsMutex.Lock()
    s.statistics.TotalConnections++
    s.statistics.CurrentConnections = uint32(s.currentConnections.Load())
    s.statisticsMutex.Unlock()
    
    log.Printf("New session created for %s (Total: %d)", addr, s.currentConnections.Load())
    
    // Start session handler
    go session.Handle()
    
    return session
}

// createSession creates a new session instance
func (s *RakNetServer) createSession(addr *net.UDPAddr) *Session {
    session := &Session{
        address:       addr,
        conn:          s.conn,
        mtuSize:       s.mtuSize,
        guid:          int64(s.guid),
        state:         StateUnconnected,
        lastActivity:  time.Now(),
        server:        s,
        world:         s.world,
        packetChan:    make(chan *minecraft.Packet, 100),
        disconnectChan: make(chan bool, 1),
        closeChan:     make(chan bool, 1),
        latency:       0,
    }
    
    session.reliableManager = NewReliableManager(session)
    return session
}

// removeSession removes a session from the server
func (s *RakNetServer) removeSession(session *Session) {
    key := session.address.String()
    
    s.sessionsMutex.Lock()
    delete(s.sessions, key)
    s.sessionsMutex.Unlock()
    
    s.currentConnections.Add(-1)
    s.statisticsMutex.Lock()
    s.statistics.CurrentConnections = uint32(s.currentConnections.Load())
    s.statisticsMutex.Unlock()
    
    log.Printf("Session removed for %s (Remaining: %d)", session.address, s.currentConnections.Load())
}

// cleanupInactiveSessions removes sessions that have been inactive
func (s *RakNetServer) cleanupInactiveSessions() {
    timeout := 30 * time.Second
    now := time.Now()
    
    s.sessionsMutex.Lock()
    defer s.sessionsMutex.Unlock()
    
    for key, session := range s.sessions {
        if now.Sub(session.lastActivity) > timeout {
            log.Printf("Removing inactive session: %s", key)
            delete(s.sessions, key)
            s.currentConnections.Add(-1)
            session.Close()
        }
    }
    
    s.statisticsMutex.Lock()
    s.statistics.CurrentConnections = uint32(s.currentConnections.Load())
    s.statisticsMutex.Unlock()
}

// broadcastToSessions broadcasts a message to matching sessions
func (s *RakNetServer) broadcastToSessions(msg *BroadcastMessage) {
    s.sessionsMutex.RLock()
    defer s.sessionsMutex.RUnlock()
    
    for _, session := range s.sessions {
        if msg.Filter == nil || msg.Filter(session) {
            if msg.Reliable {
                session.reliableManager.SendPacket(&EncapsulatedPacket{
                    Reliability: ReliabilityReliable,
                    Data:        msg.Data,
                }, ReliabilityReliable)
            } else {
                session.SendGamePacket(msg.Data)
            }
        }
    }
}

// handleQueuedConnection processes a queued connection
func (s *RakNetServer) handleQueuedConnection(addr *net.UDPAddr) {
    // Implement connection queue processing logic
    // This can handle rate limiting, whitelist checks, etc.
}

// updateStatistics updates server statistics
func (s *RakNetServer) updateStatistics(packetsSent, bytesReceived uint64) {
    s.statisticsMutex.Lock()
    defer s.statisticsMutex.Unlock()
    
    s.statistics.TotalPacketsReceived++
    s.statistics.BytesReceived += bytesReceived
    s.statistics.Uptime = time.Since(s.startTime)
}

// logStatistics logs current server statistics
func (s *RakNetServer) logStatistics() {
    s.statisticsMutex.RLock()
    stats := s.statistics
    s.statisticsMutex.RUnlock()
    
    log.Printf("Server Stats - Connections: %d/%d, Packets: ↑%d ↓%d, Bytes: ↑%d ↓%d, Uptime: %v",
        stats.CurrentConnections, s.maxPlayers,
        stats.TotalPacketsSent, stats.TotalPacketsReceived,
        stats.BytesSent, stats.BytesReceived,
        stats.Uptime.Truncate(time.Second))
}

// performCleanup performs periodic cleanup tasks
func (s *RakNetServer) performCleanup() {
    // Force garbage collection
    // runtime.GC()
    
    // Log memory statistics
    // var m runtime.MemStats
    // runtime.ReadMemStats(&m)
    // log.Printf("Memory - Alloc: %.2fMB, Sys: %.2fMB, Goroutines: %d",
    //     float64(m.Alloc)/1024/1024, float64(m.Sys)/1024/1024, runtime.NumGoroutine())
}

// sendAllAcks sends ACKs for all sessions
func (s *RakNetServer) sendAllAcks() {
    s.sessionsMutex.RLock()
    defer s.sessionsMutex.RUnlock()
    
    for _, session := range s.sessions {
        session.reliableManager.SendAcks()
    }
}

// checkAllTimeouts checks timeouts for all sessions
func (s *RakNetServer) checkAllTimeouts() {
    s.sessionsMutex.RLock()
    defer s.sessionsMutex.RUnlock()
    
    for _, session := range s.sessions {
        session.reliableManager.CheckTimeouts()
    }
}

// SetWorld sets the world instance for the server
func (s *RakNetServer) SetWorld(world *world.World) {
    s.world = world
}

// SetPluginManager sets the plugin manager for the server
func (s *RakNetServer) SetPluginManager(manager *plugin.PluginManager) {
    s.pluginManager = manager
}

// SetServerAPI sets the server API implementation
func (s *RakNetServer) SetServerAPI(api ServerAPI) {
    s.serverAPI = api
}

// SetMaxPlayers sets the maximum number of players
func (s *RakNetServer) SetMaxPlayers(max int) {
    s.maxPlayers = max
}

// GetPlayerCount returns the current number of connected players
func (s *RakNetServer) GetPlayerCount() int {
    return int(s.currentConnections.Load())
}

// GetMaxPlayers returns the maximum number of players
func (s *RakNetServer) GetMaxPlayers() int {
    return s.maxPlayers
}

// IsRunning returns whether the server is running
func (s *RakNetServer) IsRunning() bool {
    return s.running.Load()
}

// GetGUID returns the server GUID
func (s *RakNetServer) GetGUID() uint64 {
    return s.guid
}

// Stop gracefully shuts down the server
func (s *RakNetServer) Stop() {
    if !s.running.Load() {
        return
    }
    
    log.Println("Shutting down RakNet server...")
    s.running.Store(false)
    
    // Close shutdown channel to stop all workers
    close(s.shutdownChan)
    
    // Close all sessions
    s.sessionsMutex.Lock()
    for _, session := range s.sessions {
        session.Close()
    }
    s.sessions = make(map[string]*Session)
    s.sessionsMutex.Unlock()
    
    // Close UDP connection
    if s.conn != nil {
        s.conn.Close()
    }
    
    log.Println("RakNet server stopped successfully")
}

// BroadcastMessage broadcasts a message to all sessions
func (s *RakNetServer) BroadcastMessage(data []byte, reliable bool) {
    msg := &BroadcastMessage{
        Data:    data,
        Filter:  nil, // Send to all
        Reliable: reliable,
    }
    s.broadcastChan <- msg
}

// BroadcastToFiltered broadcasts to sessions matching a filter
func (s *RakNetServer) BroadcastToFiltered(data []byte, filter func(*Session) bool, reliable bool) {
    msg := &BroadcastMessage{
        Data:    data,
        Filter:  filter,
        Reliable: reliable,
    }
    s.broadcastChan <- msg
}

// GetSession returns a session by address
func (s *RakNetServer) GetSession(addr *net.UDPAddr) *Session {
    key := addr.String()
    
    s.sessionsMutex.RLock()
    defer s.sessionsMutex.RUnlock()
    
    return s.sessions[key]
}

// GetSessions returns all active sessions
func (s *RakNetServer) GetSessions() []*Session {
    s.sessionsMutex.RLock()
    defer s.sessionsMutex.RUnlock()
    
    sessions := make([]*Session, 0, len(s.sessions))
    for _, session := range s.sessions {
        sessions = append(sessions, session)
    }
    return sessions
}

// GetStatistics returns server statistics
func (s *RakNetServer) GetStatistics() *ServerStatistics {
    s.statisticsMutex.RLock()
    defer s.statisticsMutex.RUnlock()
    
    // Return a copy to avoid race conditions
    stats := *s.statistics
    return &stats
}

// generateGUID generates a random server GUID
func generateGUID() uint64 {
    var guid uint64
    binary.Read(rand.Reader, binary.BigEndian, &guid)
    return guid
}

// Constants
const (
    maxMTUSize      = 1492
    minMTUSize      = 400
    defaultPort     = 19132
    protocolVersion = 11
)

// Implement the remaining packet handler methods...
func (s *RakNetServer) handleUnconnectedPing(addr *net.UDPAddr, data []byte) {
    // Implementation from previous code...
}

func (s *RakNetServer) handleOpenConnectionRequest1(addr *net.UDPAddr, data []byte) {
    // Implementation from previous code...
}

func (s *RakNetServer) handleOpenConnectionRequest2(addr *net.UDPAddr, data []byte) {
    // Implementation from previous code...
}

func (s *RakNetServer) handleConnectionRequest(session *Session, data []byte) {
    // Implementation from previous code...
}

func (s *RakNetServer) handleNewIncomingConnection(session *Session, data []byte) {
    // Implementation from previous code...
}

func (s *RakNetServer) handleDisconnect(session *Session) {
    // Implementation from previous code...
}

func (s *RakNetServer) handleConnectedPing(session *Session, data []byte) {
    // Implementation from previous code...
}