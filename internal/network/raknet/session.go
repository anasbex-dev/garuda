package raknet

import (
    "bufio"
    "fmt"
    "log"
    "math"
    "net"
    "strings"
    "sync"
    "time"

    "garuda/internal/protocol/minecraft"
    "garuda/internal/world"
)

type SessionState int

const (
    StateUnconnected SessionState = iota
    StateConnecting
    StateConnected
    StateSpawned
    StateDisconnected
)

type Session struct {
    address        *net.UDPAddr
    conn           *net.UDPConn
    reader         *bufio.Reader
    writer         *bufio.Writer
    mtuSize        int
    guid           int64
    state          SessionState
    lastActivity   time.Time
    server         *RakNetServer
    reliableManager *ReliableManager
    mutex          sync.RWMutex
    
    // Game session data
    playerName     string
    playerID       int64
    entityID       int64
    world          *world.World
    playerEntity   *world.Player
    inventory      *world.Inventory
    
    // Network buffers
    packetChan     chan *minecraft.Packet
    disconnectChan chan bool
    closeChan      chan bool
    
    // Session timing
    pingTime       time.Time
    latency        time.Duration
}

func (s *RakNetServer) createSession(addr *net.UDPAddr) *Session {
    session := &Session{
        address:       addr,
        mtuSize:       minMTUSize,
        state:         StateUnconnected,
        lastActivity:  time.Now(),
        server:        s,
        packetChan:    make(chan *minecraft.Packet, 100),
        disconnectChan: make(chan bool, 1),
        closeChan:     make(chan bool, 1),
        latency:       0,
    }
    
    session.reliableManager = NewReliableManager(session)
    return session
}

func (s *Session) UpdateActivity() {
    s.mutex.Lock()
    s.lastActivity = time.Now()
    s.mutex.Unlock()
}

func (s *Session) GetState() SessionState {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.state
}

func (s *Session) SetState(state SessionState) {
    s.mutex.Lock()
    s.state = state
    s.mutex.Unlock()
}

func (s *Session) Handle() {
    defer s.Close()
    
    log.Printf("Session started for %s", s.address)
    
    // Main session loop
    for {
        select {
        case <-s.closeChan:
            return
        case <-s.disconnectChan:
            s.handleDisconnect()
            return
        case packet := <-s.packetChan:
            s.handleGamePacket(packet)
        default:
            // Handle incoming packets with timeout
            s.handleIncomingPackets()
        }
    }
}

func (s *Session) handleIncomingPackets() {
    // Set read deadline for non-blocking read
    deadline := time.Now().Add(100 * time.Millisecond)
    
    buffer := make([]byte, s.mtuSize)
    s.conn.SetReadDeadline(deadline)
    
    n, addr, err := s.conn.ReadFromUDP(buffer)
    if err != nil {
        if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
            return // No data available, continue loop
        }
        log.Printf("Read error from %s: %v", s.address, err)
        s.disconnectChan <- true
        return
    }
    
    if n > 0 && addr.String() == s.address.String() {
        s.UpdateActivity()
        s.server.handlePacket(addr, buffer[:n])
    }
}

func (s *Session) handleGamePacket(packet *minecraft.Packet) {
    if packet == nil || len(packet.Data) < 1 {
        return
    }
    
    packetID := packet.Data[0]
    
    switch packetID {
    case minecraft.IDLogin:
        s.handleMinecraftLogin(packet.Data)
    case minecraft.IDClientToServerHandshake:
        s.handleClientHandshake(packet.Data)
    case minecraft.IDResourcePackClientResponse:
        s.handleResourcePackResponse(packet.Data)
    case minecraft.IDMovePlayer:
        s.handleMovePlayer(packet.Data)
    case minecraft.IDPlayerAction:
        s.handlePlayerAction(packet.Data)
    case minecraft.IDInventoryTransaction:
        s.handleInventoryTransaction(packet.Data)
    case minecraft.IDMobEquipment:
        s.handleMobEquipment(packet.Data)
    case minecraft.IDAnimate:
        s.handleAnimatePacket(packet.Data)
    case minecraft.IDText:
        s.handleTextPacket(packet.Data)
    case minecraft.IDCommandRequest:
        s.handleCommandRequest(packet.Data)
    case minecraft.IDInteract:
        s.handleInteractPacket(packet.Data)
    case minecraft.IDBlockPickRequest:
        s.handleBlockPickRequest(packet.Data)
    case minecraft.IDPlayerHotbar:
        s.handlePlayerHotbar(packet.Data)
    case minecraft.IDRequestChunkRadius:
        s.handleChunkRadiusRequest(packet.Data)
    case minecraft.IDContainerClose:
        s.handleContainerClose(packet.Data)
    default:
        log.Printf("Unhandled Minecraft packet ID: 0x%02x from %s", packetID, s.address)
    }
}

