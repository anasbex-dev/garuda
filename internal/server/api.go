package server

import (
    "log"

    "github.com/anabex-dev/garuda/internal/world"
    "github.com/anabex-dev/garuda/pkg/plugin"
)

// GarudaServer implements plugin.ServerAPI
type GarudaServer struct {
    world   *world.World
    players map[string]*world.Player
}

func NewGarudaServer(world *world.World) *GarudaServer {
    return &GarudaServer{
        world:   world,
        players: make(map[string]*world.Player),
    }
}

func (s *GarudaServer) BroadcastMessage(message string) {
    log.Printf("[CHAT] %s", message)
    // TODO: Implement broadcast to all players
}

func (s *GarudaServer) GetPlayer(name string) *world.Player {
    // This is a simplified implementation
    // In real implementation, you'd search through connected players
    for _, player := range s.players {
        if player.Username == name {
            return player
        }
    }
    return nil
}

func (s *GarudaServer) GetOnlinePlayers() []*world.Player {
    players := make([]*world.Player, 0, len(s.players))
    for _, player := range s.players {
        players = append(players, player)
    }
    return players
}

func (s *GarudaServer) ExecuteCommand(command string) bool {
    log.Printf("Executing command: %s", command)
    // TODO: Implement command execution
    return true
}

func (s *GarudaServer) GetWorld() *world.World {
    return s.world
}

// Helper method untuk menambah player (dipanggil dari session)
func (s *GarudaServer) AddPlayer(player *world.Player) {
    s.players[player.Username] = player
}

// Helper method untuk remove player (dipanggil dari session)
func (s *GarudaServer) RemovePlayer(player *world.Player) {
    delete(s.players, player.Username)
}