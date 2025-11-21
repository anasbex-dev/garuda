package world

import (
    "compress/zlib"
    "encoding/binary"
    "garuda/pkg/utils"
    "io"
    "sync"
)

const (
    ChunkWidth  = 16
    ChunkHeight = 256
    ChunkLength = 16
    TotalBlocks = ChunkWidth * ChunkHeight * ChunkLength
    SectionHeight = 16
    TotalSections = ChunkHeight / SectionHeight
)

// Compact Block representation dengan palette optimization
type Block struct {
    ID    uint16
    Data  uint8
    Light uint8
}

type ChunkSection struct {
    Blocks     [4096]uint16 // 16x16x16 = 4096 blocks, menggunakan palette indices
    Palette    []uint16     // Block palette untuk section ini
    BlockLight [2048]byte   // 4 bits per block = 2048 bytes
    SkyLight   [2048]byte   // 4 bits per block = 2048 bytes
}

type Chunk struct {
    X, Z    int32
    Sections [TotalSections]*ChunkSection
    mutex   sync.RWMutex
    // Cache untuk performance
    biomeData [256]byte
    heightMap [256]int16
}

func NewChunk(x, z int32) *Chunk {
    chunk := &Chunk{
        X: x,
        Z: z,
    }
    
    // Initialize sections
    for i := range chunk.Sections {
        chunk.Sections[i] = NewChunkSection()
    }
    
    // Initialize biome data (plain biome)
    for i := range chunk.biomeData {
        chunk.biomeData[i] = 1 // Plains biome
    }
    
    // Initialize height map
    for i := range chunk.heightMap {
        chunk.heightMap[i] = 64 // Default height
    }
    
    return chunk
}

func NewChunkSection() *ChunkSection {
    section := &ChunkSection{
        Palette: make([]uint16, 1),
    }
    section.Palette[0] = 0 // Air block
    
    return section
}

func (c *Chunk) GetBlock(x, y, z int) Block {
    if x < 0 || x >= ChunkWidth || y < 0 || y >= ChunkHeight || z < 0 || z >= ChunkLength {
        return Block{ID: 0}
    }
    
    sectionIndex := y / SectionHeight
    sectionY := y % SectionHeight
    
    if sectionIndex < 0 || sectionIndex >= TotalSections {
        return Block{ID: 0}
    }
    
    section := c.Sections[sectionIndex]
    if section == nil {
        return Block{ID: 0}
    }
    
    blockIndex := (sectionY * ChunkWidth * ChunkLength) + (z * ChunkWidth) + x
    if blockIndex >= 4096 {
        return Block{ID: 0}
    }
    
    paletteIndex := section.Blocks[blockIndex]
    if int(paletteIndex) >= len(section.Palette) {
        return Block{ID: 0}
    }
    
    blockID := section.Palette[paletteIndex]
    
    // Extract light data
    lightIndex := blockIndex / 2
    lightShift := (blockIndex % 2) * 4
    
    blockLight := (section.BlockLight[lightIndex] >> lightShift) & 0x0F
    skyLight := (section.SkyLight[lightIndex] >> lightShift) & 0x0F
    
    return Block{
        ID:    blockID,
        Data:  0, // Data disimpan terpisah jika needed
        Light: uint8((skyLight << 4) | blockLight),
    }
}

func (c *Chunk) SetBlock(x, y, z int, block Block) {
    if x < 0 || x >= ChunkWidth || y < 0 || y >= ChunkHeight || z < 0 || z >= ChunkLength {
        return
    }
    
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    sectionIndex := y / SectionHeight
    sectionY := y % SectionHeight
    
    if sectionIndex < 0 || sectionIndex >= TotalSections {
        return
    }
    
    section := c.Sections[sectionIndex]
    if section == nil {
        section = NewChunkSection()
        c.Sections[sectionIndex] = section
    }
    
    blockIndex := (sectionY * ChunkWidth * ChunkLength) + (z * ChunkWidth) + x
    if blockIndex >= 4096 {
        return
    }
    
    // Find or add block to palette
    paletteIndex := c.findOrAddToPalette(section, block.ID)
    section.Blocks[blockIndex] = paletteIndex
    
    // Update light data
    lightIndex := blockIndex / 2
    lightShift := (blockIndex % 2) * 4
    
    blockLight := block.Light & 0x0F
    skyLight := (block.Light >> 4) & 0x0F
    
    section.BlockLight[lightIndex] = (section.BlockLight[lightIndex] & ^(0x0F << lightShift)) | (blockLight << lightShift)
    section.SkyLight[lightIndex] = (section.SkyLight[lightIndex] & ^(0x0F << lightShift)) | (skyLight << lightShift)
    
    // Update height map jika needed
    if block.ID != 0 && y > int(c.heightMap[x+z*ChunkWidth]) {
        c.heightMap[x+z*ChunkWidth] = int16(y)
    }
}

