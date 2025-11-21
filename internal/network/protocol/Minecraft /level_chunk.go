package minecraft

const IDLevelChunk = 0x3a

type LevelChunkPacket struct {
    ChunkX        int32
    ChunkZ        int32
    SubChunkCount uint32
    Data          []byte
    CacheEnabled  bool
    BlobHashes    []uint64
}

func EncodeLevelChunkPacket(packet *LevelChunkPacket) ([]byte, error) {
    encoder := NewPacketEncoder()
    
    // Packet ID
    if err := encoder.WriteByte(IDLevelChunk); err != nil {
        return nil, err
    }
    
    // Chunk coordinates
    if err := encoder.WriteInt32(packet.ChunkX); err != nil {
        return nil, err
    }
    if err := encoder.WriteInt32(packet.ChunkZ); err != nil {
        return nil, err
    }
    
    // Subchunk count
    if err := encoder.WriteVarInt(int32(packet.SubChunkCount)); err != nil {
        return nil, err
    }
    
    // Cache enabled
    if err := encoder.WriteBool(packet.CacheEnabled); err != nil {
        return nil, err
    }
    
    if packet.CacheEnabled {
        // Blob hashes count
        if err := encoder.WriteUint32(uint32(len(packet.BlobHashes))); err != nil {
            return nil, err
        }
        
        // Blob hashes
        for _, hash := range packet.BlobHashes {
            if err := encoder.WriteUint64(hash); err != nil {
                return nil, err
            }
        }
    }
    
    // Chunk data
    if err := encoder.WriteUint32(uint32(len(packet.Data))); err != nil {
        return nil, err
    }
    if err := encoder.WriteBytes(packet.Data); err != nil {
        return nil, err
    }
    
    return encoder.Bytes(), nil
}