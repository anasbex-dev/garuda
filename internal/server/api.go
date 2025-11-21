package server

import (
    "fmt"
    "log"
    "sort"
    "strings"
    "sync"
    "time"

    "garuda/internal/protocol/minecraft"
    "garuda/internal/world"
    "garuda/pkg/plugin"
)

// GarudaServer implements the main server API for plugins and systems
type GarudaServer struct {
    world        *world.World
    players      map[string]*world.Player
    playersMutex sync.RWMutex
    commands     map[string]CommandHandler
    commandsMutex sync.RWMutex
    broadcastChan chan BroadcastMessage
    running      bool
    startTime    time.Time
    statistics   *ServerStatistics
    banManager   *BanManager
    whitelist    *WhitelistManager
    ops          *OpsManager
    playerBalances map[string]int
    economyMutex   sync.RWMutex
}

// ServerStatistics holds server performance statistics
type ServerStatistics struct {
    Uptime              time.Duration
    TicksProcessed      uint64
    PlayersJoined       uint64
    PlayersLeft         uint64
    ChunksLoaded        uint64
    EntitiesSpawned     uint64
    PacketsProcessed    uint64
    MemoryAllocated     uint64
    mutex               sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
    Message   string
    Type      MessageType
    Target    BroadcastTarget
    Exclude   *world.Player
}

// MessageType defines the type of broadcast message
type MessageType int

const (
    MessageTypeChat MessageType = iota
    MessageTypeSystem
    MessageTypePopup
    MessageTypeTip
    MessageTypeAnnouncement
    MessageTypeJson
)

// BroadcastTarget defines who should receive the broadcast
type BroadcastTarget int

const (
    TargetAllPlayers BroadcastTarget = iota
    TargetOpsOnly
    TargetSpecificPlayers
    TargetWorld
)

// CommandHandler defines the interface for command handlers
type CommandHandler func(sender CommandSender, args []string) bool

// CommandSender defines who can execute commands
type CommandSender interface {
    GetName() string
    HasPermission(permission string) bool
    SendMessage(message string)
    IsPlayer() bool
    GetPlayer() *world.Player
}

// PlayerCommandSender implements CommandSender for players
type PlayerCommandSender struct {
    player *world.Player
}

// ConsoleCommandSender implements CommandSender for console
type ConsoleCommandSender struct {
    name string
}

// BanManager manages player bans
type BanManager struct {
    bannedPlayers map[string]BanEntry
    bannedIPs     map[string]BanEntry
    mutex         sync.RWMutex
}

// BanEntry represents a ban record
type BanEntry struct {
    Name       string
    IP         string
    Reason     string
    Source     string
    Created    time.Time
    Expires    *time.Time
    Permanent  bool
}

// WhitelistManager manages the whitelist
type WhitelistManager struct {
    enabled    bool
    players    map[string]WhitelistEntry
    mutex      sync.RWMutex
}

// WhitelistEntry represents a whitelist entry
type WhitelistEntry struct {
    Name    string
    AddedBy string
    AddedAt time.Time
}

// OpsManager manages server operators
type OpsManager struct {
    ops   map[string]OpEntry
    mutex sync.RWMutex
}

// OpEntry represents an operator entry
type OpEntry struct {
    Name      string
    Level     int
    BypassesPlayerLimit bool
}

// NewGarudaServer creates a new Garuda server instance
func NewGarudaServer(world *world.World) *GarudaServer {
    server := &GarudaServer{
        world:        world,
        players:      make(map[string]*world.Player),
        commands:     make(map[string]CommandHandler),
        broadcastChan: make(chan BroadcastMessage, 100),
        running:      true,
        startTime:    time.Now(),
        statistics:   &ServerStatistics{},
        banManager:   NewBanManager(),
        whitelist:    NewWhitelistManager(),
        ops:          NewOpsManager(),
        playerBalances: make(map[string]int),
    }

    // Register built-in commands
    server.registerBuiltinCommands()

    // Start background workers
    go server.broadcastLoop()
    go server.statisticsLoop()
    go server.autoSaveLoop()

    log.Printf("Garuda server API initialized")
    return server
}

