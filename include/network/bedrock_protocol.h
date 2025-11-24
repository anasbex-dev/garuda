#pragma once
#include <cstdint>
#include <vector>
#include <string>

enum class PlayStatus {
    LOGIN_SUCCESS = 0,
    FAILED_CLIENT = 1,
    FAILED_SERVER = 2,
    PLAYER_SPAWN = 3
};

enum class ProtocolVersion {
    V1_21_10 = 662,
    V1_21_20 = 663,
    V1_21_30 = 664,
    V1_21_40 = 665,
    V1_21_50 = 666
};

class BedrockProtocol {
public:
    static bool validateProtocolVersion(uint32_t version);
    static std::vector<uint8_t> createPlayStatusPacket(PlayStatus status);
    static std::vector<uint8_t> createDisconnectPacket(const std::string& reason);
    static uint32_t getCurrentProtocolVersion();
    static std::string versionToString(uint32_t protocol_version);
};