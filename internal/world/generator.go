package world

import (
    "math"
)

type Generator interface {
    GenerateChunk(x, z int32) *Chunk
    GetName() string
}

type FlatGenerator struct {
    layers []BlockLayer
}

type BlockLayer struct {
    Block  Block
    Height int
    Name   string
}

func NewFlatGenerator() *FlatGenerator {
    return &FlatGenerator{
        layers: []BlockLayer{
            {Block: Block{ID: 7, Data: 0}, Height: 1, Name: "Bedrock"},
            {Block: Block{ID: 3, Data: 0}, Height: 3, Name: "Dirt"},
            {Block: Block{ID: 2, Data: 0}, Height: 1, Name: "Grass"},
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
    
    // Add some random flowers and grass for variety
    for i := 0; i < 10; i++ {
        blockX := int32(math.Mod(float64(x*16+i*13), 16))
        blockZ := int32(math.Mod(float64(z*16+i*7), 16))
        
        if blockX < 0 { blockX += 16 }
        if blockZ < 0 { blockZ += 16 }
        
        // Place flower or tall grass on grass blocks
        if chunk.GetBlock(blockX, int32(currentY-1), blockZ).ID == 2 {
            if i%3 == 0 {
                chunk.SetBlock(blockX, int32(currentY), blockZ, Block{ID: 37, Data: 0}) // Dandelion
            } else if i%3 == 1 {
                chunk.SetBlock(blockX, int32(currentY), blockZ, Block{ID: 38, Data: 0}) // Poppy
            } else {
                chunk.SetBlock(blockX, int32(currentY), blockZ, Block{ID: 31, Data: 1}) // Tall grass
            }
        }
    }
    
    return chunk
}

func (g *FlatGenerator) GetName() string {
    return "flat"
}