#pragma once
#include <vector>
#include "network/raknet_server.h"

class LoginSequence {
public:
    static bool handleClientConnect(Connection& conn, const std::vector<uint8_t>& packetData);
    static void sendEncryptionHandshake(Connection& conn);
    static bool handleEncryptionResponse(Connection& conn, const std::vector<uint8_t>& responseData);
    static void completeLogin(Connection& conn);
};