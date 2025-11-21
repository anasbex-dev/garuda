package world

const (
    ChunkWidth = 16
    ChunkHeight = 256
    ChunkDepth = 16
    SubChunkCount = 16
)

type Chunk struct {
    X int32
    Z int32
    SubChunks []*SubChunk
    Biomes    []byte
    HeightMap []int16
}

type SubChunk struct {
    Blocks []byte
    Data   []byte // 4-bit block data
}

type Block struct {
    ID    uint16
    Data  byte
}

func NewChunk(x, z int32) *Chunk {
    chunk := &Chunk{
        X: x,
        Z: z,
        SubChunks: make([]*SubChunk, SubChunkCount),
        Biomes: make([]byte, 256), // 16x16 biome map
        HeightMap: make([]int16, 256), // 16x16 height map
    }
    
    // Initialize subchunks
    for i := 0; i < SubChunkCount; i++ {
        chunk.SubChunks[i] = NewSubChunk()
    }
    
    // Initialize with default biome (plains)
    for i := range chunk.Biomes {
        chunk.Biomes[i] = 1 // Plains biome
    }
    
    return chunk
}

func NewSubChunk() *SubChunk {
    return &SubChunk{
        Blocks: make([]byte, ChunkWidth*ChunkWidth*16), // 16 layers high
        Data:   make([]byte, ChunkWidth*ChunkWidth*8),  // 4-bit data, so half size
    }
}

func (c *Chunk) GetBlock(x, y, z int32) Block {
    if x < 0 || x >= ChunkWidth || z < 0 || z >= ChunkDepth || y < 0 || y >= ChunkHeight {
        return Block{ID: 0} // Air block
    }
    
    subChunkIndex := y / 16
    if subChunkIndex < 0 || subChunkIndex >= SubChunkCount {
        return Block{ID: 0}
    }
    
    subChunk := c.SubChunks[subChunkIndex]
    localY := y % 16
    
    blockIndex := (localY * ChunkWidth * ChunkWidth) + (z * ChunkWidth) + x
    dataIndex := blockIndex / 2
    
    blockID := subChunk.Blocks[blockIndex]
    
    var data byte
    if blockIndex%2 == 0 {
        data = subChunk.Data[dataIndex] & 0x0F
    } else {
        data = (subChunk.Data[dataIndex] >> 4) & 0x0F
    }
    
    return Block{
        ID:   uint16(blockID),
        Data: data,
    }
}

func (c *Chunk) SetBlock(x, y, z int32, block Block) {
    if x < 0 || x >= ChunkWidth || z < 0 || z >= ChunkDepth || y < 0 || y >= ChunkHeight {
        return
    }
    
    subChunkIndex := y / 16
    if subChunkIndex < 0 || subChunkIndex >= SubChunkCount {
        return
    }
    
    subChunk := c.SubChunks[subChunkIndex]
    localY := y % 16
    
    blockIndex := (localY * ChunkWidth * ChunkWidth) + (z * ChunkWidth) + x
    dataIndex := blockIndex / 2
    
    // Set block ID
    subChunk.Blocks[blockIndex] = byte(block.ID)
    
    // Set block data (4-bit)
    if blockIndex%2 == 0 {
        subChunk.Data[dataIndex] = (subChunk.Data[dataIndex] & 0xF0) | (block.Data & 0x0F)
    } else {
        subChunk.Data[dataIndex] = (subChunk.Data[dataIndex] & 0x0F) | ((block.Data & 0x0F) << 4)
    }
    
    // Update height map if this is the highest block at this position
    if block.ID != 0 { // Not air
        surfaceX := x
        surfaceZ := z
        if y > int32(c.HeightMap[surfaceZ*ChunkWidth+surfaceX]) {
            c.HeightMap[surfaceZ*ChunkWidth+surfaceX] = int16(y)
        }
    }
}

func (c *Chunk) ToNetworkBytes() []byte {
    // This will encode the chunk for network transmission
    // Simplified version - real implementation would be more complex
    
    var data []byte
    
    // For each subchunk
    for _, subChunk := range c.SubChunks {
        // Subchunk version (8)
        data = append(data, 8)
        
        // Block IDs (4096 bytes)
        data = append(data, subChunk.Blocks...)
        
        // Block data (2048 bytes for 4-bit data)
        data = append(data, subChunk.Data...)
    }
    
    // Biomes (256 bytes)
    data = append(data, c.Biomes...)
    
    return data
}