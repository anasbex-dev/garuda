package world

import (
    "garuda/pkg/utils"
    "sync"
)

type World struct {
    Name          string
    Seed          string
    Chunks        map[string]*Chunk
    BlockRegistry *BlockRegistry
    PhysicsEngine *PhysicsEngine
    mutex         sync.RWMutex
    logger        *utils.Logger
}

func NewWorld(name, seed string, logger *utils.Logger) *World {
    blockRegistry := NewBlockRegistry()
    world := &World{
        Name:          name,
        Seed:          seed,
        Chunks:        make(map[string]*Chunk),
        BlockRegistry: blockRegistry,
        logger:        logger,
    }
    
    world.PhysicsEngine = NewPhysicsEngine(world, blockRegistry)
    return world
}

func (w *World) GetChunk(x, z int32) *Chunk {
    key := chunkKey(x, z)
    
    w.mutex.RLock()
    chunk, exists := w.Chunks[key]
    w.mutex.RUnlock()
    
    if exists {
        return chunk
    }
    
    return w.GenerateChunk(x, z)
}

func (w *World) GenerateChunk(x, z int32) *Chunk {
    w.mutex.Lock()
    defer w.mutex.Unlock()
    
    key := chunkKey(x, z)
    
    if chunk, exists := w.Chunks[key]; exists {
        return chunk
    }
    
    chunk := NewChunk(x, z)
    w.generateTerrain(chunk)
    w.Chunks[key] = chunk
    
    w.logger.Debug("Generated chunk at %d, %d", x, z)
    return chunk
}

func (w *World) generateTerrain(chunk *Chunk) {
    for x := 0; x < ChunkWidth; x++ {
        for z := 0; z < ChunkLength; z++ {
            worldX := int(chunk.X)*ChunkWidth + x
            worldZ := int(chunk.Z)*ChunkLength + z
            
            height := w.getHeightAt(worldX, worldZ)
            
            chunk.SetBlock(x, 0, z, Block{ID: 7})
            
            for y := 1; y < height-3; y++ {
                chunk.SetBlock(x, y, z, Block{ID: 1})
            }
            
            for y := height - 3; y < height-1; y++ {
                chunk.SetBlock(x, y, z, Block{ID: 3})
            }
            
            chunk.SetBlock(x, height-1, z, Block{ID: 2})
            
            if x%4 == 0 && z%4 == 0 && height < ChunkHeight-2 {
                chunk.SetBlock(x, height, z, Block{ID: 17})
                for ly := height + 1; ly <= height+4; ly++ {
                    if ly < ChunkHeight {
                        chunk.SetBlock(x, ly, z, Block{ID: 18})
                    }
                }
            }
            
            if x%3 == 0 && z%3 == 0 {
                chunk.SetBlock(x, height, z, Block{ID: 31})
            } else if x%5 == 0 && z%5 == 0 {
                chunk.SetBlock(x, height, z, Block{ID: 37})
            }
        }
    }
}

func (w *World) getHeightAt(x, z int) int {
    noise := float32((x*x + z*z) % 10)
    return 55 + int(noise)
}

func (w *World) SetBlock(x, y, z int, block Block) {
    if y < 0 || y >= ChunkHeight {
        return
    }
    
    chunkX := x >> 4
    chunkZ := z >> 4
    localX := x & (ChunkWidth - 1)
    localZ := z & (ChunkLength - 1)
    
    chunk := w.GetChunk(int32(chunkX), int32(chunkZ))
    chunk.SetBlock(localX, y, localZ, block)
    
    w.PhysicsEngine.UpdateBlockPhysics(x, y, z)
}

func (w *World) GetBlock(x, y, z int) Block {
    if y < 0 || y >= ChunkHeight {
        return Block{ID: 0}
    }
    
    chunkX := x >> 4
    chunkZ := z >> 4
    localX := x & (ChunkWidth - 1)
    localZ := z & (ChunkLength - 1)
    
    chunk := w.GetChunk(int32(chunkX), int32(chunkZ))
    return chunk.GetBlock(localX, y, localZ)
}

