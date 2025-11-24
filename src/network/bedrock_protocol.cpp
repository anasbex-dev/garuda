#include "network/bedrock_protocol.h"
#include <iostream>

bool BedrockProtocol::validateProtocolVersion(uint32_t version) {
    // Accept all versions for now, or implement specific validation
    return (version >= 660 && version <= 670); // Accept around 1.21.x
}

std::vector<uint8_t> BedrockProtocol::createPlayStatusPacket(PlayStatus status) {
    std::vector<uint8_t> packet;
    
    // Packet ID untuk PlayStatus (0x02)
    packet.push_back(0x02);
    
    // Status (4 bytes little endian)
    uint32_t status_val = static_cast<uint32_t>(status);
    for (int i = 0; i < 4; i++) {
        packet.push_back((status_val >> (i * 8)) & 0xFF);
    }
    
    std::cout << "Created PlayStatus packet: " << static_cast<int>(status_val) << std::endl;
    return packet;
}

std::vector<uint8_t> BedrockProtocol::createDisconnectPacket(const std::string& reason) {
    std::vector<uint8_t> packet;
    
    // Packet ID untuk Disconnect (0x05)
    packet.push_back(0x05);
    
    // Hide disconnect reason for now (optional field)
    packet.push_back(0x00); // No message
    
    return packet;
}

uint32_t BedrockProtocol::getCurrentProtocolVersion() {
    return static_cast<uint32_t>(ProtocolVersion::V1_21_50); // 666
}

std::string BedrockProtocol::versionToString(uint32_t protocol_version) {
    switch (protocol_version) {
        case 662: return "1.21.10";
        case 663: return "1.21.20"; 
        case 664: return "1.21.30";
        case 665: return "1.21.40";
        case 666: return "1.21.50";
        default: return "Unknown";
    }
}