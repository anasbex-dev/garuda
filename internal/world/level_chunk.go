package world

import (
    "garuda/minecraft"
)

type LevelChunkPacket struct {
    ChunkX        int32
    ChunkZ        int32
    SubChunkCount uint32
    Data          []byte
}

func (p *LevelChunkPacket) ID() byte { return minecraft.ID_LEVEL_CHUNK }

func (p *LevelChunkPacket) Encode() ([]byte, error) {
    encoder := minecraft.NewEncoder()
    
    encoder.WriteVarInt(p.ChunkX)
    encoder.WriteVarInt(p.ChunkZ)
    encoder.WriteVarInt(int32(p.SubChunkCount))
    encoder.WriteBool(false)
    encoder.WriteUShort(0)
    
    encoder.WriteVarInt(int32(len(p.Data)))
    encoder.stream.WriteBytes(p.Data)
    
    encoder.WriteVarInt(0)
    
    return encoder.Bytes(), nil
}

func (p *LevelChunkPacket) Decode(data []byte) error {
    decoder := minecraft.NewDecoder(data)
    
    p.ChunkX = decoder.ReadVarInt()
    p.ChunkZ = decoder.ReadVarInt()
    p.SubChunkCount = uint32(decoder.ReadVarInt())
    _ = decoder.ReadBool()
    
    blobCount := decoder.ReadVarInt()
    for i := int32(0); i < blobCount; i++ {
        _ = decoder.ReadULong()
    }
    
    dataLength := decoder.ReadVarInt()
    p.Data = decoder.ReadByteArray()
    
    _ = decoder.ReadVarInt()
    
    return nil
}