package server

import (
    "garuda/internal/config"
    "garuda/internal/network/raknet"
    "garuda/minecraft"
    "garuda/pkg/plugin"
    "garuda/pkg/utils"
    "garuda/world"
    "math"
    "sync"
    "time"
)

type Player struct {
    Username    string
    UUID        string
    EntityID    int64
    RuntimeID   uint64
    Session     *raknet.Session
    Position    [3]float32
    Rotation    [2]float32
    World       *world.World
    Gamemode    int32
    Health      float32
    IsSpawned   bool
    WorldPlayer *world.Player
}

type Server struct {
    raknetServer    *raknet.Server
    config          *config.Config
    logger          *utils.Logger
    players         map[string]*Player
    playerMutex     sync.RWMutex
    world           *world.World
    entityManager   *world.EntityManager
    pluginManager   *plugin.PluginManagerImpl
    combatManager   *world.CombatManager
    inventoryManager *world.InventoryManager
    protocolManager *protocol.ProtocolManager
    running         bool
    ticker          *time.Ticker
    tickCount       int64
}

// Server implementation for plugin interface
type ServerForPlugins struct {
    server *Server
}

func (s *ServerForPlugins) GetName() string {
    return "Garuda"
}

func (s *ServerForPlugins) GetVersion() string {
    return s.server.config.Server.Version
}

func (s *ServerForPlugins) GetMaxPlayers() int {
    return s.server.config.Server.MaxPlayers
}

func (s *ServerForPlugins) GetPlayerCount() int {
    s.server.playerMutex.RLock()
    defer s.server.playerMutex.RUnlock()
    return len(s.server.players)
}

func (s *ServerForPlugins) BroadcastPacket(packetData []byte) {
    s.server.BroadcastPacket(packetData)
}

func (s *ServerForPlugins) ExecuteCommand(command string) bool {
    consoleSender := &plugin.ConsoleCommandSender{}
    return s.server.pluginManager.ExecuteCommand(consoleSender, command)
}

func NewServer(raknetServer *raknet.Server, cfg *config.Config, logger *utils.Logger) *Server {
    worldInstance := world.NewWorld(cfg.World.Name, cfg.World.Seed, logger)
    entityManager := world.NewEntityManager(logger)
    combatManager := world.NewCombatManager(logger)
    inventoryManager := world.NewInventoryManager(logger)
    
    server := &Server{
        raknetServer:    raknetServer,
        config:          cfg,
        logger:          logger,
        players:         make(map[string]*Player),
        world:           worldInstance,
        entityManager:   entityManager,
        combatManager:   combatManager,
        inventoryManager: inventoryManager,
        protocolManager: protocolManager,
        running:         false,
        tickCount:       0,
    }
    
    // Initialize plugin manager with server reference
    if !protocolManager.SetServerVersion(cfg.Protocol.Version) {
        logger.Warn("Failed to set server version to %s, using latest", cfg.Protocol.Version)
    }
    
    serverPlugin := &ServerForPlugins{server: server}
    server.pluginManager = plugin.NewPluginManager(serverPlugin, logger)
    
    return server
}

func (s *Server) Start() error {
    s.running = true
    s.ticker = time.NewTicker(50 * time.Millisecond)
    go s.tickLoop()
    
    s.logger.Info("Garuda Minecraft Server starting...")
    s.logger.Info("Version: %s", s.config.Server.Version)
    s.logger.Info("Max Players: %d", s.config.Server.MaxPlayers)
    s.logger.Info("MOTD: %s", s.config.Server.MOTD)
    s.logger.Info("World: %s (Seed: %s)", s.config.World.Name, s.config.World.Seed)
    
    s.protocolManager.LogSupportedVersions
    
    s.loadPlugins()
    s.spawnInitialMobs()
    
    if err := s.raknetServer.Start(); err != nil {
        return err
    }
    
    s.logger.Info("Minecraft server is ready!")
    
    select {}
}

