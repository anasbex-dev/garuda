package world

import (
    "math"
)

type AABB struct {
    MinX float64
    MinY float64
    MinZ float64
    MaxX float64
    MaxY float64
    MaxZ float64
}

func (a *AABB) CollidesWith(other *AABB) bool {
    return a.MinX < other.MaxX && a.MaxX > other.MinX &&
           a.MinY < other.MaxY && a.MaxY > other.MinY &&
           a.MinZ < other.MaxZ && a.MaxZ > other.MinZ
}

func GetPlayerAABB(position Vector3) AABB {
    // Player collision box (0.6 width, 1.8 height)
    return AABB{
        MinX: position.X - 0.3,
        MinY: position.Y,
        MinZ: position.Z - 0.3,
        MaxX: position.X + 0.3,
        MaxY: position.Y + 1.8,
        MaxZ: position.Z + 0.3,
    }
}

func (w *World) CheckCollision(aabb AABB) bool {
    // Convert AABB to block coordinates
    minX := int32(math.Floor(aabb.MinX))
    maxX := int32(math.Ceil(aabb.MaxX))
    minY := int32(math.Floor(aabb.MinY))
    maxY := int32(math.Ceil(aabb.MaxY))
    minZ := int32(math.Floor(aabb.MinZ))
    maxZ := int32(math.Ceil(aabb.MaxZ))
    
    // Check all blocks in the AABB volume
    for x := minX; x <= maxX; x++ {
        for y := minY; y <= maxY; y++ {
            for z := minZ; z <= maxZ; z++ {
                chunkX := x >> 4
                chunkZ := z >> 4
                localX := x & 0xF
                localZ := z & 0xF
                
                chunk := w.GetChunk(chunkX, chunkZ)
                block := chunk.GetBlock(localX, y, localZ)
                
                if block.IsSolid() {
                    // Create AABB for this block
                    blockAABB := AABB{
                        MinX: float64(x),
                        MinY: float64(y),
                        MinZ: float64(z),
                        MaxX: float64(x + 1),
                        MaxY: float64(y + 1),
                        MaxZ: float64(z + 1),
                    }
                    
                    if aabb.CollidesWith(&blockAABB) {
                        return true
                    }
                }
            }
        }
    }
    
    return false
}

func (w *World) FindCollisionResponse(aabb AABB, velocity Vector3) Vector3 {
    // Simple collision response - stop movement in colliding direction
    response := velocity
    
    // Test X movement
    if velocity.X != 0 {
        testAABB := aabb
        testAABB.MinX += velocity.X
        testAABB.MaxX += velocity.X
        
        if w.CheckCollision(testAABB) {
            response.X = 0
        }
    }
    
    // Test Y movement
    if velocity.Y != 0 {
        testAABB := aabb
        testAABB.MinY += velocity.Y
        testAABB.MaxY += velocity.Y
        
        if w.CheckCollision(testAABB) {
            response.Y = 0
        }
    }
    
    // Test Z movement
    if velocity.Z != 0 {
        testAABB := aabb
        testAABB.MinZ += velocity.Z
        testAABB.MaxZ += velocity.Z
        
        if w.CheckCollision(testAABB) {
            response.Z = 0
        }
    }
    
    return response
}