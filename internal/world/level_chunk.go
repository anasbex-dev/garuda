package minecraft

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"garuda/pkg/utils"
	"garuda/internal/world"
	"io"
)

// LevelChunkPacket untuk mengirim chunk data ke client
type LevelChunkPacket struct {
	ChunkX        int32
	ChunkZ        int32
	SubChunkCount uint32
	CacheEnabled  bool
	Payload       []byte
}

func (p *LevelChunkPacket) ID() byte { return 0x3A } // ID untuk LevelChunkPacket di Bedrock

func (p *LevelChunkPacket) Encode() ([]byte, error) {
	stream := utils.NewBinaryStream(nil)
	
	// Header
	stream.WriteUint32(uint32(p.ChunkX))
	stream.WriteUint32(uint32(p.ChunkZ))
	stream.WriteUint32(p.SubChunkCount)
	stream.WriteByte(0) // Highest subchunk version
	
	if p.CacheEnabled {
		stream.WriteByte(1)
	} else {
		stream.WriteByte(0)
	}
	
	// Payload length
	stream.WriteUint32(uint32(len(p.Payload)))
	stream.WriteBytes(p.Payload)
	
	return stream.Bytes(), nil
}

func (p *LevelChunkPacket) Decode(data []byte) error {
	stream := utils.NewBinaryStream(data)
	
	p.ChunkX = int32(stream.ReadUint32())
	p.ChunkZ = int32(stream.ReadUint32())
	p.SubChunkCount = stream.ReadUint32()
	_ = stream.ReadByte() // Skip highest subchunk version
	
	cacheEnabled := stream.ReadByte()
	p.CacheEnabled = (cacheEnabled == 1)
	
	payloadLen := stream.ReadUint32()
	p.Payload = stream.ReadBytes(int(payloadLen))
	
	return nil
}

// ChunkEncoder menangani encoding chunk untuk network
type ChunkEncoder struct {
	compressionLevel int
}

func NewChunkEncoder() *ChunkEncoder {
	return &ChunkEncoder{
		compressionLevel: zlib.DefaultCompression,
	}
}

// EncodeChunk mengencode chunk world menjadi payload untuk network
func (e *ChunkEncoder) EncodeChunk(chunk *world.Chunk) ([]byte, error) {
	// Encode chunk sections
	var buffer bytes.Buffer
	
	// Subchunk count (semua 16 section untuk full chunk)
	subChunkCount := byte(world.TotalSections)
	buffer.WriteByte(subChunkCount)
	
	// Encode setiap section yang tidak empty
	sectionsEncoded := 0
	for y := 0; y < world.TotalSections; y++ {
		section := chunk.Sections[y]
		if section != nil && !isSectionEmpty(section) {
			if err := e.encodeSubChunk(&buffer, section, byte(y)); err != nil {
				return nil, err
			}
			sectionsEncoded++
		} else {
			// Empty section - tetap kirim dengan palette air saja
			emptySection := world.NewChunkSection()
			if err := e.encodeSubChunk(&buffer, emptySection, byte(y)); err != nil {
				return nil, err
			}
		}
	}
	
	// Heightmap (256 entries, each 2 bytes)
	heightmap := make([]byte, 512)
	for i, height := range chunk.HeightMap {
		binary.LittleEndian.PutUint16(heightmap[i*2:], uint16(height))
	}
	buffer.Write(heightmap)
	
	// Biomes (256 bytes - 1 byte per column)
	biomeData := make([]byte, 256)
	for i := 0; i < 256; i++ {
		biomeData[i] = 1 // Plains biome default
	}
	buffer.Write(biomeData)
	
	// Border blocks (0 untuk sekarang)
	buffer.WriteByte(0) // No border blocks
	
	// Extra data (none untuk sekarang)
	buffer.WriteByte(0) // No extra data
	
	// Compress payload
	return e.compressPayload(buffer.Bytes())
}