func (s *Session) handleMinecraftLogin(data []byte) {
    loginPacket, err := minecraft.DecodeLoginPacket(data)
    if err != nil {
        log.Printf("Error decoding login packet from %s: %v", s.address, err)
        s.disconnectWithReason("Invalid login packet")
        return
    }
    
    // Extract player name from connection request data (simplified)
    s.playerName = s.extractPlayerName(loginPacket.ConnectionRequestData)
    if s.playerName == "" {
        s.playerName = fmt.Sprintf("Player_%d", time.Now().Unix())
    }
    
    log.Printf("Player login: %s (protocol=%d) from %s", 
        s.playerName, loginPacket.ProtocolVersion, s.address)
    
    // Send login success
    playStatus, err := minecraft.EncodePlayStatusPacket(0) // Success
    if err != nil {
        log.Printf("Error encoding play status: %v", err)
        return
    }
    
    s.SendGamePacket(playStatus)
    s.SetState(StateConnected)
    
    // Initialize player in world
    s.initializePlayer()
    
    // Send resource packs info (empty for now)
    // Send resource pack stack
    
    // Send start game packet
    s.sendStartGamePacket()
}

func (s *Session) extractPlayerName(connectionData []byte) string {
    // Simplified player name extraction
    // In real implementation, you'd properly decode the JWT/chain data
    if len(connectionData) > 10 {
        // This is a simplified approach - real implementation would decode JWT
        return fmt.Sprintf("Player_%d", time.Now().UnixNano()%10000)
    }
    return ""
}

func (s *Session) initializePlayer() {
    // Generate entity ID
    s.entityID = s.generateEntityID()
    
    // Create player entity
    s.playerEntity = world.NewPlayer(s.entityID, s.playerName)
    s.inventory = s.playerEntity.Inventory
    
    // Add player to world
    s.world.AddPlayer(s.playerEntity)
    s.server.serverAPI.AddPlayer(s.playerEntity)
    
    log.Printf("Player %s initialized with entity ID %d", s.playerName, s.entityID)
}

func (s *Session) generateEntityID() int64 {
    return time.Now().UnixNano()
}

func (s *Session) sendStartGamePacket() {
    startGamePacket := &minecraft.StartGamePacket{
        EntityID:              s.entityID,
        RuntimeEntityID:       uint64(s.entityID),
        PlayerGameType:        int32(s.world.GetGameRule("gameType").Value.(int)),
        PlayerPosition:        s.world.GetSpawnPoint(),
        Rotation:             minecraft.Vector2{X: 0, Y: 0},
        Seed:                 s.world.GetSeed(),
        BiomeType:            1,
        BiomeName:            "plains", 
        Dimension:            0, // Overworld
        Generator:            1, // Flat
        WorldGameMode:        int32(s.world.GetGameRule("gameType").Value.(int)),
        Difficulty:           int32(s.world.GetDifficulty()),
        SpawnPosition:        minecraft.BlockPos{
            X: int32(s.world.GetSpawnPoint().X),
            Y: int32(s.world.GetSpawnPoint().Y),
            Z: int32(s.world.GetSpawnPoint().Z),
        },
        AchievementsDisabled: true,
        Time:                 int32(s.world.GetDayTime()),
        EduMode:              false,
        CommandsEnabled:      true,
        TexturePacksRequired: false,
        GameRules:           s.encodeGameRules(),
        Experiments:         []minecraft.Experiment{},
        ChunkRadius:         int32(s.world.GetGameRule("viewDistance").Value.(int)),
    }
    
    packetData, err := minecraft.EncodeStartGamePacket(startGamePacket)
    if err != nil {
        log.Printf("Error encoding start game packet: %v", err)
        return
    }
    
    s.SendGamePacket(packetData)
    s.SetState(StateSpawned)
    
    log.Printf("Player %s spawned in world", s.playerName)
    
    // Send existing entities to player
    s.sendExistingEntities()
    
    // Send time update
    s.sendTimeUpdate()
    
    // Send initial chunks
    s.sendInitialChunks()
}

