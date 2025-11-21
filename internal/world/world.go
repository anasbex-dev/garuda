package world

import (
    "log"
    "math"
    "math/rand"
    "sync"
    "time"

    "github.com/anabex-dev/garuda/internal/protocol/minecraft"
)

type World struct {
    name          string
    seed          int64
    chunks        map[ChunkCoord]*Chunk
    chunksMutex   sync.RWMutex
    players       map[int64]*Player
    playersMutex  sync.RWMutex
    entities      map[int64]*Entity
    entitiesMutex sync.RWMutex
    generator     Generator
    entityManager *EntityManager
    combatManager *CombatManager
    random        *rand.Rand
    time          int64
    daytime       int64
    weather       WeatherType
    running       bool
    spawnPoint    minecraft.Vector3
    difficulty    int
    gameRules     map[string]GameRule
}

type WeatherType int

const (
    WeatherClear WeatherType = iota
    WeatherRain
    WeatherThunderstorm
)

type GameRule struct {
    Name  string
    Value interface{}
    Type  string // "bool", "int", "float"
}

func NewWorld(name string, seed int64) *World {
    world := &World{
        name:       name,
        seed:       seed,
        chunks:     make(map[ChunkCoord]*Chunk),
        players:    make(map[int64]*Player),
        entities:   make(map[int64]*Entity),
        generator:  NewFlatGenerator(),
        random:     rand.New(rand.NewSource(seed)),
        time:       0,
        daytime:    6000, // Start at noon
        weather:    WeatherClear,
        running:    true,
        spawnPoint: minecraft.Vector3{X: 0, Y: 70, Z: 0},
        difficulty: 2, // Normal
        gameRules:  make(map[string]GameRule),
    }

    // Initialize default game rules
    world.initializeGameRules()

    world.entityManager = NewEntityManager(world)
    world.combatManager = NewCombatManager(world)

    // Start world tick loop
    go world.tickLoop()

    log.Printf("World '%s' created with seed %d", name, seed)
    return world
}

func (w *World) initializeGameRules() {
    w.gameRules["doDaylightCycle"] = GameRule{Name: "doDaylightCycle", Value: true, Type: "bool"}
    w.gameRules["doWeatherCycle"] = GameRule{Name: "doWeatherCycle", Value: true, Type: "bool"}
    w.gameRules["doMobSpawning"] = GameRule{Name: "doMobSpawning", Value: true, Type: "bool"}
    w.gameRules["doFireTick"] = GameRule{Name: "doFireTick", Value: true, Type: "bool"}
    w.gameRules["mobGriefing"] = GameRule{Name: "mobGriefing", Value: true, Type: "bool"}
    w.gameRules["keepInventory"] = GameRule{Name: "keepInventory", Value: false, Type: "bool"}
    w.gameRules["naturalRegeneration"] = GameRule{Name: "naturalRegeneration", Value: true, Type: "bool"}
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
    w.daytime++

    // Day/night cycle (20 minutes = 24000 ticks)
    if w.daytime >= 24000 {
        w.daytime = 0
    }

    // Update entities
    w.entityManager.Update()

    // Update weather
    if w.time%12000 == 0 { // Every 10 minutes
        w.updateWeather()
    }

    // Natural mob spawning
    if w.getGameRuleBool("doMobSpawning") && w.time%200 == 0 {
        w.naturalMobSpawning()
    }

    // Player regeneration
    if w.getGameRuleBool("naturalRegeneration") && w.time%100 == 0 {
        w.updatePlayerRegeneration()
    }

    // Time-based events
    if w.time%20 == 0 { // Every second
        w.handleTimeBasedEvents()
    }
}

func (w *World) updateWeather() {
    if !w.getGameRuleBool("doWeatherCycle") {
        return
    }

    // Simple weather transition
    if w.random.Float32() < 0.3 {
        if w.weather == WeatherClear {
            w.weather = WeatherRain
            log.Printf("Weather changed to rain")
        } else if w.weather == WeatherRain && w.random.Float32() < 0.2 {
            w.weather = WeatherThunderstorm
            log.Printf("Weather changed to thunderstorm")
        }
    } else if w.random.Float32() < 0.1 {
        w.weather = WeatherClear
        log.Printf("Weather cleared")
    }
}

func (w *World) updatePlayerRegeneration() {
    w.playersMutex.RLock()
    defer w.playersMutex.RUnlock()

    for _, player := range w.players {
        if player.Health < player.MaxHealth && player.Hunger >= 18 {
            player.Health = math.Min(player.Health+1.0, player.MaxHealth)
            player.Hunger -= 0.5
        }

        // Hunger regeneration when not hungry
        if player.Hunger < 20 && w.time%800 == 0 {
            player.Hunger++
        }
    }
}

func (w *World) handleTimeBasedEvents() {
    // Time-specific events (monster spawning at night, etc.)
    isNight := w.daytime > 13000 && w.daytime < 23000

    if isNight {
        // Increase mob spawn rates at night
        if w.time%100 == 0 {
            w.naturalMobSpawning()
        }
    }
}