func (w *World) BreakBlock(x, y, z int, player *Player) bool {
    block := w.GetBlock(x, y, z)
    if block.ID == 0 {
        return false
    }
    
    blockInfo := w.BlockRegistry.GetBlock(block.ID)
    if blockInfo == nil || !blockInfo.Diggable {
        return false
    }
    
    selectedItem := player.GetSelectedItem()
    canHarvest := w.BlockRegistry.CanHarvestWith(block.ID, selectedItem.ID)
    digTime := w.BlockRegistry.GetDigTime(block.ID, selectedItem.ID, canHarvest)
    
    if digTime < 0 {
        return false
    }
    
    w.SetBlock(x, y, z, Block{ID: 0})
    
    if canHarvest {
        w.dropBlockDrops(x, y, z, block.ID, player)
    }
    
    w.logger.Debug("Block broken at %d,%d,%d by %s", x, y, z, player.Username)
    return true
}

func (w *World) PlaceBlock(x, y, z int, blockID uint32, player *Player) bool {
    if blockID == 0 {
        return false
    }
    
    currentBlock := w.GetBlock(x, y, z)
    if currentBlock.ID != 0 {
        return false
    }
    
    if !w.canPlaceBlockAt(x, y, z, blockID) {
        return false
    }
    
    w.SetBlock(x, y, z, Block{ID: blockID})
    
    w.logger.Debug("Block placed at %d,%d,%d by %s", x, y, z, player.Username)
    return true
}

func (w *World) canPlaceBlockAt(x, y, z int, blockID uint32) bool {
    if y >= ChunkHeight-1 {
        return false
    }
    
    block := w.BlockRegistry.GetBlock(blockID)
    if block == nil {
        return false
    }
    
    if block.BoundingBox == "empty" {
        return true
    }
    
    for _, offset := range [][3]int{{0, 0, 0}, {0, 1, 0}} {
        checkX, checkY, checkZ := x+offset[0], y+offset[1], z+offset[2]
        neighbor := w.GetBlock(checkX, checkY, checkZ)
        if neighbor.ID != 0 && w.BlockRegistry.IsSolid(neighbor.ID) {
            return false
        }
    }
    
    return true
}

func (w *World) dropBlockDrops(x, y, z int, blockID uint32, player *Player) {
    drops := w.getBlockDrops(blockID, player)
    
    for _, drop := range drops {
        itemEntity := NewItemEntity(drop.ID, drop.Count, [3]float32{float32(x) + 0.5, float32(y) + 0.5, float32(z) + 0.5})
        w.PhysicsEngine.world.entityManager.SpawnItem(drop.ID, w, itemEntity.GetPosition())
    }
}

func (w *World) getBlockDrops(blockID uint32, player *Player) []ItemStack {
    switch blockID {
    case 1: // Stone
        return []ItemStack{{ID: 1, Count: 1, Data: 0}}
    case 2: // Grass
        return []ItemStack{{ID: 3, Count: 1, Data: 0}}
    case 3: // Dirt
        return []ItemStack{{ID: 3, Count: 1, Data: 0}}
    case 16: // Coal Ore
        return []ItemStack{{ID: 263, Count: 1, Data: 0}}
    case 15: // Iron Ore
        return []ItemStack{{ID: 265, Count: 1, Data: 0}}
    case 14: // Gold Ore
        return []ItemStack{{ID: 266, Count: 1, Data: 0}}
    case 56: // Diamond Ore
        return []ItemStack{{ID: 264, Count: 1, Data: 0}}
    case 17: // Oak Log
        return []ItemStack{{ID: 17, Count: 1, Data: 0}}
    case 18: // Oak Leaves
        if w.randomFloat() < 0.05 {
            return []ItemStack{{ID: 6, Count: 1, Data: 0}}
        }
        return nil
    default:
        return []ItemStack{{ID: blockID, Count: 1, Data: 0}}
    }
}

func (w *World) randomFloat() float32 {
    return float32((w.Seed[0] * byte(w.Seed[1])) % 100) / 100.0
}

func NewItemEntity(itemID uint32, count byte, position [3]float32) *Entity {
    entity := &Entity{
        ID:       -1,
        Type:     EntityItem,
        Position: position,
        Rotation: [2]float32{0, 0},
        Velocity: [3]float32{0, 0, 0},
        Health:   1.0,
        MaxHealth: 1.0,
        Metadata: make(map[string]interface{}),
    }
    
    entity.Metadata["item_id"] = itemID
    entity.Metadata["item_count"] = count
    
    return entity
}

func chunkKey(x, z int32) string {
    return string(rune(x)) + ":" + string(rune(z))
}   