func (s *Session) encodeGameRules() []minecraft.GameRule {
    // Convert world game rules to Minecraft format
    var rules []minecraft.GameRule
    // Add default rules for now
    rules = append(rules, minecraft.GameRule{Name: "naturalRegeneration", Value: true})
    rules = append(rules, minecraft.GameRule{Name: "doDaylightCycle", Value: true})
    rules = append(rules, minecraft.GameRule{Name: "doMobSpawning", Value: true})
    return rules
}

func (s *Session) sendExistingEntities() {
    entities := s.world.GetEntityManager().GetEntities()
    for _, entity := range entities {
        if entity.Type == world.EntityPlayer && entity.EntityID == s.entityID {
            continue // Skip self
        }
        
        s.sendEntityToClient(entity)
    }
}

func (s *Session) sendEntityToClient(entity *world.Entity) {
    var packetData []byte
    var err error
    
    switch entity.Type {
    case world.EntityItem:
        packetData, err = s.encodeItemEntityPacket(entity)
    default:
        packetData, err = s.encodeMobEntityPacket(entity)
    }
    
    if err != nil {
        log.Printf("Error encoding entity packet: %v", err)
        return
    }
    
    s.SendGamePacket(packetData)
}

func (s *Session) encodeMobEntityPacket(entity *world.Entity) ([]byte, error) {
    entityType := s.getEntityTypeString(entity.Type)
    
    packet := &minecraft.AddActorPacket{
        RuntimeID: entity.RuntimeID,
        Type:      entityType,
        Position:  entity.Position,
        Motion:    entity.Motion,
        Rotation:  entity.Rotation,
        Attributes: []minecraft.EntityAttribute{
            {
                Name:      "minecraft:health",
                MinValue:  0.0,
                MaxValue:  entity.MaxHealth,
                Value:     entity.Health,
                Default:   entity.MaxHealth,
            },
        },
    }
    
    return minecraft.EncodeAddActorPacket(packet)
}

func (s *Session) encodeItemEntityPacket(entity *world.Entity) ([]byte, error) {
    itemData, ok := entity.Data.(*world.ItemEntityData)
    if !ok {
        return nil, fmt.Errorf("entity is not an item entity")
    }
    
    // Use AddActor packet for item entities for now
    return s.encodeMobEntityPacket(entity)
}

func (s *Session) getEntityTypeString(entityType world.EntityType) string {
    switch entityType {
    case world.EntityZombie:
        return "minecraft:zombie"
    case world.EntitySkeleton:
        return "minecraft:skeleton"
    case world.EntityCreeper:
        return "minecraft:creeper"
    case world.EntitySpider:
        return "minecraft:spider"
    case world.EntityCow:
        return "minecraft:cow"
    case world.EntityPig:
        return "minecraft:pig"
    case world.EntitySheep:
        return "minecraft:sheep"
    case world.EntityChicken:
        return "minecraft:chicken"
    case world.EntityItem:
        return "minecraft:item"
    default:
        return "minecraft:zombie"
    }
}

func (s *Session) sendTimeUpdate() {
    // TODO: Implement time update packet
}

func (s *Session) sendInitialChunks() {
    centerX := int32(s.playerEntity.Position.X) >> 4
    centerZ := int32(s.playerEntity.Position.Z) >> 4
    radius := int32(s.world.GetGameRule("viewDistance").Value.(int))
    
    for x := centerX - radius; x <= centerX + radius; x++ {
        for z := centerZ - radius; z <= centerZ + radius; z++ {
            chunk := s.world.GetChunk(x, z)
            s.sendChunkToPlayer(chunk)
        }
    }
}

func (s *Session) sendChunkToPlayer(chunk *world.Chunk) {
    // TODO: Implement chunk sending
}

// Packet handling methods
func (s *Session) handleClientHandshake(data []byte) {
    log.Printf("Client handshake received from %s", s.playerName)
    // Handshake complete, continue with game setup
}