// BroadcastMessage broadcasts a message to players
func (s *GarudaServer) BroadcastMessage(message string) {
    msg := BroadcastMessage{
        Message: message,
        Type:    MessageTypeSystem,
        Target:  TargetAllPlayers,
    }
    s.broadcastChan <- msg
}

// BroadcastMessageToOps broadcasts a message to operators only
func (s *GarudaServer) BroadcastMessageToOps(message string) {
    msg := BroadcastMessage{
        Message: message,
        Type:    MessageTypeSystem,
        Target:  TargetOpsOnly,
    }
    s.broadcastChan <- msg
}

// GetPlayer returns a player by name
func (s *GarudaServer) GetPlayer(name string) *world.Player {
    s.playersMutex.RLock()
    defer s.playersMutex.RUnlock()

    // Case-insensitive search
    lowerName := strings.ToLower(name)
    for username, player := range s.players {
        if strings.ToLower(username) == lowerName {
            return player
        }
    }
    return nil
}

// GetOnlinePlayers returns all online players
func (s *GarudaServer) GetOnlinePlayers() []*world.Player {
    s.playersMutex.RLock()
    defer s.playersMutex.RUnlock()

    players := make([]*world.Player, 0, len(s.players))
    for _, player := range s.players {
        players = append(players, player)
    }
    
    // Sort by username for consistent ordering
    sort.Slice(players, func(i, j int) bool {
        return players[i].Username < players[j].Username
    })
    
    return players
}

// GetPlayerCount returns the number of online players
func (s *GarudaServer) GetPlayerCount() int {
    s.playersMutex.RLock()
    defer s.playersMutex.RUnlock()
    return len(s.players)
}

// AddPlayer adds a player to the server
func (s *GarudaServer) AddPlayer(player *world.Player) {
    s.playersMutex.Lock()
    defer s.playersMutex.Unlock()

    s.players[player.Username] = player
    
    // Update statistics
    s.statistics.mutex.Lock()
    s.statistics.PlayersJoined++
    s.statistics.mutex.Unlock()

    log.Printf("Player %s added to server (Total: %d)", player.Username, len(s.players))
}

// RemovePlayer removes a player from the server
func (s *GarudaServer) RemovePlayer(player *world.Player) {
    s.playersMutex.Lock()
    defer s.playersMutex.Unlock()

    delete(s.players, player.Username)
    
    // Update statistics
    s.statistics.mutex.Lock()
    s.statistics.PlayersLeft++
    s.statistics.mutex.Unlock()

    log.Printf("Player %s removed from server (Remaining: %d)", player.Username, len(s.players))
}

// ExecuteCommand executes a server command
func (s *GarudaServer) ExecuteCommand(command string) bool {
    parts := strings.Split(command, " ")
    if len(parts) == 0 {
        return false
    }

    cmd := strings.ToLower(parts[0])
    args := parts[1:]

    s.commandsMutex.RLock()
    handler, exists := s.commands[cmd]
    s.commandsMutex.RUnlock()

    if !exists {
        log.Printf("Unknown command: %s", cmd)
        return false
    }

    // Execute as console
    sender := &ConsoleCommandSender{name: "CONSOLE"}
    return handler(sender, args)
}

// ExecuteCommandAs executes a command as a specific sender
func (s *GarudaServer) ExecuteCommandAs(sender CommandSender, command string) bool {
    parts := strings.Split(command, " ")
    if len(parts) == 0 {
        return false
    }

    cmd := strings.ToLower(parts[0])
    args := parts[1:]

    s.commandsMutex.RLock()
    handler, exists := s.commands[cmd]
    s.commandsMutex.RUnlock()

    if !exists {
        sender.SendMessage(fmt.Sprintf("§cUnknown command: %s", cmd))
        return false
    }

    return handler(sender, args)
}

