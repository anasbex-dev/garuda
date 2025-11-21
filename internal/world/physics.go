package world

import (
    "math"
)

type PhysicsEngine struct {
    world    *World
    registry *BlockRegistry
}

func NewPhysicsEngine(world *World, registry *BlockRegistry) *PhysicsEngine {
    return &PhysicsEngine{
        world:    world,
        registry: registry,
    }
}

func (pe *PhysicsEngine) UpdateBlockPhysics(x, y, z int) {
    block := pe.world.GetBlock(x, y, z)
    
    if pe.registry.IsLiquid(block.ID) {
        pe.updateLiquidPhysics(x, y, z, block.ID)
    }
    
    if !pe.isBlockSupported(x, y, z, block.ID) {
        pe.handleBlockFall(x, y, z, block.ID)
    }
    
    pe.updateBlockUpdates(x, y, z)
}

func (pe *PhysicsEngine) updateLiquidPhysics(x, y, z int, liquidID uint32) {
    if liquidID != 8 && liquidID != 9 && liquidID != 10 && liquidID != 11 {
        return
    }
    
    below := pe.world.GetBlock(x, y-1, z)
    if below.ID == 0 || pe.registry.IsLiquid(below.ID) {
        pe.world.SetBlock(x, y-1, z, Block{ID: liquidID})
        return
    }
    
    for _, offset := range [][3]int{{1, 0, 0}, {-1, 0, 0}, {0, 0, 1}, {0, 0, -1}} {
        nx, ny, nz := x+offset[0], y+offset[1], z+offset[2]
        neighbor := pe.world.GetBlock(nx, ny, nz)
        
        if neighbor.ID == 0 {
            pe.world.SetBlock(nx, ny, nz, Block{ID: liquidID})
        }
    }
}

func (pe *PhysicsEngine) isBlockSupported(x, y, z int, blockID uint32) bool {
    if blockID == 0 {
        return true
    }
    
    block := pe.registry.GetBlock(blockID)
    if block == nil {
        return true
    }
    
    if block.Material == "plant" || block.Material == "air" || pe.registry.IsLiquid(blockID) {
        return true
    }
    
    below := pe.world.GetBlock(x, y-1, z)
    if below.ID == 0 {
        return false
    }
    
    return pe.registry.IsSolid(below.ID)
}

func (pe *PhysicsEngine) handleBlockFall(x, y, z int, blockID uint32) {
    if !pe.canFall(blockID) {
        return
    }
    
    currentY := y - 1
    for currentY > 0 {
        below := pe.world.GetBlock(x, currentY, z)
        if below.ID != 0 && pe.registry.IsSolid(below.ID) {
            break
        }
        currentY--
    }
    
    fallY := currentY + 1
    
    if fallY != y {
        pe.world.SetBlock(x, y, z, Block{ID: 0})
        pe.world.SetBlock(x, fallY, z, Block{ID: blockID})
        
        for checkY := y - 1; checkY >= fallY; checkY-- {
            pe.UpdateBlockPhysics(x, checkY, z)
        }
    }
}

func (pe *PhysicsEngine) canFall(blockID uint32) bool {
    switch blockID {
    case 12, 13, 24, 80, 82: // Sand, Gravel, Sandstone, Soul Sand, Clay
        return true
    default:
        return false
    }
}