func (s *Session) handleResourcePackResponse(data []byte) {
    log.Printf("Resource pack response received from %s", s.playerName)
    // Resource packs handled, game can continue
}

func (s *Session) handleMovePlayer(data []byte) {
    movePacket, err := minecraft.DecodeMovePlayerPacket(data)
    if err != nil {
        log.Printf("Error decoding move player packet from %s: %v", s.playerName, err)
        return
    }
    
    if movePacket.RuntimeID != uint64(s.entityID) {
        log.Printf("Move packet for wrong entity: %d != %d", movePacket.RuntimeID, s.entityID)
        return
    }
    
    // Validate movement with physics
    if s.validateMovement(movePacket) {
        // Update player position
        oldPosition := s.playerEntity.Position
        s.playerEntity.Position = movePacket.Position
        s.playerEntity.Rotation = movePacket.Rotation
        s.playerEntity.OnGround = movePacket.OnGround
        
        // Update chunk coordinates if needed
        newChunkX := int32(movePacket.Position.X) >> 4
        newChunkZ := int32(movePacket.Position.Z) >> 4
        
        oldChunkX := int32(oldPosition.X) >> 4
        oldChunkZ := int32(oldPosition.Z) >> 4
        
        if newChunkX != oldChunkX || newChunkZ != oldChunkZ {
            s.playerEntity.ChunkCoord = world.ChunkCoord{X: newChunkX, Z: newChunkZ}
            s.handleChunkChange(oldChunkX, oldChunkZ, newChunkX, newChunkZ)
        }
        
        // Dispatch move event to plugins
        s.server.pluginManager.DispatchPlayerMove(s.playerEntity, oldPosition, movePacket.Position)
        
    } else {
        // Send correction packet
        s.sendPositionCorrection()
    }
}

func (s *Session) validateMovement(packet *minecraft.MovePlayerPacket) bool {
    currentPos := s.playerEntity.Position
    newPos := packet.Position
    
    // Calculate movement vector
    movement := minecraft.Vector3{
        X: newPos.X - currentPos.X,
        Y: newPos.Y - currentPos.Y,
        Z: newPos.Z - currentPos.Z,
    }
    
    // Check for unreasonable movement speed (anti-cheat)
    movementDistance := math.Sqrt(float64(movement.X*movement.X + movement.Y*movement.Y + movement.Z*movement.Z))
    if movementDistance > 10.0 { // Max movement per packet
        log.Printf("Suspicious movement distance from %s: %.2f", s.playerName, movementDistance)
        return false
    }
    
    // Check collision
    playerAABB := world.GetPlayerAABB(newPos)
    if s.world.CheckCollision(playerAABB) {
        log.Printf("Movement would cause collision for %s", s.playerName)
        return false
    }
    
    return true
}

func (s *Session) handleChunkChange(oldX, oldZ, newX, newZ int32) {
    radius := int32(s.world.GetGameRule("viewDistance").Value.(int))
    
    // Unload chunks that are now out of range
    for x := oldX - radius; x <= oldX + radius; x++ {
        for z := oldZ - radius; z <= oldZ + radius; z++ {
            if math.Abs(float64(x-newX)) > float64(radius) || math.Abs(float64(z-newZ)) > float64(radius) {
                s.world.UnloadChunk(x, z)
            }
        }
    }
    
    // Load new chunks that are now in range
    for x := newX - radius; x <= newX + radius; x++ {
        for z := newZ - radius; z <= newZ + radius; z++ {
            if math.Abs(float64(x-oldX)) > float64(radius) || math.Abs(float64(z-oldZ)) > float64(radius) {
                chunk := s.world.GetChunk(x, z)
                s.sendChunkToPlayer(chunk)
            }
        }
    }
}

func (s *Session) sendPositionCorrection() {
    correctionPacket := &minecraft.MovePlayerPacket{
        RuntimeID: uint64(s.entityID),
        Position:  s.playerEntity.Position,
        Rotation:  s.playerEntity.Rotation,
        Mode:      0, // Normal
        OnGround:  s.playerEntity.OnGround,
        Tick:      uint64(time.Now().UnixNano() / int64(time.Millisecond)),
    }
    
    packetData, err := minecraft.EncodeMovePlayerPacket(correctionPacket)
    if err != nil {
        log.Printf("Error encoding position correction: %v", err)
        return
    }
    
    s.SendGamePacket(packetData)
}