// RegisterCommand registers a new command
func (s *GarudaServer) RegisterCommand(command string, handler CommandHandler) {
    s.commandsMutex.Lock()
    defer s.commandsMutex.Unlock()

    s.commands[strings.ToLower(command)] = handler
    log.Printf("Command registered: /%s", command)
}

// UnregisterCommand removes a command
func (s *GarudaServer) UnregisterCommand(command string) {
    s.commandsMutex.Lock()
    defer s.commandsMutex.Unlock()

    delete(s.commands, strings.ToLower(command))
    log.Printf("Command unregistered: /%s", command)
}

// GetWorld returns the server world
func (s *GarudaServer) GetWorld() *world.World {
    return s.world
}

// GetServerStats returns server statistics
func (s *GarudaServer) GetServerStats() map[string]interface{} {
    s.statistics.mutex.RLock()
    defer s.statistics.mutex.RUnlock()

    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    return map[string]interface{}{
        "uptime":           s.statistics.Uptime.String(),
        "players_online":   len(s.players),
        "players_joined":   s.statistics.PlayersJoined,
        "players_left":     s.statistics.PlayersLeft,
        "ticks_processed":  s.statistics.TicksProcessed,
        "chunks_loaded":    s.statistics.ChunksLoaded,
        "entities_spawned": s.statistics.EntitiesSpawned,
        "packets_processed": s.statistics.PacketsProcessed,
        "memory_allocated": m.Alloc,
        "memory_system":    m.Sys,
        "start_time":       s.startTime.Format(time.RFC3339),
    }
}

// KickPlayer kicks a player from the server
func (s *GarudaServer) KickPlayer(player *world.Player, reason string) {
    // TODO: Implement player kicking logic
    log.Printf("Kicking player %s: %s", player.Username, reason)
    
    // Send disconnect packet
    // disconnectPacket := &minecraft.DisconnectPacket{
    //     HideDisconnectionScreen: false,
    //     Message: reason,
    // }
    // player.SendPacket(disconnectPacket)
    
    // Remove player from server
    s.RemovePlayer(player)
}

// BanPlayer bans a player from the server
func (s *GarudaServer) BanPlayer(name, reason, source string) bool {
    return s.banManager.BanPlayer(name, reason, source)
}

// PardonPlayer removes a player ban
func (s *GarudaServer) PardonPlayer(name string) bool {
    return s.banManager.PardonPlayer(name)
}

// IsPlayerBanned checks if a player is banned
func (s *GarudaServer) IsPlayerBanned(name string) bool {
    return s.banManager.IsPlayerBanned(name)
}

// AddToWhitelist adds a player to the whitelist
func (s *GarudaServer) AddToWhitelist(name, addedBy string) bool {
    return s.whitelist.AddPlayer(name, addedBy)
}

// RemoveFromWhitelist removes a player from the whitelist
func (s *GarudaServer) RemoveFromWhitelist(name string) bool {
    return s.whitelist.RemovePlayer(name)
}

// IsWhitelisted checks if a player is whitelisted
func (s *GarudaServer) IsWhitelisted(name string) bool {
    return s.whitelist.IsWhitelisted(name)
}

// SetWhitelistEnabled enables or disables the whitelist
func (s *GarudaServer) SetWhitelistEnabled(enabled bool) {
    s.whitelist.SetEnabled(enabled)
}

// AddOp adds a player as an operator
func (s *GarudaServer) AddOp(name string, level int) bool {
    return s.ops.AddOp(name, level)
}

// RemoveOp removes a player as an operator
func (s *GarudaServer) RemoveOp(name string) bool {
    return s.ops.RemoveOp(name)
}

// IsOp checks if a player is an operator
func (s *GarudaServer) IsOp(name string) bool {
    return s.ops.IsOp(name)
}

// GetOps returns all operators
func (s *GarudaServer) GetOps() []string {
    return s.ops.GetOps()
}

// SendMessageToPlayer sends a message to a specific player
func (s *GarudaServer) SendMessageToPlayer(player *world.Player, message string) {
    // TODO: Implement message sending to player
    log.Printf("[MSG to %s] %s", player.Username, message)
    
    // textPacket := &minecraft.TextPacket{
    //     TextType: minecraft.TextTypeSystem,
    //     Message: message,
    // }
    // player.SendPacket(textPacket)
}