func (w *World) naturalMobSpawning() {
    w.playersMutex.RLock()
    players := make([]*Player, 0, len(w.players))
    for _, player := range w.players {
        players = append(players, player)
    }
    w.playersMutex.RUnlock()

    for _, player := range players {
        if w.random.Float32() < 0.3 {
            w.spawnMobsNearPlayer(player, 1, 3) // Spawn 1-3 mobs
        }
    }
}

func (w *World) spawnMobsNearPlayer(player *Player, minCount, maxCount int) {
    count := minCount + w.random.Intn(maxCount-minCount+1)

    for i := 0; i < count; i++ {
        // Spawn position near player but not too close
        angle := w.random.Float64() * 2 * math.Pi
        distance := 12 + w.random.Float64()*8
        x := player.Position.X + float32(math.Cos(angle)*distance)
        z := player.Position.Z + float32(math.Sin(angle)*distance)

        // Find suitable spawn location
        y := w.findSpawnLocation(int32(x), int32(z))
        if y > 0 {
            var mobType EntityType

            // Different mobs based on time and location
            isNight := w.daytime > 13000 && w.daytime < 23000
            if isNight && w.random.Float32() < 0.7 {
                // Hostile mobs at night
                hostileMobs := []EntityType{EntityZombie, EntitySkeleton, EntityCreeper, EntitySpider}
                mobType = hostileMobs[w.random.Intn(len(hostileMobs))]
            } else {
                // Passive mobs during day
                passiveMobs := []EntityType{EntityCow, EntityPig, EntitySheep, EntityChicken}
                mobType = passiveMobs[w.random.Intn(len(passiveMobs))]
            }

            entity := w.entityManager.CreateEntity(mobType, minecraft.Vector3{
                X: x,
                Y: float32(y) + 1.0,
                Z: z,
            })

            log.Printf("Spawned %s at %.1f,%.1f,%.1f near %s", 
                w.getEntityTypeName(mobType), x, float32(y)+1.0, z, player.Username)
        }
    }
}

func (w *World) findSpawnLocation(x, z int32) int32 {
    chunkX := x >> 4
    chunkZ := z >> 4
    localX := x & 0xF
    localZ := z & 0xF

    chunk := w.GetChunk(chunkX, chunkZ)

    // Find highest solid block with air above
    for y := ChunkHeight - 2; y >= 0; y-- {
        block := chunk.GetBlock(localX, int32(y), localZ)
        blockAbove := chunk.GetBlock(localX, int32(y+1), localZ)

        if block.ID != 0 && blockAbove.ID == 0 {
            // Check if block is spawnable (not leaves, glass, etc.)
            if w.isSpawnableBlock(block) {
                return int32(y)
            }
        }
    }

    return -1
}

func (w *World) isSpawnableBlock(block Block) bool {
    // Blocks that mobs can spawn on
    spawnableBlocks := map[uint16]bool{
        1: true,  // Stone
        2: true,  // Grass
        3: true,  // Dirt
        4: true,  // Cobblestone
        5: true,  // Wood
        6: true,  // Planks
        7: false, // Bedrock
        8: false, // Water
        9: false, // Water
        12: true, // Sand
        13: true, // Gravel
        17: true, // Wood
        18: false, // Leaves
    }

    return spawnableBlocks[block.ID]
}

func (w *World) getEntityTypeName(entityType EntityType) string {
    switch entityType {
    case EntityZombie:
        return "Zombie"
    case EntitySkeleton:
        return "Skeleton"
    case EntityCreeper:
        return "Creeper"
    case EntitySpider:
        return "Spider"
    case EntityCow:
        return "Cow"
    case EntityPig:
        return "Pig"
    case EntitySheep:
        return "Sheep"
    case EntityChicken:
        return "Chicken"
    default:
        return "Unknown"
    }
}

// World management methods
func (w *World) GetName() string {
    return w.name
}

func (w *World) GetSeed() int64 {
    return w.seed
}

func (w *World) GetTime() int64 {
    return w.time
}

func (w *World) GetDayTime() int64 {
    return w.daytime
}

func (w *World) GetSpawnPoint() minecraft.Vector3 {
    return w.spawnPoint
}

func (w *World) SetSpawnPoint(pos minecraft.Vector3) {
    w.spawnPoint = pos
}

func (w *World) GetDifficulty() int {
    return w.difficulty
}

func (w *World) SetDifficulty(difficulty int) {
    w.difficulty = difficulty
}

// Chunk management
func (w *World) GetChunk(x, z int32) *Chunk {
    coord := ChunkCoord{X: x, Z: z}

    w.chunksMutex.RLock()
    chunk, exists := w.chunks[coord]
    w.chunksMutex.RUnlock()

    if exists {
        return chunk
    }

    // Generate new chunk
    return w.generateChunk(x, z)
}

