package raknet

type RakNetServer struct {
    conn         *net.UDPConn
    sessions     map[string]*Session
    sessionsMutex sync.RWMutex
    running      bool
    address      string
    port         int
    world        *world.World
    
    ackTicker    *time.Ticker
    timeoutTicker *time.Ticker
}

func NewServer(address string, port int) *RakNetServer {
    // Create default world
    defaultWorld := world.NewWorld("garuda-world", 12345)
    
    return &RakNetServer{
        address:  address,
        port:     port,
        sessions: make(map[string]*Session),
        world:    defaultWorld,
        running:  false,
    }
}

func (s *RakNetServer) createSession(addr *net.UDPAddr) *Session {
    session := &Session{
        address:      addr,
        mtuSize:      minMTUSize,
        state:        StateUnconnected,
        lastActivity: time.Now(),
        server:       s,
        world:        s.world, // Share the world instance
    }
    
    session.reliableManager = NewReliableManager(session)
    return session
}