// TeleportPlayer teleports a player to a location
func (s *GarudaServer) TeleportPlayer(player *world.Player, pos minecraft.Vector3) bool {
    oldPos := player.Position
    player.Position = pos
    
    log.Printf("Teleported player %s from %v to %v", 
        player.Username, oldPos, pos)
    
    // TODO: Send move player packet to client
    return true
}

// GiveItem gives an item to a player
func (s *GarudaServer) GiveItem(player *world.Player, item *world.ItemStack) bool {
    if player.Inventory.AddItem(item) {
        log.Printf("Gave item %d x%d to %s", item.ID, item.Count, player.Username)
        
        // TODO: Send inventory update packet
        return true
    }
    
    // Inventory full, drop item in world
    dropPos := minecraft.Vector3{
        X: player.Position.X,
        Y: player.Position.Y + 1.0,
        Z: player.Position.Z,
    }
    
    s.world.GetEntityManager().CreateItemEntity(dropPos, item)
    log.Printf("Dropped item %d x%d for %s (inventory full)", item.ID, item.Count, player.Username)
    
    return true
}

// SetGameRule sets a game rule
func (s *GarudaServer) SetGameRule(name string, value interface{}) bool {
    return s.world.SetGameRule(name, value)
}

// GetGameRule gets a game rule value
func (s *GarudaServer) GetGameRule(name string) interface{} {
    rule, exists := s.world.GetGameRule(name)
    if exists {
        return rule.Value
    }
    return nil
}

// SetTime sets the world time
func (s *GarudaServer) SetTime(time int32) {
    // TODO: Implement world time setting
    log.Printf("Setting world time to %d", time)
}

// SetWeather sets the world weather
func (s *GarudaServer) SetWeather(weatherType int, duration int) {
    // TODO: Implement weather setting
    log.Printf("Setting weather to %d for %d ticks", weatherType, duration)
}

// SaveWorld saves the world
func (s *GarudaServer) SaveWorld() bool {
    log.Printf("Saving world...")
    // TODO: Implement world saving
    return true
}

// Stop stops the server
func (s *GarudaServer) Stop() {
    s.running = false
    close(s.broadcastChan)
    
    log.Printf("Garuda server API stopped")
}

// ===== Background Workers =====

func (s *GarudaServer) broadcastLoop() {
    for msg := range s.broadcastChan {
        s.handleBroadcast(msg)
    }
}

func (s *GarudaServer) handleBroadcast(msg BroadcastMessage) {
    players := s.GetOnlinePlayers()
    
    for _, player := range players {
        // Check if player should be excluded
        if msg.Exclude != nil && player.Username == msg.Exclude.Username {
            continue
        }
        
        // Check broadcast target
        switch msg.Target {
        case TargetAllPlayers:
            // Send to all players
            s.SendMessageToPlayer(player, msg.Message)
        case TargetOpsOnly:
            // Send only to ops
            if s.IsOp(player.Username) {
                s.SendMessageToPlayer(player, msg.Message)
            }
        case TargetSpecificPlayers:
            // Specific target handling would be implemented here
        case TargetWorld:
            // Send to players in specific world (future feature)
            s.SendMessageToPlayer(player, msg.Message)
        }
    }
}

func (s *GarudaServer) statisticsLoop() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for s.running {
        <-ticker.C
        
        s.statistics.mutex.Lock()
        s.statistics.Uptime = time.Since(s.startTime)
        s.statistics.mutex.Unlock()
        
        // Log statistics periodically
        stats := s.GetServerStats()
        log.Printf("Server Stats - Uptime: %s, Players: %d, Memory: %.2fMB",
            stats["uptime"], stats["players_online"], 
            float64(stats["memory_allocated"].(uint64))/1024/1024)
    }
}

func (s *GarudaServer) autoSaveLoop() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for s.running {
        <-ticker.C
        s.SaveWorld()
    }
}

// ===== Built-in Commands =====

