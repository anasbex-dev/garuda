package plugin

import (
    "garuda/internal/server"
    "garuda/pkg/utils"
    "garuda/internal/world"
    "sync"
)

type PluginManagerImpl struct {
    server      *server.Server
    logger      *utils.Logger
    plugins     map[string]Plugin
    eventHandlers map[EventType][]EventHandler
    commands    map[string]CommandHandler
    scheduler   *SimpleScheduler
    permissions *SimplePermissionManager
    mutex       sync.RWMutex
}

func NewPluginManager(server *server.Server, logger *utils.Logger) *PluginManagerImpl {
    return &PluginManagerImpl{
        server:      server,
        logger:      logger,
        plugins:     make(map[string]Plugin),
        eventHandlers: make(map[EventType][]EventHandler),
        commands:    make(map[string]CommandHandler),
        scheduler:   NewSimpleScheduler(),
        permissions: NewSimplePermissionManager(),
    }
}

func (pm *PluginManagerImpl) LoadPlugin(plugin Plugin) bool {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    pluginName := plugin.GetName()
    
    if _, exists := pm.plugins[pluginName]; exists {
        pm.logger.Error("Plugin %s is already loaded", pluginName)
        return false
    }
    
    pm.plugins[pluginName] = plugin
    
    pm.logger.Info("Loading plugin: %s v%s by %s", 
        plugin.GetName(), plugin.GetVersion(), plugin.GetAuthor())
    
    plugin.OnEnable(pm)
    
    pm.logger.Info("Plugin %s enabled successfully", pluginName)
    return true
}

func (pm *PluginManagerImpl) UnloadPlugin(name string) bool {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    plugin, exists := pm.plugins[name]
    if !exists {
        pm.logger.Error("Plugin %s is not loaded", name)
        return false
    }
    
    plugin.OnDisable()
    delete(pm.plugins, name)
    
    pm.scheduler.CancelTasks(plugin)
    
    pm.logger.Info("Plugin %s unloaded successfully", name)
    return true
}

func (pm *PluginManagerImpl) GetPlugin(name string) Plugin {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    
    return pm.plugins[name]
}

func (pm *PluginManagerImpl) GetPlugins() []Plugin {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    
    plugins := make([]Plugin, 0, len(pm.plugins))
    for _, plugin := range pm.plugins {
        plugins = append(plugins, plugin)
    }
    return plugins
}

func (pm *PluginManagerImpl) RegisterEvent(eventType EventType, handler EventHandler) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    pm.eventHandlers[eventType] = append(pm.eventHandlers[eventType], handler)
    pm.logger.Debug("Registered event handler for %v", eventType)
}

func (pm *PluginManagerImpl) RegisterCommand(command string, handler CommandHandler) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    pm.commands[command] = handler
    pm.logger.Debug("Registered command: /%s", command)
}

func (pm *PluginManagerImpl) CallEvent(event Event) {
    pm.mutex.RLock()
    handlers, exists := pm.eventHandlers[event.GetType()]
    pm.mutex.RUnlock()
    
    if !exists {
        return
    }
    
    for _, handler := range handlers {
        handler(event)
        if event.IsCancelled() {
            break
        }
    }
}

func (pm *PluginManagerImpl) ExecuteCommand(sender CommandSender, commandLine string) bool {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    
    command, args := parseCommand(commandLine)
    
    handler, exists := pm.commands[command]
    if !exists {
        sender.SendMessage("Unknown command: " + command)
        return false
    }
    
    return handler(sender, command, args)
}

func parseCommand(commandLine string) (string, []string) {
    if len(commandLine) == 0 {
        return "", nil
    }
    
    if commandLine[0] == '/' {
        commandLine = commandLine[1:]
    }
    
    var parts []string
    var current string
    inQuotes := false
    
    for _, char := range commandLine {
        if char == '"' {
            inQuotes = !inQuotes
        } else if char == ' ' && !inQuotes {
            if current != "" {
                parts = append(parts, current)
                current = ""
            }
        } else {
            current += string(char)
        }
    }
    
    if current != "" {
        parts = append(parts, current)
    }
    
    if len(parts) == 0 {
        return "", nil
    }
    
    return parts[0], parts[1:]
}

func (pm *PluginManagerImpl) GetServer() Server {
    return pm.server
}

func (pm *PluginManagerImpl) BroadcastMessage(message string) {
    pm.server.BroadcastMessage(message)
}

func (pm *PluginManagerImpl) GetPlayer(name string) world.Player {
    // This will be implemented when we integrate with server
    return nil
}

func (pm *PluginManagerImpl) GetOnlinePlayers() []world.Player {
    // This will be implemented when we integrate with server
    return nil
}

func (pm *PluginManagerImpl) Tick() {
    pm.scheduler.Tick()
}

func (pm *PluginManagerImpl) GetScheduler() *SimpleScheduler {
    return pm.scheduler
}

func (pm *PluginManagerImpl) GetPermissionManager() *SimplePermissionManager {
    return pm.permissions
}

func (pm *PluginManagerImpl) DispatchPlayerJoin(player world.Player) {
    event := &PlayerJoinEvent{Player: player}
    pm.CallEvent(event)
}

func (pm *PluginManagerImpl) DispatchPlayerQuit(player world.Player, reason string) {
    event := &PlayerQuitEvent{Player: player, Reason: reason}
    pm.CallEvent(event)
}

func (pm *PluginManagerImpl) DispatchPlayerChat(player world.Player, message string) *PlayerChatEvent {
    event := &PlayerChatEvent{Player: player, Message: message}
    pm.CallEvent(event)
    return event
}

func (pm *PluginManagerImpl) DispatchPlayerMove(player world.Player, fromPos, toPos [3]float32) *PlayerMoveEvent {
    event := &PlayerMoveEvent{Player: player, FromPos: fromPos, ToPos: toPos}
    pm.CallEvent(event)
    return event
}

func (pm *PluginManagerImpl) DispatchPlayerBreakBlock(player world.Player, blockPos [3]int, blockID uint32) *PlayerBreakBlockEvent {
    event := &PlayerBreakBlockEvent{Player: player, BlockPos: blockPos, BlockID: blockID}
    pm.CallEvent(event)
    return event
}

func (pm *PluginManagerImpl) DispatchPlayerPlaceBlock(player world.Player, blockPos [3]int, blockID uint32) *PlayerPlaceBlockEvent {
    event := &PlayerPlaceBlockEvent{Player: player, BlockPos: blockPos, BlockID: blockID}
    pm.CallEvent(event)
    return event
}