func (s *Server) loadPlugins() {
    s.logger.Info("Loading plugins...")
    
    // In real implementation, this would scan plugins directory
    // For now, we'll manually load essentials
    // essentials := &essentials.EssentialsPlugin{}
    // if s.pluginManager.LoadPlugin(essentials) {
    //     s.logger.Info("Loaded Essentials plugin")
    // }
    
    s.logger.Info("Plugin system ready (%d plugins loaded)", len(s.pluginManager.GetPlugins()))
}

func (s *Server) spawnInitialMobs() {
    positions := [][3]float32{
        {10, 60, 10},
        {-10, 60, -10},
        {15, 60, -5},
        {-5, 60, 15},
    }
    
    for i, pos := range positions {
        mobType := world.EntityZombie
        if i%2 == 0 {
            mobType = world.EntitySkeleton
        }
        s.entityManager.SpawnMob(mobType, s.world, pos)
    }
    
    s.logger.Info("Spawned initial mobs")
}

func (s *Server) tickLoop() {
    for range s.ticker.C {
        if !s.running {
            return
        }
        s.tick()
    }
}

func (s *Server) tick() {
    s.tickCount++
    
    s.pluginManager.Tick()
    s.entityManager.UpdateEntities()
    
    // Clean up old combat events setiap 10 detik
    if s.tickCount%200 == 0 {
        s.combatManager.ClearOldEvents(10 * time.Second)
    }
    
    s.playerMutex.RLock()
    defer s.playerMutex.RUnlock()
    
    for _, player := range s.players {
        s.updatePlayer(player)
    }
    
    if s.tickCount%20 == 0 {
        s.sendTimeUpdates()
        s.updateEntities()
    }
}

func (s *Server) updateEntities() {
    entities := s.entityManager.GetEntitiesInRange([3]float32{0, 60, 0}, 50.0)
    
    for _, entity := range entities {
        if !entity.IsAlive() {
            s.entityManager.RemoveEntity(entity.ID)
            removePacket := &minecraft.RemoveEntityPacket{EntityID: entity.ID}
            if packetData, err := removePacket.Encode(); err == nil {
                s.BroadcastPacket(packetData)
            }
            continue
        }
        
        s.broadcastEntityMovement(entity)
    }
}

func (s *Server) broadcastEntityMovement(entity *world.Entity) {
    if entity.Type == world.EntityPlayer {
        return
    }
    
    movePacket := &minecraft.MovePlayerPacket{
        EntityID: entity.ID,
        Position: entity.GetPosition(),
        Pitch:    entity.GetRotation()[0],
        Yaw:      entity.GetRotation()[1],
        HeadYaw:  entity.GetRotation()[1],
        Mode:     0,
        OnGround: true,
    }
    
    packetData, err := movePacket.Encode()
    if err != nil {
        s.logger.Error("Failed to encode entity move packet: %v", err)
        return
    }
    
    s.BroadcastPacket(packetData)
}

func (s *Server) updatePlayer(player *Player) {
    if !player.IsSpawned {
        return
    }
    
    player.WorldPlayer.SetPosition(player.Position)
    player.WorldPlayer.SetRotation(player.Rotation)
    
    s.sendMovementUpdates(player)
    s.sendNearbyEntities(player)
}

func (s *Server) sendNearbyEntities(player *Player) {
    if s.tickCount%100 != 0 {
        return
    }
    
    entities := s.entityManager.GetEntitiesInRange(player.Position, 50.0)
    
    for _, entity := range entities {
        if entity.ID == player.EntityID {
            continue
        }
        
        s.sendEntityToPlayer(player, entity)
    }
}

