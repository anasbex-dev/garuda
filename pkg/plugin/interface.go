package plugin

import (
    "github.com/anabex-dev/garuda/internal/protocol/minecraft"
    "github.com/anabex-dev/garuda/internal/world"
)

// Plugin interface yang harus diimplement oleh semua plugin
type Plugin interface {
    // Metadata plugin
    GetName() string
    GetVersion() string
    GetAuthor() string
    
    // Lifecycle methods
    OnEnable(manager *PluginManager)
    OnDisable()
    
    // Optional event handlers
    EventHandler
}

// EventHandler interface untuk handle berbagai events
type EventHandler interface {
    OnPlayerJoin(player *world.Player)
    OnPlayerQuit(player *world.Player)
    OnPlayerChat(player *world.Player, message string) bool
    OnPlayerMove(player *world.Player, from, to minecraft.Vector3) bool
    OnBlockBreak(player *world.Player, pos minecraft.BlockPos, block world.Block) bool
    OnBlockPlace(player *world.Player, pos minecraft.BlockPos, block world.Block) bool
    OnPlayerCommand(player *world.Player, command string, args []string) bool
}

// BasePlugin provides default implementations untuk memudahkan development
type BasePlugin struct{}

func (p *BasePlugin) OnPlayerJoin(player *world.Player) {}
func (p *BasePlugin) OnPlayerQuit(player *world.Player) {}
func (p *BasePlugin) OnPlayerChat(player *world.Player, message string) bool { return true }
func (p *BasePlugin) OnPlayerMove(player *world.Player, from, to minecraft.Vector3) bool { return true }
func (p *BasePlugin) OnBlockBreak(player *world.Player, pos minecraft.BlockPos, block world.Block) bool { return true }
func (p *BasePlugin) OnBlockPlace(player *world.Player, pos minecraft.BlockPos, block world.Block) bool { return true }
func (p *BasePlugin) OnPlayerCommand(player *world.Player, command string, args []string) bool { return true }