func (e *ChunkEncoder) encodeSubChunk(buffer *bytes.Buffer, section *world.ChunkSection, subChunkIndex byte) error {
	// SubChunk version (8 = dengan palette)
	buffer.WriteByte(8)
	
	// Block count (non-air blocks)
	blockCount := uint16(0)
	for i := 0; i < 4096; i++ {
		if section.Blocks[i] != 0 {
			blockCount++
		}
	}
	binary.Write(buffer, binary.LittleEndian, blockCount)
	
	// Bits per block (berdasarkan palette size)
	bitsPerBlock := byte(0)
	paletteSize := len(section.Palette)
	
	switch {
	case paletteSize <= 1:
		bitsPerBlock = 0
	case paletteSize <= 16:
		bitsPerBlock = 4
	case paletteSize <= 256:
		bitsPerBlock = 8
	default:
		bitsPerBlock = 16 // Direct IDs
	}
	
	buffer.WriteByte(bitsPerBlock)
	
	// Encode blocks berdasarkan bitsPerBlock
	if err := e.encodeBlocks(buffer, section, bitsPerBlock); err != nil {
		return err
	}
	
	// Palette (jika menggunakan palette)
	if bitsPerBlock != 16 {
		binary.Write(buffer, binary.LittleEndian, uint32(len(section.Palette)))
		for _, blockID := range section.Palette {
			binary.Write(buffer, binary.LittleEndian, blockID)
		}
	}
	
	return nil
}

func (e *ChunkEncoder) encodeBlocks(buffer *bytes.Buffer, section *world.ChunkSection, bitsPerBlock byte) error {
	if bitsPerBlock == 0 {
		// Single block type - hanya perlu 1 word
		var blockID uint32
		if len(section.Palette) > 0 {
			blockID = uint32(section.Palette[0])
		}
		binary.Write(buffer, binary.LittleEndian, blockID)
		return nil
	}
	
	// Calculate words needed
	blocksPerWord := 32 / int(bitsPerBlock)
	totalWords := (4096 + blocksPerWord - 1) / blocksPerWord
	
	for wordIndex := 0; wordIndex < totalWords; wordIndex++ {
		var word uint32
		
		for blockInWord := 0; blockInWord < blocksPerWord; blockInWord++ {
			blockIndex := wordIndex*blocksPerWord + blockInWord
			if blockIndex >= 4096 {
				break
			}
			
			var value uint32
			if bitsPerBlock == 16 {
				// Direct block IDs
				value = uint32(section.Blocks[blockIndex])
			} else {
				// Palette indices
				value = uint32(section.Blocks[blockIndex])
			}
			
			word |= value << (uint(blockInWord) * uint(bitsPerBlock))
		}
		
		binary.Write(buffer, binary.LittleEndian, word)
	}
	
	return nil
}

func (e *ChunkEncoder) compressPayload(data []byte) ([]byte, error) {
	var compressed bytes.Buffer
	writer, err := zlib.NewWriterLevel(&compressed, e.compressionLevel)
	if err != nil {
		return nil, err
	}
	
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	
	if err := writer.Close(); err != nil {
		return nil, err
	}
	
	return compressed.Bytes(), nil
}

func isSectionEmpty(section *world.ChunkSection) bool {
	// Check jika section hanya berisi air blocks
	if len(section.Palette) != 1 {
		return false
	}
	
	if section.Palette[0] != 0 { // 0 = air
		return false
	}
	
	// Verify semua blocks adalah air
	for i := 0; i < 4096; i++ {
		if section.Blocks[i] != 0 {
			return false
		}
	}
	
	return true
}

// NetworkChunkManager mengelola chunk sending ke clients
type NetworkChunkManager struct {
	encoder *ChunkEncoder
}

func NewNetworkChunkManager() *NetworkChunkManager {
	return &NetworkChunkManager{
		encoder: NewChunkEncoder(),
	}
}

// CreateLevelChunkPacket membuat packet untuk mengirim chunk ke client
func (m *NetworkChunkManager) CreateLevelChunkPacket(chunk *world.Chunk) (*LevelChunkPacket, error) {
	payload, err := m.encoder.EncodeChunk(chunk)
	if err != nil {
		return nil, err
	}
	
	return &LevelChunkPacket{
		ChunkX:        chunk.X,
		ChunkZ:        chunk.Z,
		SubChunkCount: uint32(world.TotalSections),
		CacheEnabled:  false,
		Payload:       payload,
	}, nil
}

// BatchChunkPackets untuk mengirim multiple chunks
func (m *NetworkChunkManager) BatchChunkPackets(chunks []*world.Chunk) ([]*LevelChunkPacket, error) {
	packets := make([]*LevelChunkPacket, 0, len(chunks))
	
	for _, chunk := range chunks {
		packet, err := m.CreateLevelChunkPacket(chunk)
		if err != nil {
			return nil, err
		}
		packets = append(packets, packet)
	}
	
	return packets, nil
}