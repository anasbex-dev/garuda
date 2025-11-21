package world

import (
    "sync"
    "github.com/anabex-dev/garuda/internal/protocol/minecraft"
)

type World struct {
    name      string
    seed      int64
    chunks    map[ChunkCoord]*Chunk
    chunksMutex sync.RWMutex
    players   map[int64]*Player
    playersMutex sync.RWMutex
    generator Generator
}

type ChunkCoord struct {
    X int32
    Z int32
}

type Player struct {
    EntityID      int64
    RuntimeID     uint64
    Username      string
    Position      minecraft.Vector3
    Rotation      minecraft.Vector2
    GameMode      int32
    ChunkCoord    ChunkCoord
}

func NewWorld(name string, seed int64) *World {
    return &World{
        name:      name,
        seed:      seed,
        chunks:    make(map[ChunkCoord]*Chunk),
        players:   make(map[int64]*Player),
        generator: NewFlatGenerator(),
    }
}

func (w *World) AddPlayer(player *Player) {
    w.playersMutex.Lock()
    defer w.playersMutex.Unlock()
    
    w.players[player.EntityID] = player
    w.sendChunksToPlayer(player)
}

func (w *World) RemovePlayer(entityID int64) {
    w.playersMutex.Lock()
    defer w.playersMutex.Unlock()
    
    delete(w.players, entityID)
}

func (w *World) GetPlayer(entityID int64) *Player {
    w.playersMutex.RLock()
    defer w.playersMutex.RUnlock()
    
    return w.players[entityID]
}

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
    
    return chunk
}

func (w *World) sendChunksToPlayer(player *Player) {
    // Send chunks around player
    centerX := player.ChunkCoord.X
    centerZ := player.ChunkCoord.Z
    radius := int32(2) // 5x5 chunk area
    
    for x := centerX - radius; x <= centerX + radius; x++ {
        for z := centerZ - radius; z <= centerZ + radius; z++ {
            chunk := w.GetChunk(x, z)
            // Send chunk to player (implement later)
            w.sendChunkToPlayer(player, chunk)
        }
    }
}

func (w *World) sendChunkToPlayer(player *Player, chunk *Chunk) {
    // TODO: Implement chunk packet sending
    // This will use LevelChunkPacket
}