func (s *GarudaServer) registerBuiltinCommands() {
    s.RegisterCommand("help", s.helpCommand)
    s.RegisterCommand("list", s.listCommand)
    s.RegisterCommand("tp", s.teleportCommand)
    s.RegisterCommand("give", s.giveCommand)
    s.RegisterCommand("time", s.timeCommand)
    s.RegisterCommand("gamemode", s.gamemodeCommand)
    s.RegisterCommand("kick", s.kickCommand)
    s.RegisterCommand("ban", s.banCommand)
    s.RegisterCommand("pardon", s.pardonCommand)
    s.RegisterCommand("op", s.opCommand)
    s.RegisterCommand("deop", s.deopCommand)
    s.RegisterCommand("whitelist", s.whitelistCommand)
    s.RegisterCommand("save", s.saveCommand)
    s.RegisterCommand("stop", s.stopCommand)
}

func (s *GarudaServer) helpCommand(sender CommandSender, args []string) bool {
    sender.SendMessage("§6=== Garuda Server Commands ===")
    sender.SendMessage("§a/help §7- Show this help")
    sender.SendMessage("§a/list §7- List online players")
    sender.SendMessage("§a/tp <player> §7- Teleport to player")
    sender.SendMessage("§a/give <player> <item> [count] §7- Give item to player")
    sender.SendMessage("§a/gamemode <mode> [player] §7- Change gamemode")
    sender.SendMessage("§a/time set <value> §7- Set world time")
    sender.SendMessage("§a/kick <player> [reason] §7- Kick player")
    sender.SendMessage("§a/ban <player> [reason] §7- Ban player")
    sender.SendMessage("§a/op <player> §7- Make player operator")
    sender.SendMessage("§a/whitelist <add|remove|list> §7- Manage whitelist")
    sender.SendMessage("§a/save §7- Save the world")
    sender.SendMessage("§a/stop §7- Stop the server")
    return true
}

func (s *GarudaServer) listCommand(sender CommandSender, args []string) bool {
    players := s.GetOnlinePlayers()
    if len(players) == 0 {
        sender.SendMessage("§7No players online")
        return true
    }
    
    playerNames := make([]string, len(players))
    for i, player := range players {
        playerNames[i] = player.Username
    }
    
    sender.SendMessage(fmt.Sprintf("§7Online players (%d): §a%s", 
        len(players), strings.Join(playerNames, ", ")))
    return true
}

// Implement other command handlers...
func (s *GarudaServer) teleportCommand(sender CommandSender, args []string) bool {
    // Implementation for teleport command
    return true
}

func (s *GarudaServer) giveCommand(sender CommandSender, args []string) bool {
    // Implementation for give command
    return true
}

func (s *GarudaServer) timeCommand(sender CommandSender, args []string) bool {
    // Implementation for time command
    return true
}

func (s *GarudaServer) gamemodeCommand(sender CommandSender, args []string) bool {
    // Implementation for gamemode command
    return true
}

func (s *GarudaServer) kickCommand(sender CommandSender, args []string) bool {
    // Implementation for kick command
    return true
}

func (s *GarudaServer) banCommand(sender CommandSender, args []string) bool {
    // Implementation for ban command
    return true
}

func (s *GarudaServer) pardonCommand(sender CommandSender, args []string) bool {
    // Implementation for pardon command
    return true
}

func (s *GarudaServer) opCommand(sender CommandSender, args []string) bool {
    // Implementation for op command
    return true
}

func (s *GarudaServer) deopCommand(sender CommandSender, args []string) bool {
    // Implementation for deop command
    return true
}

func (s *GarudaServer) whitelistCommand(sender CommandSender, args []string) bool {
    // Implementation for whitelist command
    return true
}

func (s *GarudaServer) saveCommand(sender CommandSender, args []string) bool {
    if s.SaveWorld() {
        sender.SendMessage("§aWorld saved successfully")
    } else {
        sender.SendMessage("§cFailed to save world")
    }
    return true
}