func (c *Chunk) findOrAddToPalette(section *ChunkSection, blockID uint16) uint16 {
    // Cari di palette existing
    for i, id := range section.Palette {
        if id == blockID {
            return uint16(i)
        }
    }
    
    // Tambah ke palette
    section.Palette = append(section.Palette, blockID)
    return uint16(len(section.Palette) - 1)
}

func (c *Chunk) GetHighestBlockY(x, z int) int {
    if x < 0 || x >= ChunkWidth || z < 0 || z >= ChunkLength {
        return 0
    }
    
    return int(c.heightMap[x+z*ChunkWidth])
}

func (c *Chunk) Encode() []byte {
    // Optimized chunk encoding dengan compression
    stream := utils.NewBinaryStream(nil)
    
    // Encode each section
    sectionsEncoded := 0
    for i, section := range c.Sections {
        if section != nil && !c.isSectionEmpty(section) {
            c.encodeSection(stream, section, i)
            sectionsEncoded++
        }
    }
    
    // Add biome data
    stream.WriteBytes(c.biomeData[:])
    
    // Compress jika worth it
    if sectionsEncoded > 0 {
        return c.compressData(stream.Bytes())
    }
    
    return stream.Bytes()
}

func (c *Chunk) isSectionEmpty(section *ChunkSection) bool {
    // Check jika section hanya berisi air
    return len(section.Palette) == 1 && section.Palette[0] == 0
}

func (c *Chunk) encodeSection(stream *utils.BinaryStream, section *ChunkSection, sectionIndex int) {
    // Storage version
    stream.WriteByte(8) // Version 8
    
    // Block count (non-air blocks)
    blockCount := uint16(0)
    for i := 0; i < 4096; i++ {
        if section.Blocks[i] != 0 {
            blockCount++
        }
    }
    stream.WriteUint16(blockCount)
    
    // Palette
    stream.WriteByte(byte(len(section.Palette))) // Bits per block
    for _, blockID := range section.Palette {
        stream.WriteUint16(blockID)
    }
    
    // Block data
    for i := 0; i < 4096; i++ {
        stream.WriteUint16(section.Blocks[i])
    }
    
    // Light data
    stream.WriteBytes(section.BlockLight[:])
    stream.WriteBytes(section.SkyLight[:])
}

func (c *Chunk) compressData(data []byte) []byte {
    // Simple compression - bisa diimprove dengan zlib nanti
    compressed := make([]byte, 0, len(data))
    
    // Run-length encoding sederhana
    var currentByte byte
    var count byte = 1
    
    for i, b := range data {
        if i == 0 {
            currentByte = b
            continue
        }
        
        if b == currentByte && count < 255 {
            count++
        } else {
            compressed = append(compressed, count, currentByte)
            currentByte = b
            count = 1
        }
    }
    
    // Add last run
    compressed = append(compressed, count, currentByte)
    
    // Jika compression tidak efektif, return original
    if len(compressed) >= len(data) {
        return data
    }
    
    return compressed
}

// Memory optimization methods
func (c *Chunk) GetMemoryUsage() int {
    size := binary.Size(c.X) + binary.Size(c.Z) + 
            binary.Size(c.biomeData) + binary.Size(c.heightMap)
    
    for _, section := range c.Sections {
        if section != nil {
            size += binary.Size(section.Blocks) + 
                   binary.Size(section.BlockLight) + 
                   binary.Size(section.SkyLight) + 
                   binary.Size(section.Palette)
        }
    }
    
    return size
}

func (c *Chunk) OptimizeMemory() {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    for i, section := range c.Sections {
        if section != nil && c.isSectionEmpty(section) {
            c.Sections[i] = nil // Free empty sections
        }
    }
}