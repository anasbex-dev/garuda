#include <iostream>
#include "network/raknet_server.h"
#include "network/bedrock_protocol.h"

int main() {
    std::cout << "Starting GarudaMC Server..." << std::endl;
    
    RakNetServer server;
    if (!server.start(19132)) {
        std::cerr << "Failed to start server!" << std::endl;
        return 1;
    }
    
    std::cout << "Server running on port 19132" << std::endl;
    
    // Main loop
    while (true) {
        server.handleIncomingPackets();
        // Other processing...
    }
    
    return 0;
}