func (s *Server) sendEntityToPlayer(player *Player, entity *world.Entity) {
    var packet minecraft.Packet
    
    switch entity.Type {
    case world.EntityPlayer:
        return
    case world.EntityItem:
        itemPacket := &minecraft.AddItemEntityPacket{
            EntityID:  entity.ID,
            RuntimeID: uint64(entity.ID),
            Item:      minecraft.ItemStack{ID: 1, Count: 1, Data: 0},
            Position:  entity.GetPosition(),
            Velocity:  entity.GetVelocity(),
        }
        packet = itemPacket
    default:
        entityPacket := &minecraft.AddEntityPacket{
            EntityID:   entity.ID,
            RuntimeID:  uint64(entity.ID),
            EntityType: s.getEntityTypeString(entity.Type),
            Position:   entity.GetPosition(),
            Velocity:   entity.GetVelocity(),
            Pitch:      entity.GetRotation()[0],
            Yaw:        entity.GetRotation()[1],
        }
        packet = entityPacket
    }
    
    if packetData, err := packet.Encode(); err == nil {
        player.Session.SendMinecraftPacket(packetData)
    }
}

func (s *Server) getEntityTypeString(entityType int) string {
    switch entityType {
    case world.EntityZombie:
        return "minecraft:zombie"
    case world.EntitySkeleton:
        return "minecraft:skeleton"
    case world.EntityCreeper:
        return "minecraft:creeper"
    default:
        return "minecraft:unknown"
    }
}

func (s *Server) sendTimeUpdates() {
    timeValue := int32(s.tickCount % 24000)
    timePacket := &minecraft.SetTimePacket{Time: timeValue}
    
    packetData, err := timePacket.Encode()
    if err != nil {
        s.logger.Error("Failed to encode time packet: %v", err)
        return
    }
    
    s.BroadcastPacket(packetData)
}

func (s *Server) sendMovementUpdates(player *Player) {
    movePacket := &minecraft.MovePlayerPacket{
        EntityID: player.EntityID,
        Position: player.Position,
        Pitch:    player.Rotation[0],
        Yaw:      player.Rotation[1],
        HeadYaw:  player.Rotation[1],
        Mode:     0,
        OnGround: true,
    }
    
    packetData, err := movePacket.Encode()
    if err != nil {
        s.logger.Error("Failed to encode move packet: %v", err)
        return
    }
    
    s.BroadcastToOthers(player, packetData)
}

// ===== COMBAT SYSTEM INTEGRATION =====

func (s *Server) HandlePlayerAttack(session *raknet.Session, attackPacket *minecraft.PlayerActionPacket) {
    s.playerMutex.RLock()
    player, exists := s.players[session.GetAddress()]
    s.playerMutex.RUnlock()
    
    if !exists {
        return
    }
    
    // Calculate target entity based on player's look direction
    target := s.findTargetEntity(player)
    if target == nil {
        return
    }
    
    // Get player's equipped item
    weapon := player.WorldPlayer.GetSelectedItem()
    
    // Execute combat
    event := s.combatManager.Attack(player.WorldPlayer.Entity, target, &weapon)
    
    if event != nil {
        s.logger.Debug("Player %s attacked %s for %.1f damage", 
            player.Username, s.combatManager.GetEntityName(target), event.Damage)
        
        // TODO: Send combat packets to clients
        // Send damage indicator, sound effects, etc.
    }
}

func (s *Server) findTargetEntity(player *Player) *world.Entity {
    // Simple raycast untuk find target entity
    playerPos := player.Position
    playerRot := player.Rotation
    
    // Calculate look direction
    yaw := playerRot[1] * (math.Pi / 180)
    pitch := playerRot[0] * (math.Pi / 180)
    
    dirX := -float32(math.Sin(float64(yaw)) * math.Cos(float64(pitch)))
    dirY := -float32(math.Sin(float64(pitch)))
    dirZ := float32(math.Cos(float64(yaw)) * math.Cos(float64(pitch)))
    
    // Raycast for 5 blocks
    for dist := float32(1.0); dist <= 5.0; dist += 0.5 {
        checkPos := [3]float32{
            playerPos[0] + dirX * dist,
            playerPos[1] + dirY * dist,
            playerPos[2] + dirZ * dist,
        }
        
        // Check for entities at this position
        entities := s.entityManager.GetEntitiesInRange(checkPos, 1.5)
        for _, entity := range entities {
            if entity != player.WorldPlayer.Entity {
                return entity
            }
        }
    }
    
    return nil
}

