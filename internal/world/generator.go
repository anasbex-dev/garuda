package world

type Generator interface {
    GenerateChunk(x, z int32) *Chunk
}

type FlatGenerator struct {
    World *World
}

func NewFlatGenerator(world *World) *FlatGenerator {
    return &FlatGenerator{
        World: world,
    }
}

func (g *FlatGenerator) GenerateChunk(x, z int32) *Chunk {
    chunk := NewChunk(x, z)
    
    // Generate flat world similar to the basic generation
    for localX := 0; localX < ChunkWidth; localX++ {
        for localZ := 0; localZ < ChunkLength; localZ++ {
            chunk.SetBlock(localX, 0, localZ, Block{ID: 7})  // Bedrock
            chunk.SetBlock(localX, 1, localZ, Block{ID: 1})  // Stone
            chunk.SetBlock(localX, 2, localZ, Block{ID: 1})  // Stone
            chunk.SetBlock(localX, 3, localZ, Block{ID: 3})  // Dirt
            chunk.SetBlock(localX, 4, localZ, Block{ID: 2})  // Grass
        }
    }
    
    return chunk
}