func (s *Session) handlePlayerAction(data []byte) {
    actionPacket, err := minecraft.DecodePlayerActionPacket(data)
    if err != nil {
        log.Printf("Error decoding player action packet from %s: %v", s.playerName, err)
        return
    }
    
    if actionPacket.RuntimeID != uint64(s.entityID) {
        return
    }
    
    switch actionPacket.Action {
    case 0: // Start break
        s.handleBlockBreakStart(actionPacket.Position, actionPacket.Face)
    case 1: // Abort break
        s.handleBlockBreakAbort(actionPacket.Position)
    case 2: // Stop break
        s.handleBlockBreakComplete(actionPacket.Position)
    case 3: // Get updated block
        s.handleGetUpdatedBlock(actionPacket.Position)
    case 4: // Drop item
        s.handleItemDrop()
    case 5: // Start sleep
        s.handleSleepStart(actionPacket.Position)
    case 6: // Stop sleep
        s.handleSleepStop()
    case 7: // Respawn
        s.handleRespawn()
    case 8: // Jump
        s.handleJump()
    case 9: // Start sprint
        s.handleSprintStart()
    case 10: // Stop sprint
        s.handleSprintStop()
    case 11: // Start sneak
        s.handleSneakStart()
    case 12: // Stop sneak
        s.handleSneakStop()
    case 13: // Creative player destroy block
        s.handleCreativeDestroy(actionPacket.Position)
    default:
        log.Printf("Unknown player action from %s: %d", s.playerName, actionPacket.Action)
    }
}

func (s *Session) handleBlockBreakStart(pos minecraft.BlockPos, face int32) {
    log.Printf("Player %s started breaking block at %d,%d,%d", s.playerName, pos.X, pos.Y, pos.Z)
    
    block := s.world.GetBlock(pos)
    if block.ID == 0 { // Air
        return
    }
    
    // Dispatch event to plugins
    if !s.server.pluginManager.DispatchBlockBreak(s.playerEntity, pos, block) {
        return // Plugin cancelled the break
    }
    
    // Send block break animation to other players
    // s.broadcastBlockBreakAnimation(pos, face)
}

func (s *Session) handleBlockBreakComplete(pos minecraft.BlockPos) {
    log.Printf("Player %s completed breaking block at %d,%d,%d", s.playerName, pos.X, pos.Y, pos.Z)
    
    block := s.world.GetBlock(pos)
    if block.ID == 0 { // Air
        return
    }
    
    // Dispatch event to plugins
    if !s.server.pluginManager.DispatchBlockBreak(s.playerEntity, pos, block) {
        return // Plugin cancelled the break
    }
    
    // Drop block as item
    if block.ID != 7 { // Don't drop bedrock
        droppedItem := &world.ItemStack{
            ID:    block.ID,
            Count: 1,
            Damage: block.Data,
        }
        
        // Add to player inventory or drop in world
        if !s.playerEntity.Inventory.AddItem(droppedItem) {
            s.dropItem(&minecraft.ItemStack{
                ID:     uint16(droppedItem.ID),
                Count:  droppedItem.Count,
                Damage: uint16(droppedItem.Damage),
            })
        }
    }
    
    // Set block to air
    s.world.SetBlock(pos, world.Block{ID: 0, Data: 0})
    
    // Send block update to all players
    s.broadcastBlockUpdate(pos, 0)
}

// Implement other handler methods...
func (s *Session) handleTextPacket(data []byte) {
    textPacket, err := minecraft.DecodeTextPacket(data)
    if err != nil {
        log.Printf("Error decoding text packet: %v", err)
        return
    }
    
    if textPacket.TextType == 1 { // Chat message
        message := strings.TrimSpace(textPacket.Message)
        
        // Dispatch chat event to plugins
        if !s.server.pluginManager.DispatchPlayerChat(s.playerEntity, message) {
            return // Plugin cancelled the chat
        }
        
        // Handle commands
        if strings.HasPrefix(message, "/") {
            s.handleCommand(message)
        } else {
            // Broadcast chat message
            s.server.serverAPI.BroadcastMessage("<" + s.playerName + "> " + message)
        }
    }
}

