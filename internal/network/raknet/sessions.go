package raknet

import (
    "time"
)

func (s *Session) HandleGamePacket(packetData []byte) {
    if len(packetData) < 1 {
        return
    }
    
    packetID := packetData[0]
    
    switch packetID {
    case minecraft.IDLogin:
        s.handleMinecraftLogin(packetData)
    case minecraft.IDClientToServerHandshake:
        s.handleClientHandshake(packetData)
    case minecraft.IDResourcePackClientResponse:
        s.handleResourcePackResponse(packetData)
    case minecraft.IDMovePlayer:
        s.handleMovePlayer(packetData)
    case minecraft.IDPlayerAction:
        s.handlePlayerAction(packetData)
    default:
        log.Printf("Unhandled Minecraft packet ID: 0x%02x", packetID)
    }
}

func (s *Session) handleMovePlayer(data []byte) {
    movePacket, err := minecraft.DecodeMovePlayerPacket(data)
    if err != nil {
        log.Printf("Error decoding move player packet: %v", err)
        return
    }
    
    if movePacket.RuntimeID != uint64(s.entityID) {
        log.Printf("Move packet for wrong entity: %d != %d", movePacket.RuntimeID, s.entityID)
        return
    }
    
    // Validate movement with physics
    if s.validateMovement(movePacket) {
        // Update player position
        s.playerEntity.Position = movePacket.Position
        s.playerEntity.Rotation = movePacket.Rotation
        
        // Update chunk coordinates if needed
        newChunkX := int32(movePacket.Position.X) >> 4
        newChunkZ := int32(movePacket.Position.Z) >> 4
        
        if newChunkX != s.playerEntity.ChunkCoord.X || newChunkZ != s.playerEntity.ChunkCoord.Z {
            s.playerEntity.ChunkCoord.X = newChunkX
            s.playerEntity.ChunkCoord.Z = newChunkZ
            s.world.sendChunksToPlayer(s.playerEntity)
        }
        
        // Broadcast movement to other players (would implement later)
        // s.broadcastMovement(movePacket)
    } else {
        // Send correction packet
        s.sendPositionCorrection()
    }
}

func (s *Session) validateMovement(packet *minecraft.MovePlayerPacket) bool {
    // Simple validation - check if player is trying to move through blocks
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
        log.Printf("Suspicious movement distance: %.2f", movementDistance)
        return false
    }
    
    // Check collision
    playerAABB := world.GetPlayerAABB(newPos)
    if s.world.CheckCollision(playerAABB) {
        log.Printf("Movement would cause collision")
        return false
    }
    
    return true
}

func (s *Session) sendPositionCorrection() {
    correctionPacket := &minecraft.MovePlayerPacket{
        RuntimeID: uint64(s.entityID),
        Position:  s.playerEntity.Position,
        Rotation:  s.playerEntity.Rotation,
        Mode:      0, // Normal
        OnGround:  true,
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
        log.Printf("Error decoding player action packet: %v", err)
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
        // Handle block update
    case 4: // Drop item
        s.handleItemDrop()
    case 5: // Start sleep
        // Handle sleep
    case 6: // Stop sleep
        // Handle wake up
    case 7: // Respawn
        s.handleRespawn()
    case 8: // Jump
        // Handle jump
    case 9: // Start sprint
        s.handleSprintStart()
    case 10: // Stop sprint
        s.handleSprintStop()
    case 11: // Start sneak
        s.handleSneakStart()
    case 12: // Stop sneak
        s.handleSneakStop()
    default:
        log.Printf("Unknown player action: %d", actionPacket.Action)
    }
}

func (s *Session) handleBlockBreakStart(pos minecraft.BlockPos, face int32) {
    log.Printf("Player started breaking block at %d,%d,%d", pos.X, pos.Y, pos.Z)
    
    // Convert to world coordinates
    chunkX := pos.X >> 4
    chunkZ := pos.Z >> 4
    localX := pos.X & 0xF
    localZ := pos.Z & 0xF
    
    chunk := s.world.GetChunk(chunkX, chunkZ)
    block := chunk.GetBlock(localX, pos.Y, localZ)
    
    if block.ID == 0 { // Air
        return
    }
    
    // Send block break animation to other players
    // s.broadcastBlockBreakAnimation(pos, face)
}

func (s *Session) handleBlockBreakComplete(pos minecraft.BlockPos) {
    log.Printf("Player completed breaking block at %d,%d,%d", pos.X, pos.Y, pos.Z)
    
    // Convert to world coordinates
    chunkX := pos.X >> 4
    chunkZ := pos.Z >> 4
    localX := pos.X & 0xF
    localZ := pos.Z & 0xF
    
    chunk := s.world.GetChunk(chunkX, chunkZ)
    
    // Set block to air
    chunk.SetBlock(localX, pos.Y, localZ, world.Block{ID: 0, Data: 0})
    
    // Send block update to all players
    s.broadcastBlockUpdate(pos, 0)
}

func (s *Session) broadcastBlockUpdate(pos minecraft.BlockPos, blockID uint32) {
    updatePacket := &minecraft.UpdateBlockPacket{
        Position: pos,
        BlockID:  blockID,
        Flags:    1, // Network
        Layer:    0, // Normal layer
    }
    
    packetData, err := minecraft.EncodeUpdateBlockPacket(updatePacket)
    if err != nil {
        log.Printf("Error encoding block update: %v", err)
        return
    }
    
    // Broadcast to all players in world (simplified - just send to self for now)
    s.SendGamePacket(packetData)
}

func (s *Session) handleRespawn() {
    log.Printf("Player respawning")
    
    // Reset player position
    s.playerEntity.Position = minecraft.Vector3{X: 0, Y: 70, Z: 0}
    s.playerEntity.Rotation = minecraft.Vector2{X: 0, Y: 0}
    
    // Send new position
    s.sendPositionCorrection()
}

func (s *Session) handleSprintStart() {
    log.Printf("Player started sprinting")
}

func (s *Session) handleSprintStop() {
    log.Printf("Player stopped sprinting")
}

func (s *Session) handleSneakStart() {
    log.Printf("Player started sneaking")
}

func (s *Session) handleSneakStop() {
    log.Printf("Player stopped sneaking")
}

func (s *Session) handleItemDrop() {
    log.Printf("Player dropped item")
}