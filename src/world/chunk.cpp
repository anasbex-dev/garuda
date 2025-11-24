#include "world/chunk.h"
#include <iostream>
#include <cstring>

Chunk::Chunk(int x, int z) : chunkX(x), chunkZ(z), generated(false) {
    blockData.resize(16 * 16 * 256); // 16x16x256 blocks
    biomeData.resize(16 * 16); // 16x16 biomes
    
    // Initialize dengan air/void
    std::fill(blockData.begin(), blockData.end(), 0);
    std::fill(biomeData.begin(), biomeData.end(), 1); // Plains biome
}

void Chunk::generateTerrain() {
    std::cout << "Generating terrain for chunk [" << chunkX << ", " << chunkZ << "]" << std::endl;
    
    // Simple terrain generation
    for (int x = 0; x < 16; x++) {
        for (int z = 0; z < 16; z++) {
            // Bedrock layer
            setBlock(x, 0, z, 7); // Bedrock
            
            // Stone layers
            for (int y = 1; y < 50; y++) {
                setBlock(x, y, z, 1); // Stone
            }
            
            // Grass layer
            setBlock(x, 50, z, 2); // Grass
        }
    }
    
    generated = true;
}

uint8_t Chunk::getBlock(int x, int y, int z) const {
    if (x < 0 || x >= 16 || y < 0 || y >= 256 || z < 0 || z >= 16) {
        return 0; // Air/void jika out of bounds
    }
    return blockData[(y * 16 + z) * 16 + x];
}

void Chunk::setBlock(int x, int y, int z, uint8_t blockId) {
    if (x >= 0 && x < 16 && y >= 0 && y < 256 && z >= 0 && z < 16) {
        blockData[(y * 16 + z) * 16 + x] = blockId;
    }
}

std::vector<uint8_t> Chunk::serialize() const {
    // Simple serialization untuk chunk data
    std::vector<uint8_t> data;
    
    // Chunk coordinates
    data.push_back((chunkX >> 24) & 0xFF);
    data.push_back((chunkX >> 16) & 0xFF);
    data.push_back((chunkX >> 8) & 0xFF);
    data.push_back(chunkX & 0xFF);
    
    data.push_back((chunkZ >> 24) & 0xFF);
    data.push_back((chunkZ >> 16) & 0xFF);
    data.push_back((chunkZ >> 8) & 0xFF);
    data.push_back(chunkZ & 0xFF);
    
    // Block data
    data.insert(data.end(), blockData.begin(), blockData.end());
    
    // Biome data
    data.insert(data.end(), biomeData.begin(), biomeData.end());
    
    return data;
}

void Chunk::deserialize(const std::vector<uint8_t>& data) {
    // Deserialize chunk data
    if (data.size() >= (16*16*256 + 16*16 + 8)) {
        // Skip coordinates for now
        size_t offset = 8;
        
        // Copy block data
        std::copy(data.begin() + offset, data.begin() + offset + 16*16*256, blockData.begin());
        offset += 16*16*256;
        
        // Copy biome data
        std::copy(data.begin() + offset, data.begin() + offset + 16*16, biomeData.begin());
        
        generated = true;
    }
}