func (s *GarudaServer) stopCommand(sender CommandSender, args []string) bool {
    if sender.HasPermission("server.stop") {
        sender.SendMessage("§cStopping server...")
        s.Stop()
        return true
    }
    sender.SendMessage("§cYou don't have permission to stop the server")
    return false
}

// ===== Command Sender Implementations =====

// PlayerCommandSender implementation
func (p *PlayerCommandSender) GetName() string {
    return p.player.Username
}

func (p *PlayerCommandSender) HasPermission(permission string) bool {
    // TODO: Implement permission system
    return p.player.Username == "CONSOLE" || s.ops.IsOp(p.player.Username)
}

func (p *PlayerCommandSender) SendMessage(message string) {
    s.SendMessageToPlayer(p.player, message)
}

func (p *PlayerCommandSender) IsPlayer() bool {
    return true
}

func (p *PlayerCommandSender) GetPlayer() *world.Player {
    return p.player
}

// ConsoleCommandSender implementation
func (c *ConsoleCommandSender) GetName() string {
    return c.name
}

func (c *ConsoleCommandSender) HasPermission(permission string) bool {
    return true // Console has all permissions
}

func (c *ConsoleCommandSender) SendMessage(message string) {
    log.Printf("[CONSOLE] %s", message)
}

func (c *ConsoleCommandSender) IsPlayer() bool {
    return false
}

func (c *ConsoleCommandSender) GetPlayer() *world.Player {
    return nil
}

// ===== Manager Implementations =====

// BanManager implementation
func NewBanManager() *BanManager {
    return &BanManager{
        bannedPlayers: make(map[string]BanEntry),
        bannedIPs:     make(map[string]BanEntry),
    }
}

func (bm *BanManager) BanPlayer(name, reason, source string) bool {
    bm.mutex.Lock()
    defer bm.mutex.Unlock()

    entry := BanEntry{
        Name:    name,
        Reason:  reason,
        Source:  source,
        Created: time.Now(),
        Permanent: true,
    }
    
    bm.bannedPlayers[strings.ToLower(name)] = entry
    log.Printf("Player %s banned by %s: %s", name, source, reason)
    return true
}

// Implement other BanManager methods...

// WhitelistManager implementation
func NewWhitelistManager() *WhitelistManager {
    return &WhitelistManager{
        enabled: false,
        players: make(map[string]WhitelistEntry),
    }
}

func (wm *WhitelistManager) AddPlayer(name, addedBy string) bool {
    wm.mutex.Lock()
    defer wm.mutex.Unlock()

    entry := WhitelistEntry{
        Name:    name,
        AddedBy: addedBy,
        AddedAt: time.Now(),
    }
    
    wm.players[strings.ToLower(name)] = entry
    log.Printf("Player %s added to whitelist by %s", name, addedBy)
    return true
}

// Implement other WhitelistManager methods...

// OpsManager implementation
func NewOpsManager() *OpsManager {
    return &OpsManager{
        ops: make(map[string]OpEntry),
    }
}

func (om *OpsManager) AddOp(name string, level int) bool {
    om.mutex.Lock()
    defer om.mutex.Unlock()

    entry := OpEntry{
        Name:    name,
        Level:   level,
        BypassesPlayerLimit: true,
    }
    
    om.ops[strings.ToLower(name)] = entry
    log.Printf("Player %s made operator (level %d)", name, level)
    return true
}

// Implement other OpsManager methods...


// ===== ENTITY MANAGEMENT =====

// SpawnEntity spawns a new entity in the world
func (s *GarudaServer) SpawnEntity(entityType world.EntityType, position minecraft.Vector3) *world.Entity {
    entity := s.world.GetEntityManager().CreateEntity(entityType, position)
    log.Printf("Spawned entity %v at %.1f,%.1f,%.1f", entityType, position.X, position.Y, position.Z)
    return entity
}

// RemoveEntity removes an entity from the world
func (s *GarudaServer) RemoveEntity(entity *world.Entity) {
    s.world.GetEntityManager().RemoveEntity(entity.EntityID)
    log.Printf("Removed entity %d", entity.EntityID)
}