func (w *World) generateChunk(x, z int32) *Chunk {
    chunk := w.generator.GenerateChunk(x, z)

    w.chunksMutex.Lock()
    w.chunks[ChunkCoord{X: x, Z: z}] = chunk
    w.chunksMutex.Unlock()

    log.Printf("Generated chunk at %d,%d", x, z)
    return chunk
}

func (w *World) SaveChunk(x, z int32) error {
    // TODO: Implement chunk saving to disk
    return nil
}

func (w *World) UnloadChunk(x, z int32) {
    coord := ChunkCoord{X: x, Z: z}

    w.chunksMutex.Lock()
    delete(w.chunks, coord)
    w.chunksMutex.Unlock()

    log.Printf("Unloaded chunk at %d,%d", x, z)
}

// Block management
func (w *World) GetBlock(pos minecraft.BlockPos) Block {
    chunkX := pos.X >> 4
    chunkZ := pos.Z >> 4
    localX := pos.X & 0xF
    localZ := pos.Z & 0xF

    chunk := w.GetChunk(chunkX, chunkZ)
    return chunk.GetBlock(localX, pos.Y, localZ)
}

func (w *World) SetBlock(pos minecraft.BlockPos, block Block) {
    chunkX := pos.X >> 4
    chunkZ := pos.Z >> 4
    localX := pos.X & 0xF
    localZ := pos.Z & 0xF

    chunk := w.GetChunk(chunkX, chunkZ)
    chunk.SetBlock(localX, pos.Y, localZ, block)

    // TODO: Send block update to nearby players
}

// Player management
func (w *World) AddPlayer(player *Player) {
    w.playersMutex.Lock()
    defer w.playersMutex.Unlock()

    w.players[player.EntityID] = player
    log.Printf("Player %s joined world %s", player.Username, w.name)
}

func (w *World) RemovePlayer(entityID int64) {
    w.playersMutex.Lock()
    defer w.playersMutex.Unlock()

    if player, exists := w.players[entityID]; exists {
        log.Printf("Player %s left world %s", player.Username, w.name)
        delete(w.players, entityID)
    }
}

func (w *World) GetPlayer(entityID int64) *Player {
    w.playersMutex.RLock()
    defer w.playersMutex.RUnlock()

    return w.players[entityID]
}

func (w *World) GetPlayerByName(name string) *Player {
    w.playersMutex.RLock()
    defer w.playersMutex.RUnlock()

    for _, player := range w.players {
        if player.Username == name {
            return player
        }
    }
    return nil
}

func (w *World) GetOnlinePlayers() []*Player {
    w.playersMutex.RLock()
    defer w.playersMutex.RUnlock()

    players := make([]*Player, 0, len(w.players))
    for _, player := range w.players {
        players = append(players, player)
    }
    return players
}

// Entity management
func (w *World) GetEntityManager() *EntityManager {
    return w.entityManager
}

func (w *World) GetCombatManager() *CombatManager {
    return w.combatManager
}

// Game rules
func (w *World) GetGameRule(name string) (GameRule, bool) {
    rule, exists := w.gameRules[name]
    return rule, exists
}

func (w *World) SetGameRule(name string, value interface{}) bool {
    if rule, exists := w.gameRules[name]; exists {
        // Type checking based on existing rule
        switch rule.Type {
        case "bool":
            if _, ok := value.(bool); ok {
                rule.Value = value
                w.gameRules[name] = rule
                return true
            }
        case "int":
            if _, ok := value.(int); ok {
                rule.Value = value
                w.gameRules[name] = rule
                return true
            }
        case "float":
            if _, ok := value.(float64); ok {
                rule.Value = value
                w.gameRules[name] = rule
                return true
            }
        }
    }
    return false
}

func (w *World) getGameRuleBool(name string) bool {
    if rule, exists := w.gameRules[name]; exists && rule.Type == "bool" {
        return rule.Value.(bool)
    }
    return false
}

// Utility methods
func (w *World) findGroundLevel(x, z int32) int32 {
    chunkX := x >> 4
    chunkZ := z >> 4
    localX := x & 0xF
    localZ := z & 0xF

    chunk := w.GetChunk(chunkX, chunkZ)

    // Find highest non-air block
    for y := ChunkHeight - 1; y >= 0; y-- {
        block := chunk.GetBlock(localX, int32(y), localZ)
        if block.ID != 0 {
            return int32(y)
        }
    }

    return -1
}

func (w *World) BroadcastMessage(message string) {
    players := w.GetOnlinePlayers()
    for _, player := range players {
        // TODO: Send chat message to player
        log.Printf("[CHAT to %s] %s", player.Username, message)
    }
}

func (w *World) Stop() {
    w.running = false

    // Save all chunks
    w.chunksMutex.RLock()
    for coord := range w.chunks {
        w.SaveChunk(coord.X, coord.Z)
    }
    w.chunksMutex.RUnlock()

    log.Printf("World '%s' stopped", w.name)
}