// ===== INVENTORY SYSTEM INTEGRATION =====

func (s *Server) HandleInventoryTransaction(session *raknet.Session, transactionData []byte) {
    s.playerMutex.RLock()
    player, exists := s.players[session.GetAddress()]
    s.playerMutex.RUnlock()
    
    if !exists {
        return
    }
    
    // Process inventory transaction
    // TODO: Implement full transaction parsing
    s.logger.Debug("Inventory transaction from player %s", player.Username)
}

func (s *Server) HandleContainerClose(session *raknet.Session, windowID byte) {
    s.playerMutex.RLock()
    player, exists := s.players[session.GetAddress()]
    s.playerMutex.RUnlock()
    
    if !exists {
        return
    }
    
    s.logger.Debug("Player %s closed container window %d", player.Username, windowID)
}

// ===== PLAYER MANAGEMENT =====

func (s *Server) HandleLogin(session *raknet.Session, loginPacket *minecraft.LoginPacket) {
    s.logger.Info("Handling login from session %s", session.GetAddress())
    
    statusPacket := &minecraft.PlayStatusPacket{Status: 0}
    if packetData, err := statusPacket.Encode(); err == nil {
        session.SendMinecraftPacket(packetData)
    }
    
    player := s.createPlayer(session)
    
    s.playerMutex.Lock()
    s.players[session.GetAddress()] = player
    s.playerMutex.Unlock()
    
    s.sendStartGame(player)
    s.sendSpawnChunks(player)
    s.sendInventory(player)
    
    player.IsSpawned = true
    
    // Dispatch plugin event
    s.pluginManager.DispatchPlayerJoin(player.WorldPlayer)
    
    s.logger.Info("Player %s joined the game", player.Username)
    s.BroadcastMessage("Player " + player.Username + " joined the game")
}

func (s *Server) createPlayer(session *raknet.Session) *Player {
    worldPlayer := world.NewPlayer("Player_"+session.GetAddress(), utils.FormatUUID(utils.GenerateUUID()), s.world, [3]float32{0, 60, 0})
    
    return &Player{
        Username:    "Player_" + session.GetAddress(),
        UUID:        utils.FormatUUID(utils.GenerateUUID()),
        EntityID:    s.generateEntityID(),
        RuntimeID:   uint64(s.generateEntityID()),
        Session:     session,
        Position:    [3]float32{0, 60, 0},
        Rotation:    [2]float32{0, 0},
        World:       s.world,
        Gamemode:    1,
        Health:      20.0,
        IsSpawned:   false,
        WorldPlayer: worldPlayer,
    }
}

func (s *Server) sendInventory(player *Player) {
    inventoryPacket := &minecraft.InventoryContentPacket{
        WindowID: 0,
        Items:    make([]minecraft.ItemStack, 36),
    }
    
    for i := 0; i < 36; i++ {
        item := player.WorldPlayer.GetItemInSlot(i)
        inventoryPacket.Items[i] = minecraft.ItemStack{
            ID:    item.ID,
            Count: item.Count,
            Data:  item.Data,
        }
    }
    
    if packetData, err := inventoryPacket.Encode(); err == nil {
        player.Session.SendMinecraftPacket(packetData)
    }
}

func (s *Server) HandleInventoryClick(session *raknet.Session, slotPacket *minecraft.InventorySlotPacket) {
    s.playerMutex.RLock()
    player, exists := s.players[session.GetAddress()]
    s.playerMutex.RUnlock()
    
    if !exists {
        return
    }
    
    item := world.ItemStack{
        ID:    slotPacket.Item.ID,
        Count: slotPacket.Item.Count,
        Data:  slotPacket.Item.Data,
    }
    
    player.WorldPlayer.SetItemInSlot(int(slotPacket.Slot), item)
    s.logger.Debug("Player %s updated slot %d to item %d", player.Username, slotPacket.Slot, item.ID)
}