func (s *Session) handleCommand(command string) {
    parts := strings.Split(command[1:], " ")
    cmd := strings.ToLower(parts[0])
    args := parts[1:]
    
    // Dispatch command event to plugins
    if !s.server.pluginManager.DispatchPlayerCommand(s.playerEntity, cmd, args) {
        return // Plugin handled the command
    }
    
    // Handle built-in commands
    switch cmd {
    case "help":
        s.sendMessage("§6--- Garuda Commands ---")
        s.sendMessage("§a/help §7- Show this help")
        s.sendMessage("§a/tp <player> §7- Teleport to player")
        s.sendMessage("§a/gamemode <mode> §7- Change gamemode")
    case "tp":
        if len(args) < 1 {
            s.sendMessage("§cUsage: /tp <player>")
            return
        }
        // TODO: Implement teleport
        s.sendMessage("§cTeleport not implemented yet")
    default:
        s.sendMessage("§cUnknown command. Type /help for help.")
    }
}

func (s *Session) sendMessage(message string) {
    // TODO: Implement message sending to player
    log.Printf("[MSG to %s] %s", s.playerName, message)
}

// Network methods
func (s *Session) SendGamePacket(packetData []byte) error {
    if s.GetState() == StateDisconnected {
        return fmt.Errorf("session disconnected")
    }
    
    packet := &minecraft.Packet{
        Data: packetData,
    }
    
    encapsulated := &EncapsulatedPacket{
        Reliability: ReliabilityReliable,
        Data:        packetData,
    }
    
    return s.reliableManager.SendPacket(encapsulated, ReliabilityReliable)
}

func (s *Session) Close() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if s.state == StateDisconnected {
        return
    }
    
    s.state = StateDisconnected
    close(s.closeChan)
    
    // Remove player from world
    if s.playerEntity != nil {
        s.world.RemovePlayer(s.entityID)
        s.server.serverAPI.RemovePlayer(s.playerEntity)
        
        // Dispatch quit event to plugins
        s.server.pluginManager.DispatchPlayerQuit(s.playerEntity)
        
        log.Printf("Player %s disconnected", s.playerName)
    }
    
    log.Printf("Session closed for %s", s.address)
}

func (s *Session) disconnectWithReason(reason string) {
    disconnectPacket, err := minecraft.EncodeDisconnectPacket(false, reason)
    if err != nil {
        log.Printf("Error encoding disconnect packet: %v", err)
    } else {
        s.SendGamePacket(disconnectPacket)
    }
    s.Close()
}

func (s *Session) RemoteAddr() net.Addr {
    return s.address
}

func (s *Session) GetPlayerName() string {
    return s.playerName
}

func (s *Session) GetPlayerEntity() *world.Player {
    return s.playerEntity
}

// Implement other handler stubs...
func (s *Session) handleAnimatePacket(data []byte) {
    // Implement animate packet handling
}

func (s *Session) handleCommandRequest(data []byte) {
    // Implement command request handling
}

func (s *Session) handleInteractPacket(data []byte) {
    // Implement interact packet handling
}

func (s *Session) handleBlockPickRequest(data []byte) {
    // Implement block pick request handling
}

func (s *Session) handlePlayerHotbar(data []byte) {
    // Implement player hotbar handling
}

func (s *Session) handleChunkRadiusRequest(data []byte) {
    // Implement chunk radius request handling
}

func (s *Session) handleContainerClose(data []byte) {
    // Implement container close handling
}

func (s *Session) handleGetUpdatedBlock(pos minecraft.BlockPos) {}
func (s *Session) handleItemDrop() {}
func (s *Session) handleSleepStart(pos minecraft.BlockPos) {}
func (s *Session) handleSleepStop() {}
func (s *Session) handleRespawn() {}
func (s *Session) handleJump() {}
func (s *Session) handleSprintStart() {}
func (s *Session) handleSprintStop() {}
func (s *Session) handleSneakStart() {}
func (s *Session) handleSneakStop() {}
func (s *Session) handleCreativeDestroy(pos minecraft.BlockPos) {}
func (s *Session) handleBlockBreakAbort(pos minecraft.BlockPos) {}
func (s *Session) dropItem(item *minecraft.ItemStack) {}
func (s *Session) broadcastBlockUpdate(pos minecraft.BlockPos, blockID uint32) {}