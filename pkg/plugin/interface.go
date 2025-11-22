package plugin

import (
    "garuda/internal/protocol/minecraft"
    "garuda/internal/world"
)

type Plugin interface {
    GetName() string
    GetVersion() string
    GetAuthor() string
    GetDescription() string
    
    OnEnable(manager PluginManager)
    OnDisable()
}

type PluginManager interface {
    GetServer() Server
    GetPlugin(name string) Plugin
    RegisterEvent(eventType EventType, handler EventHandler)
    RegisterCommand(command string, handler CommandHandler)
    BroadcastMessage(message string)
    GetPlayer(name string) world.Player
    GetOnlinePlayers() []world.Player
}

type Server interface {
    GetName() string
    GetVersion() string
    GetMaxPlayers() int
    GetPlayerCount() int
    BroadcastPacket(packetData []byte)
    ExecuteCommand(command string) bool
}

type EventType int

const (
    EventPlayerJoin EventType = iota
    EventPlayerQuit
    EventPlayerChat
    EventPlayerMove
    EventPlayerBreakBlock
    EventPlayerPlaceBlock
    EventPlayerInteract
    EventEntitySpawn
    EventEntityDamage
    EventWorldLoad
    EventWorldSave
)

type EventHandler func(event Event)

type Event interface {
    GetType() EventType
    IsCancelled() bool
    SetCancelled(cancelled bool)
}

type PlayerJoinEvent struct {
    Player   world.Player
    cancelled bool
}

func (e *PlayerJoinEvent) GetType() EventType { return EventPlayerJoin }
func (e *PlayerJoinEvent) IsCancelled() bool { return e.cancelled }
func (e *PlayerJoinEvent) SetCancelled(cancelled bool) { e.cancelled = cancelled }

type PlayerQuitEvent struct {
    Player   world.Player
    Reason   string
}

func (e *PlayerQuitEvent) GetType() EventType { return EventPlayerQuit }
func (e *PlayerQuitEvent) IsCancelled() bool { return false }
func (e *PlayerQuitEvent) SetCancelled(cancelled bool) {}

type PlayerChatEvent struct {
    Player    world.Player
    Message   string
    cancelled bool
}

func (e *PlayerChatEvent) GetType() EventType { return EventPlayerChat }
func (e *PlayerChatEvent) IsCancelled() bool { return e.cancelled }
func (e *PlayerChatEvent) SetCancelled(cancelled bool) { e.cancelled = cancelled }

type PlayerMoveEvent struct {
    Player    world.Player
    FromPos   [3]float32
    ToPos     [3]float32
    cancelled bool
}

func (e *PlayerMoveEvent) GetType() EventType { return EventPlayerMove }
func (e *PlayerMoveEvent) IsCancelled() bool { return e.cancelled }
func (e *PlayerMoveEvent) SetCancelled(cancelled bool) { e.cancelled = cancelled }

type PlayerBreakBlockEvent struct {
    Player    world.Player
    BlockPos  [3]int
    BlockID   uint32
    cancelled bool
}

func (e *PlayerBreakBlockEvent) GetType() EventType { return EventPlayerBreakBlock }
func (e *PlayerBreakBlockEvent) IsCancelled() bool { return e.cancelled }
func (e *PlayerBreakBlockEvent) SetCancelled(cancelled bool) { e.cancelled = cancelled }

type PlayerPlaceBlockEvent struct {
    Player    world.Player
    BlockPos  [3]int
    BlockID   uint32
    cancelled bool
}

func (e *PlayerPlaceBlockEvent) GetType() EventType { return EventPlayerPlaceBlock }
func (e *PlayerPlaceBlockEvent) IsCancelled() bool { return e.cancelled }
func (e *PlayerPlaceBlockEvent) SetCancelled(cancelled bool) { e.cancelled = cancelled }

type CommandHandler func(sender CommandSender, command string, args []string) bool

type CommandSender interface {
    GetName() string
    IsPlayer() bool
    IsConsole() bool
    SendMessage(message string)
    HasPermission(permission string) bool
}

type PlayerCommandSender struct {
    Player world.Player
}

func (p *PlayerCommandSender) GetName() string { return p.Player.Username }
func (p *PlayerCommandSender) IsPlayer() bool { return true }
func (p *PlayerCommandSender) IsConsole() bool { return false }
func (p *PlayerCommandSender) SendMessage(message string) {
    // Will be implemented in server integration
}
func (p *PlayerCommandSender) HasPermission(permission string) bool {
    // Basic permission check - can be extended
    return true
}

type ConsoleCommandSender struct{}

func (c *ConsoleCommandSender) GetName() string { return "CONSOLE" }
func (c *ConsoleCommandSender) IsPlayer() bool { return false }
func (c *ConsoleCommandSender) IsConsole() bool { return true }
func (c *ConsoleCommandSender) SendMessage(message string) {
    // Will be implemented in server integration
}
func (c *ConsoleCommandSender) HasPermission(permission string) bool { return true }