func (s *Server) generateEntityID() int64 {
    return time.Now().UnixNano() + int64(len(s.players))
}

func (s *Server) sendStartGame(player *Player) {
    startGamePacket := &minecraft.StartGamePacket{
        EntityID:       player.EntityID,
        RuntimeID:      player.RuntimeID,
        PlayerGamemode: player.Gamemode,
        Position:       player.Position,
        WorldName:      s.config.World.Name,
    }
    
    if packetData, err := startGamePacket.Encode(); err == nil {
        player.Session.SendMinecraftPacket(packetData)
    }
}

func (s *Server) sendSpawnChunks(player *Player) {
    centerX := int32(player.Position[0]) >> 4
    centerZ := int32(player.Position[2]) >> 4
    
    viewDistance := s.config.World.ViewDistance
    if viewDistance <= 0 {
        viewDistance = 8
    }
    
    for x := centerX - int32(viewDistance); x <= centerX+int32(viewDistance); x++ {
        for z := centerZ - int32(viewDistance); z <= centerZ+int32(viewDistance); z++ {
            chunk := s.world.GetChunk(x, z)
            s.sendChunkToPlayer(player, chunk)
        }
    }
    
    s.logger.Debug("Sent %d chunks to player %s", (viewDistance*2+1)*(viewDistance*2+1), player.Username)
}

func (s *Server) sendChunkToPlayer(player *Player, chunk *world.Chunk) {
    chunkData := chunk.Encode()
    
    chunkPacket := &world.LevelChunkPacket{
        ChunkX:        chunk.X,
        ChunkZ:        chunk.Z,
        SubChunkCount: 16,
        Data:          chunkData,
    }
    
    if packetData, err := chunkPacket.Encode(); err == nil {
        player.Session.SendMinecraftPacket(packetData)
    }
}

func (s *Server) HandleDisconnect(session *raknet.Session) {
    s.playerMutex.Lock()
    defer s.playerMutex.Unlock()
    
    if player, exists := s.players[session.GetAddress()]; exists {
        s.pluginManager.DispatchPlayerQuit(player.WorldPlayer, "disconnected")
        s.logger.Info("Player %s left the game", player.Username)
        s.BroadcastMessage("Player " + player.Username + " left the game")
        delete(s.players, session.GetAddress())
    }
}

func (s *Server) HandleMovePlayer(session *raknet.Session, movePacket *minecraft.MovePlayerPacket) {
    s.playerMutex.RLock()
    player, exists := s.players[session.GetAddress()]
    s.playerMutex.RUnlock()
    
    if !exists {
        return
    }
    
    fromPos := player.Position
    player.Position = movePacket.Position
    player.Rotation = [2]float32{movePacket.Pitch, movePacket.Yaw}
    
    // Dispatch plugin event
    event := s.pluginManager.DispatchPlayerMove(player.WorldPlayer, fromPos, player.Position)
    if event.IsCancelled() {
        player.Position = fromPos // Revert position if cancelled
        return
    }
    
    s.logger.Debug("Player %s moved to %.1f,%.1f,%.1f", 
        player.Username, 
        player.Position[0], 
        player.Position[1], 
        player.Position[2])
}

func (s *Server) HandleChatMessage(session *raknet.Session, textPacket *minecraft.TextPacket) {
    s.playerMutex.RLock()
    player, exists := s.players[session.GetAddress()]
    s.playerMutex.RUnlock()
    
    if !exists || textPacket.Message == "" {
        return
    }
    
    // Dispatch plugin event
    processedMessage := s.pluginManager.DispatchPlayerChat(player.WorldPlayer, textPacket.Message)
    if processedMessage == "" {
        return // Message was cancelled by plugin
    }
    
    chatMessage := "<" + player.Username + "> " + processedMessage
    s.logger.Info("[CHAT] %s", chatMessage)
    s.BroadcastMessage(chatMessage)
}

