package raknet

import (
    "garuda/minecraft"
    "garuda/pkg/utils"
)

type ReliableFrame struct {
    SequenceNumber uint32
    Packets        [][]byte
}

type Session struct {
    addr          *net.UDPAddr
    server        *Server
    logger        *utils.Logger
    connected     bool
    mtuSize       int
    lastActivity  time.Time
    
    // Minecraft packet handling
    packetChan    chan []byte
    closeChan     chan bool
}

// Add to Session struct in session.go
func (s *Session) StartMinecraftHandler() {
    s.packetChan = make(chan []byte, 100)
    s.closeChan = make(chan bool, 1)
    
    go s.handleMinecraftPackets()
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
            // Pass to Minecraft server
            // This would be connected to the main server instance
            s.logger.Debug("Received login packet from %s", s.addr.String())
        }
    case minecraft.ID_TEXT:
        textPacket := &minecraft.TextPacket{}
        if err := textPacket.Decode(data); err == nil {
            s.logger.Info("[CHAT] %s: %s", s.addr.String(), textPacket.Message)
        }
    default:
        s.logger.Debug("Received Minecraft packet 0x%02x from %s", packetID, s.addr.String())
    }
}

// Update SendMinecraftPacket method
func (s *Session) SendMinecraftPacket(packetData []byte) {
    if !s.connected {
        return
    }
    
    // In real implementation, this would split large packets and add reliability
    s.server.sendResponse(packetData, s.addr)
    s.logger.Debug("Sent Minecraft packet 0x%02x to %s", packetData[0], s.addr.String())
}