// GetEntitiesInRange returns entities within radius of position
func (s *GarudaServer) GetEntitiesInRange(position minecraft.Vector3, radius float32) []*world.Entity {
    return s.world.GetEntityManager().GetEntitiesInRange(position, radius)
}

// GetEntityByID returns an entity by its ID
func (s *GarudaServer) GetEntityByID(entityID int64) *world.Entity {
    return s.world.GetEntityManager().GetEntity(entityID)
}

// ===== WORLD EDITING =====

// SetBlock sets a block at the specified position
func (s *GarudaServer) SetBlock(pos minecraft.BlockPos, block world.Block) {
    oldBlock := s.world.GetBlock(pos)
    s.world.SetBlock(pos, block)
    log.Printf("Set block at %d,%d,%d: %d -> %d", pos.X, pos.Y, pos.Z, oldBlock.ID, block.ID)
}

// GetBlock returns the block at the specified position
func (s *GarudaServer) GetBlock(pos minecraft.BlockPos) world.Block {
    return s.world.GetBlock(pos)
}

// FillBlocks fills an area with the specified block
func (s *GarudaServer) FillBlocks(start, end minecraft.BlockPos, block world.Block) int {
    count := 0
    minX, maxX := min(start.X, end.X), max(start.X, end.X)
    minY, maxY := min(start.Y, end.Y), max(start.Y, end.Y) 
    minZ, maxZ := min(start.Z, end.Z), max(start.Z, end.Z)
    
    for x := minX; x <= maxX; x++ {
        for y := minY; y <= maxY; y++ {
            for z := minZ; z <= maxZ; z++ {
                pos := minecraft.BlockPos{X: x, Y: y, Z: z}
                s.world.SetBlock(pos, block)
                count++
            }
        }
    }
    
    log.Printf("Filled %d blocks from %v to %v with block %d", count, start, end, block.ID)
    return count
}

// ===== INVENTORY MANAGEMENT =====

// GetPlayerInventory returns a player's inventory
func (s *GarudaServer) GetPlayerInventory(player *world.Player) *world.Inventory {
    return player.Inventory
}

// ClearPlayerInventory clears all items from a player's inventory
func (s *GarudaServer) ClearPlayerInventory(player *world.Player) {
    for i := 0; i < player.Inventory.size; i++ {
        player.Inventory.SetItem(i, &world.ItemStack{ID: 0, Count: 0})
    }
    log.Printf("Cleared inventory of player %s", player.Username)
}

// SetPlayerInventorySlot sets an item in a specific inventory slot
func (s *GarudaServer) SetPlayerInventorySlot(player *world.Player, slot int, item *world.ItemStack) bool {
    if slot < 0 || slot >= player.Inventory.size {
        return false
    }
    player.Inventory.SetItem(slot, item)
    log.Printf("Set slot %d of player %s to item %d x%d", slot, player.Username, item.ID, item.Count)
    return true
}

// GetPlayerInventorySlot returns the item in a specific inventory slot
func (s *GarudaServer) GetPlayerInventorySlot(player *world.Player, slot int) *world.ItemStack {
    if slot < 0 || slot >= player.Inventory.size {
        return nil
    }
    return player.Inventory.GetItem(slot)
}

// ===== PERMISSION SYSTEM =====

// HasPermission checks if a player has a specific permission
func (s *GarudaServer) HasPermission(player *world.Player, permission string) bool {
    // Basic implementation - ops have all permissions
    return s.IsOp(player.Username)
}

// SetPlayerPermission sets a permission for a player (basic implementation)
func (s *GarudaServer) SetPlayerPermission(player *world.Player, permission string, value bool) {
    // For now, just log the request
    log.Printf("Set permission %s for player %s to %t", permission, player.Username, value)
}

// GetPlayerPermissions returns all permissions for a player (basic implementation)
func (s *GarudaServer) GetPlayerPermissions(player *world.Player) map[string]bool {
    perms := make(map[string]bool)
    if s.IsOp(player.Username) {
        perms["*"] = true // Ops have all permissions
    }
    return perms
}

