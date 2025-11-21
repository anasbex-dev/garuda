package world

type Generator interface {
    GenerateChunk(x, z int32) *Chunk
}

type FlatGenerator struct {
    layers []BlockLayer
}

type BlockLayer struct {
    Block Block
    Height int
}

func NewFlatGenerator() *FlatGenerator {
    return &FlatGenerator{
        layers: []BlockLayer{
            {Block: Block{ID: 7, Data: 0}, Height: 1},  // Bedrock
            {Block: Block{ID: 3, Data: 0}, Height: 3},  // Dirt
            {Block: Block{ID: 2, Data: 0}, Height: 1},  // Grass
        },
    }
}

func (g *FlatGenerator) GenerateChunk(x, z int32) *Chunk {
    chunk := NewChunk(x, z)
    
    currentY := 0
    for _, layer := range g.layers {
        for i := 0; i < layer.Height; i++ {
            for blockX := int32(0); blockX < ChunkWidth; blockX++ {
                for blockZ := int32(0); blockZ < ChunkDepth; blockZ++ {
                    chunk.SetBlock(blockX, int32(currentY), blockZ, layer.Block)
                }
            }
            currentY++
        }
    }
    
    // Fill remaining with air (already done by default)
    
    return chunk
}