func (pe *PhysicsEngine) updateBlockUpdates(x, y, z int) {
    for _, offset := range [][3]int{
        {1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
        {1, 1, 0}, {1, -1, 0}, {-1, 1, 0}, {-1, -1, 0},
        {0, 1, 1}, {0, 1, -1}, {0, -1, 1}, {0, -1, -1},
        {1, 0, 1}, {1, 0, -1}, {-1, 0, 1}, {-1, 0, -1},
    } {
        nx, ny, nz := x+offset[0], y+offset[1], z+offset[2]
        if ny >= 0 && ny < ChunkHeight {
            block := pe.world.GetBlock(nx, ny, nz)
            if pe.registry.IsLiquid(block.ID) {
                pe.updateLiquidPhysics(nx, ny, nz, block.ID)
            }
        }
    }
}

func (pe *PhysicsEngine) CheckEntityCollision(entity *Entity) (bool, [3]float32) {
    entityPos := entity.GetPosition()
    entityBox := pe.getEntityBoundingBox(entity)
    
    minX := int(math.Floor(float64(entityBox[0][0])))
    maxX := int(math.Ceil(float64(entityBox[1][0])))
    minY := int(math.Floor(float64(entityBox[0][1])))
    maxY := int(math.Ceil(float64(entityBox[1][1])))
    minZ := int(math.Floor(float64(entityBox[0][2])))
    maxZ := int(math.Ceil(float64(entityBox[1][2])))
    
    collision := false
    correction := [3]float32{0, 0, 0}
    
    for x := minX; x <= maxX; x++ {
        for y := minY; y <= maxY; y++ {
            for z := minZ; z <= maxZ; z++ {
                block := pe.world.GetBlock(x, y, z)
                if block.ID != 0 && pe.registry.IsSolid(block.ID) {
                    blockBox := pe.getBlockBoundingBox(x, y, z, block.ID)
                    if pe.boxIntersect(entityBox, blockBox) {
                        collision = true
                        corr := pe.getCollisionCorrection(entityBox, blockBox)
                        correction[0] += corr[0]
                        correction[1] += corr[1]
                        correction[2] += corr[2]
                    }
                }
            }
        }
    }
    
    return collision, correction
}

func (pe *PhysicsEngine) getEntityBoundingBox(entity *Entity) [2][3]float32 {
    pos := entity.GetPosition()
    
    var width, height float32
    switch entity.Type {
    case EntityPlayer:
        width = 0.6
        height = 1.8
    case EntityZombie, EntitySkeleton:
        width = 0.6
        height = 1.95
    case EntityCreeper:
        width = 0.6
        height = 1.7
    case EntityItem:
        width = 0.25
        height = 0.25
    default:
        width = 0.6
        height = 1.8
    }
    
    halfWidth := width / 2
    return [2][3]float32{
        {pos[0] - halfWidth, pos[1], pos[2] - halfWidth},
        {pos[0] + halfWidth, pos[1] + height, pos[2] + halfWidth},
    }
}

func (pe *PhysicsEngine) getBlockBoundingBox(x, y, z int, blockID uint32) [2][3]float32 {
    block := pe.registry.GetBlock(blockID)
    if block == nil || block.BoundingBox == "empty" {
        return [2][3]float32{{0, 0, 0}, {0, 0, 0}}
    }
    
    fx := float32(x)
    fy := float32(y)
    fz := float32(z)
    
    return [2][3]float32{
        {fx, fy, fz},
        {fx + 1, fy + 1, fz + 1},
    }
}

func (pe *PhysicsEngine) boxIntersect(box1, box2 [2][3]float32) bool {
    return box1[0][0] < box2[1][0] && box1[1][0] > box2[0][0] &&
           box1[0][1] < box2[1][1] && box1[1][1] > box2[0][1] &&
           box1[0][2] < box2[1][2] && box1[1][2] > box2[0][2]
}

func (pe *PhysicsEngine) getCollisionCorrection(entityBox, blockBox [2][3]float32) [3]float32 {
    dx1 := blockBox[1][0] - entityBox[0][0]
    dx2 := blockBox[0][0] - entityBox[1][0]
    dy1 := blockBox[1][1] - entityBox[0][1]
    dy2 := blockBox[0][1] - entityBox[1][1]
    dz1 := blockBox[1][2] - entityBox[0][2]
    dz2 := blockBox[0][2] - entityBox[1][2]
    
    minDist := float32(math.MaxFloat32)
    correction := [3]float32{0, 0, 0}
    
    if dx1 >= 0 && dx1 < minDist {
        minDist = dx1
        correction = [3]float32{dx1, 0, 0}
    }
    if dx2 <= 0 && -dx2 < minDist {
        minDist = -dx2
        correction = [3]float32{dx2, 0, 0}
    }
    if dy1 >= 0 && dy1 < minDist {
        minDist = dy1
        correction = [3]float32{0, dy1, 0}
    }
    if dy2 <= 0 && -dy2 < minDist {
        minDist = -dy2
        correction = [3]float32{0, dy2, 0}
    }
    if dz1 >= 0 && dz1 < minDist {
        minDist = dz1
        correction = [3]float32{0, 0, dz1}
    }
    if dz2 <= 0 && -dz2 < minDist {
        minDist = -dz2
        correction = [3]float32{0, 0, dz2}
    }
    
    return correction
}

func (pe *PhysicsEngine) CanSeeSky(x, y, z int) bool {
    for checkY := y + 1; checkY < ChunkHeight; checkY++ {
        block := pe.world.GetBlock(x, checkY, z)
        if block.ID != 0 && !pe.registry.IsTransparent(block.ID) {
            return false
        }
    }
    return true
}

func (pe *PhysicsEngine) GetLightLevel(x, y, z int) int {
    if pe.CanSeeSky(x, y, z) {
        return 15
    }
    
    maxLight := 0
    for _, offset := range [][3]int{{0, 0, 0}, {1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
        nx, ny, nz := x+offset[0], y+offset[1], z+offset[2]
        if ny >= 0 && ny < ChunkHeight {
            block := pe.world.GetBlock(nx, ny, nz)
            light := pe.registry.GetBlock(block.ID).EmitLight
            if light > maxLight {
                maxLight = light
            }
        }
    }
    
    return maxLight
}