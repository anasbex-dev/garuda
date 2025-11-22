package world

import "garuda/pkg/utils"

type Generator struct {
	seed   string
	logger *utils.Logger
}

func NewGenerator(seed string, logger *utils.Logger) *Generator {
	return &Generator{
		seed:   seed,
		logger: logger,
	}
}

func (g *Generator) GenerateChunk(x, z int32) *Chunk {
	chunk := NewChunk(x, z)
	
	// Simple flat world generation
	for blockX := int32(0); blockX < CHUNK_WIDTH; blockX++ {
		for blockZ := int32(0); blockZ < CHUNK_DEPTH; blockZ++ {
			// Bedrock at bottom
			chunk.SetBlock(blockX, 0, blockZ, 7) // Bedrock
			
			// Stone layers
			for y := int32(1); y < 5; y++ {
				chunk.SetBlock(blockX, y, blockZ, 1) // Stone
			}
			
			// Dirt layer
			for y := int32(5); y < 10; y++ {
				chunk.SetBlock(blockX, y, blockZ, 3) // Dirt
			}
			
			// Grass on top
			chunk.SetBlock(blockX, 10, blockZ, 2) // Grass
		}
	}
	
	g.logger.Debug("Generated chunk at %d, %d", x, z)
	return chunk
}