func (s *Server) HandleBlockBreak(session *raknet.Session, actionPacket *minecraft.PlayerActionPacket) {
    s.playerMutex.RLock()
    player, exists := s.players[session.GetAddress()]
    s.playerMutex.RUnlock()
    
    if !exists {
        return
    }
    
    blockPos := [3]int{int(actionPacket.Position.X), int(actionPacket.Position.Y), int(actionPacket.Position.Z)}
    block := s.world.GetBlock(blockPos[0], blockPos[1], blockPos[2])
    
    // Dispatch plugin event
    event := s.pluginManager.DispatchPlayerBreakBlock(player.WorldPlayer, blockPos, block.ID)
    if event.IsCancelled() {
        return
    }
    
    if s.world.BreakBlock(blockPos[0], blockPos[1], blockPos[2], player.WorldPlayer) {
        s.logger.Debug("Player %s broke block at %d,%d,%d", player.Username, blockPos[0], blockPos[1], blockPos[2])
    }
}

func (s *Server) HandleBlockPlace(session *raknet.Session, blockPacket *minecraft.UpdateBlockPacket) {
    s.playerMutex.RLock()
    player, exists := s.players[session.GetAddress()]
    s.playerMutex.RUnlock()
    
    if !exists {
        return
    }
    
    blockPos := [3]int{int(blockPacket.Position.X), int(blockPacket.Position.Y), int(blockPacket.Position.Z)}
    
    // Dispatch plugin event
    event := s.pluginManager.DispatchPlayerPlaceBlock(player.WorldPlayer, blockPos, blockPacket.BlockID)
    if event.IsCancelled() {
        return
    }
    
    if s.world.PlaceBlock(blockPos[0], blockPos[1], blockPos[2], blockPacket.BlockID, player.WorldPlayer) {
        s.logger.Debug("Player %s placed block %d at %d,%d,%d", 
            player.Username, blockPacket.BlockID, blockPos[0], blockPos[1], blockPos[2])
    }
}

// ===== BROADCAST METHODS =====

func (s *Server) BroadcastMessage(message string) {
    textPacket := &minecraft.TextPacket{
        TextType: 1,
        Message:  message,
    }
    
    packetData, err := textPacket.Encode()
    if err != nil {
        s.logger.Error("Failed to encode chat message: %v", err)
        return
    }
    
    s.BroadcastPacket(packetData)
}

func (s *Server) BroadcastPacket(packetData []byte) {
    s.playerMutex.RLock()
    defer s.playerMutex.RUnlock()
    
    for _, player := range s.players {
        if player.IsSpawned {
            player.Session.SendMinecraftPacket(packetData)
        }
    }
}

func (s *Server) BroadcastToOthers(sender *Player, packetData []byte) {
    s.playerMutex.RLock()
    defer s.playerMutex.RUnlock()
    
    for _, player := range s.players {
        if player != sender && player.IsSpawned {
            player.Session.SendMinecraftPacket(packetData)
        }
    }
}

func (s *Server) GetPlayerCount() int {
    s.playerMutex.RLock()
    defer s.playerMutex.RUnlock()
    return len(s.players)
}

func (s *Server) GetMaxPlayers() int {
    return s.config.Server.MaxPlayers
}

func (s *Server) Stop() {
    s.running = false
    if s.ticker != nil {
        s.ticker.Stop()
    }
    
    s.playerMutex.Lock()
    for _, player := range s.players {
        disconnectPacket := &minecraft.DisconnectPacket{
            HideDisconnectionScreen: false,
            Message: "Server closed",
        }
        if packetData, err := disconnectPacket.Encode(); err == nil {
            player.Session.SendMinecraftPacket(packetData)
        }
    }
    s.players = make(map[string]*Player)
    s.playerMutex.Unlock()
    
    s.raknetServer.Close()
    s.logger.Info("Minecraft server stopped")
}