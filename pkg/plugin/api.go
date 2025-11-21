package plugin

import (
    "github.com/anabex-dev/garuda/internal/protocol/minecraft"
    "github.com/anabex-dev/garuda/internal/world"
)

// GarudaAPI provides methods untuk plugins berinteraksi dengan server
type GarudaAPI struct {
    manager *PluginManager
}

func NewGarudaAPI(manager *PluginManager) *GarudaAPI {
    return &GarudaAPI{
        manager: manager,
    }
}

// Player management
func (api *GarudaAPI) BroadcastMessage(message string) {
    api.manager.server.BroadcastMessage(message)
}

func (api *GarudaAPI) GetPlayer(name string) *world.Player {
    return api.manager.server.GetPlayer(name)
}

func (api *GarudaAPI) GetOnlinePlayers() []*world.Player {
    return api.manager.server.GetOnlinePlayers()
}

func (api *GarudaAPI) SendMessageToPlayer(player *world.Player, message string) {
    // TODO: Implement player message sending
}

func (api *GarudaAPI) KickPlayer(player *world.Player, reason string) {
    // TODO: Implement player kicking
}

// World management
func (api *GarudaAPI) GetWorld() *world.World {
    return api.manager.server.GetWorld()
}

func (api *GarudaAPI) SetBlock(pos minecraft.BlockPos, block world.Block) {
    world := api.GetWorld()
    chunkX := pos.X >> 4
    chunkZ := pos.Z >> 4
    localX := pos.X & 0xF
    localZ := pos.Z & 0xF
    
    chunk := world.GetChunk(chunkX, chunkZ)
    chunk.SetBlock(localX, pos.Y, localZ, block)
    
    // TODO: Broadcast block update
}

func (api *GarudaAPI) GetBlock(pos minecraft.BlockPos) world.Block {
    world := api.GetWorld()
    chunkX := pos.X >> 4
    chunkZ := pos.Z >> 4
    localX := pos.X & 0xF
    localZ := pos.Z & 0xF
    
    chunk := world.GetChunk(chunkX, chunkZ)
    return chunk.GetBlock(localX, pos.Y, localZ)
}

// Command execution
func (api *GarudaAPI) ExecuteCommand(command string) bool {
    return api.manager.server.ExecuteCommand(command)
}

// Plugin management
func (api *GarudaAPI) GetPlugin(name string) Plugin {
    return api.manager.GetPlugin(name)
}

func (api *GarudaAPI) IsPluginEnabled(name string) bool {
    return api.manager.IsEnabled(name)
}

// Utility methods
func (api *GarudaAPI) Logger() *log.Logger {
    return log.Default()
}