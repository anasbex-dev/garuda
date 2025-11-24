#pragma once
#include <cstdint>
#include <vector>

class Chunk {
private:
    int chunkX;
    int chunkZ;
    std::vector<uint8_t> blockData; // Storage untuk blocks
    std::vector<uint8_t> biomeData; // Biome information
    bool generated;
    
public:
    Chunk(int x, int z);
    
    void generateTerrain(); // Generate basic terrain
    uint8_t getBlock(int x, int y, int z) const;
    void setBlock(int x, int y, int z, uint8_t blockId);
    
    std::vector<uint8_t> serialize() const; // Untuk dikirim ke client
    void deserialize(const std::vector<uint8_t>& data);
    
    int getX() const { return chunkX; }
    int getZ() const { return chunkZ; }
    bool isGenerated() const { return generated; }
};