// ===== ECONOMY SYSTEM =====

// GetPlayerBalance returns a player's balance
func (s *GarudaServer) GetPlayerBalance(player *world.Player) int {
    s.economyMutex.RLock()
    defer s.economyMutex.RUnlock()
    
    if balance, exists := s.playerBalances[player.Username]; exists {
        return balance
    }
    return 0 // Default balance
}

// SetPlayerBalance sets a player's balance
func (s *GarudaServer) SetPlayerBalance(player *world.Player, amount int) {
    s.economyMutex.Lock()
    defer s.economyMutex.Unlock()
    
    s.playerBalances[player.Username] = amount
    log.Printf("Set balance of player %s to %d", player.Username, amount)
}

// AddPlayerBalance adds to a player's balance
func (s *GarudaServer) AddPlayerBalance(player *world.Player, amount int) int {
    s.economyMutex.Lock()
    defer s.economyMutex.Unlock()
    
    current := s.playerBalances[player.Username]
    newBalance := current + amount
    s.playerBalances[player.Username] = newBalance
    
    log.Printf("Added %d to player %s balance: %d -> %d", amount, player.Username, current, newBalance)
    return newBalance
}

// TransferBalance transfers balance between players
func (s *GarudaServer) TransferBalance(from, to *world.Player, amount int) bool {
    s.economyMutex.Lock()
    defer s.economyMutex.Unlock()
    
    fromBalance := s.playerBalances[from.Username]
    if fromBalance < amount {
        return false
    }
    
    s.playerBalances[from.Username] = fromBalance - amount
    s.playerBalances[to.Username] += amount
    
    log.Printf("Transferred %d from %s to %s", amount, from.Username, to.Username)
    return true
}

// ===== ENTITY CONTROL =====

// DamageEntity damages an entity
func (s *GarudaServer) DamageEntity(entity *world.Entity, damage float32, source *world.Entity) {
    s.world.GetCombatManager().DamageEntity(entity, damage, source)
}

// HealEntity heals an entity
func (s *GarudaServer) HealEntity(entity *world.Entity, amount float32) {
    if entity.Health+amount > entity.MaxHealth {
        entity.Health = entity.MaxHealth
    } else {
        entity.Health += amount
    }
}

// SetEntityMotion sets an entity's motion
func (s *GarudaServer) SetEntityMotion(entity *world.Entity, motion minecraft.Vector3) {
    entity.Motion = motion
}

// ===== WORLD INFORMATION =====

// GetWorldTime returns the current world time
func (s *GarudaServer) GetWorldTime() int64 {
    return s.world.GetTime()
}

// GetWorldName returns the world name
func (s *GarudaServer) GetWorldName() string {
    return s.world.GetName()
}

// GetWorldSeed returns the world seed
func (s *GarudaServer) GetWorldSeed() int64 {
    return s.world.GetSeed()
}

// ===== PLAYER STATE MANAGEMENT =====

// SetPlayerGameMode sets a player's game mode
func (s *GarudaServer) SetPlayerGameMode(player *world.Player, gameMode int32) {
    player.GameMode = gameMode
    log.Printf("Set game mode of player %s to %d", player.Username, gameMode)
}

// SetPlayerHealth sets a player's health
func (s *GarudaServer) SetPlayerHealth(player *world.Player, health float32) {
    if health < 0 {
        player.Health = 0
    } else if health > player.MaxHealth {
        player.Health = player.MaxHealth
    } else {
        player.Health = health
    }
}

// SetPlayerHunger sets a player's hunger
func (s *GarudaServer) SetPlayerHunger(player *world.Player, hunger int32) {
    if hunger < 0 {
        player.Hunger = 0
    } else if hunger > 20 {
        player.Hunger = 20
    } else {
        player.Hunger = hunger
    }
}

// ===== UTILITY FUNCTIONS =====

func min(a, b int32) int32 {
    if a < b {
        return a
    }
    return b
}

func max(a, b int32) int32 {
    if a > b {
        return a
    }
    return b
}