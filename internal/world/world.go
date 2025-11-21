package world

import (
    "math/rand"
    "time"
)

type World struct {
    name         string
    seed         int64
    chunks       map[ChunkCoord]*Chunk
    chunksMutex  sync.RWMutex
    players      map[int64]*Player
    playersMutex sync.RWMutex
    generator    Generator
    entityManager *EntityManager
    random       *rand.Rand
    time         int64
    running      bool
}

func NewWorld(name string, seed int64) *World {
    world := &World{
        name:      name,
        seed:      seed,
        chunks:    make(map[ChunkCoord]*Chunk),
        players:   make(map[int64]*Player),
        generator: NewFlatGenerator(),
        random:    rand.New(rand.NewSource(seed)),
        time:      0,
        running:   true,
    }
    
    world.entityManager = NewEntityManager(world)
    
    // Start world tick loop
    go world.tickLoop()
    
    return world
}

func (w *World) tickLoop() {
    ticker := time.NewTicker(50 * time.Millisecond) // 20 ticks per second
    defer ticker.Stop()
    
    for w.running {
        <-ticker.C
        w.tick()
    }
}

func (w *World) tick() {
    w.time++
    
    // Update entities
    w.entityManager.Update()
    
    // Natural mob spawning
    if w.time%200 == 0 { // Every 10 seconds
        w.naturalMobSpawning()
    }
}

func (w *World) naturalMobSpawning() {
    // Simple mob spawning - spawn near players
    for _, player := range w.players {
        if w.random.Float32() < 0.3 { // 30% chance per player
            // Spawn position near player
            angle := w.random.Float64() * 2 * math.Pi
            distance := 10 + w.random.Float64()*10
            x := player.Position.X + float32(math.Cos(angle)*distance)
            z := player.Position.Z + float32(math.Sin(angle)*distance)
            
            // Find ground level
            y := w.findGroundLevel(int32(x), int32(z))
            
            if y > 0 {
                // Random mob type
                mobTypes := []EntityType{EntityZombie, EntitySkeleton, EntityCreeper, EntitySpider}
                mobType := mobTypes[w.random.Intn(len(mobTypes))]
                
                w.entityManager.CreateEntity(mobType, minecraft.Vector3{
                    X: x,
                    Y: float32(y) + 1.0,
                    Z: z,
                })
            }
        }
        
        // Spawn passive mobs occasionally
        if w.random.Float32() < 0.1 { // 10% chance
            angle := w.random.Float64() * 2 * math.Pi
            distance := 15 + w.random.Float64()*15
            x := player.Position.X + float32(math.Cos(angle)*distance)
            z := player.Position.Z + float32(math.Sin(angle)*distance)
            
            y := w.findGroundLevel(int32(x), int32(z))
            if y > 0 {
                passiveTypes := []EntityType{EntityCow, EntityPig, EntitySheep, EntityChicken}
                mobType := passiveTypes[w.random.Intn(len(passiveTypes))]
                
                w.entityManager.CreateEntity(mobType, minecraft.Vector3{
                    X: x,
                    Y: float32(y) + 1.0,
                    Z: z,
                })
            }
        }
    }
}

func (w *World) findGroundLevel(x, z int32) int32 {
    chunkX := x >> 4
    chunkZ := z >> 4
    localX := x & 0xF
    localZ := z & 0xF
    
    chunk := w.GetChunk(chunkX, chunkZ)
    
    // Find highest non-air block
    for y := ChunkHeight - 1; y >= 0; y-- {
        block := chunk.GetBlock(localX, int32(y), localZ)
        if block.ID != 0 { // Not air
            return int32(y)
        }
    }
    
    return -1
}

func (w *World) GetEntityManager() *EntityManager {
    return w.entityManager
}

func (w *World) Stop() {
    w.running = false
}