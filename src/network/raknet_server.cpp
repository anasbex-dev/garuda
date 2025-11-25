#include "network/raknet_server.h"
#include "network/bedrock_protocol.h"
#include <arpa/inet.h>
#include <sys/socket.h>
#include <unistd.h>
#include <iostream>
#include <cstring>
#include <algorithm>
#include <chrono>
#include <iomanip>
#include <sstream>

// ==================== IMPROVED RAKNET HEADER HANDLING ====================

struct RakNetHeader {
    uint8_t flags;
    uint16_t sequence;
    uint32_t reliable_index;
    uint32_t ordering_index;
    uint8_t ordering_channel;
    
    static RakNetHeader parse(const uint8_t* buffer, size_t length) {
        RakNetHeader header = {};
        if (length < 3) return header;
        
        header.flags = buffer[0];
        int position = 1;
        
        // Sequence number (always present for connected packets)
        if (length >= 3) {
            header.sequence = (buffer[position + 1] << 8) | buffer[position];
            position += 2;
        }
        
        // Reliability handling
        uint8_t reliability = (header.flags >> 5) & 0x07;
        
        // Reliable packet has reliable index
        if (reliability == 2 || reliability == 3 || reliability == 4 || reliability == 6 || reliability == 7) {
            if (length >= position + 3) {
                // Read 3-byte little endian reliable index
                header.reliable_index = (buffer[position + 2] << 16) | (buffer[position + 1] << 8) | buffer[position];
                position += 3;
            }
        }
        
        // Ordered packet has ordering info
        if (reliability == 1 || reliability == 3 || reliability == 4) {
            if (length >= position + 4) {
                // Read 3-byte little endian ordering index + 1 byte channel
                header.ordering_index = (buffer[position + 2] << 16) | (buffer[position + 1] << 8) | buffer[position];
                header.ordering_channel = buffer[position + 3];
                position += 4;
            }
        }
        
        return header;
    }
};

// ==================== IMPLEMENTASI FUNCTION YANG MISSING ====================

void RakNetServer::logHex(const uint8_t* data, size_t length, const std::string& description) {
    if (!description.empty()) {
        std::cout << description << " (" << length << " bytes): ";
    }
    
    for (size_t i = 0; i < length; ++i) {
        std::cout << std::hex << std::setw(2) << std::setfill('0') 
                  << static_cast<int>(data[i]) << " ";
    }
    std::cout << std::dec << std::endl;
}

// ==================== IMPROVED CONNECTION MANAGEMENT ====================

RakNetServer::RakNetServer() : socket_(-1), running_(false) {
    // Initialize encryption handler
}

void RakNetServer::startConnectionTimer(Connection& conn) {
    conn.last_activity = std::chrono::steady_clock::now();
}

bool RakNetServer::isConnectionTimedOut(const Connection& conn) {
    auto now = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(now - conn.last_activity);
    return elapsed.count() > 30; // 30 second timeout
}

void RakNetServer::cleanupStaleConnections() {
    auto it = connections_.begin();
    while (it != connections_.end()) {
        if (isConnectionTimedOut(*it)) {
            std::cout << "Cleaning up stale connection: " << it->address << ":" << it->port << std::endl;
            it = connections_.erase(it);
        } else {
            ++it;
        }
    }
}

// ==================== IMPROVED PACKET SENDING ====================

std::vector<uint8_t> RakNetServer::addRakNetHeader(const std::vector<uint8_t>& payload, uint16_t sequence, uint8_t reliability) {
    std::vector<uint8_t> packet;
    
    // Flags: Connected (0x80) + Reliability
    packet.push_back(0x80 | ((reliability & 0x07) << 5));
    
    // Sequence number (little endian)
    packet.push_back(sequence & 0xFF);
    packet.push_back((sequence >> 8) & 0xFF);
    
    // For reliable packets, add reliable index (3 bytes little endian)
    if (reliability == 2 || reliability == 3 || reliability == 4 || reliability == 6 || reliability == 7) {
        static uint32_t reliable_counter = 0;
        reliable_counter++;
        packet.push_back(reliable_counter & 0xFF);
        packet.push_back((reliable_counter >> 8) & 0xFF);
        packet.push_back((reliable_counter >> 16) & 0xFF);
    }
    
    // Payload
    packet.insert(packet.end(), payload.begin(), payload.end());
    
    return packet;
}

void RakNetServer::sendReliablePacket(Connection& conn, const std::vector<uint8_t>& payload) {
    uint16_t sequence = conn.sequence_out++;
    auto packet = addRakNetHeader(payload, sequence, 0x04); // Reliable unordered
    
    // Store in sent queue for potential resend
    Connection::SentPacket sent_packet;
    sent_packet.data = packet;
    sent_packet.timestamp = std::chrono::steady_clock::now();
    sent_packet.sequence = sequence;
    
    conn.sent_packets[sequence] = sent_packet;
    
    sendRawPacket(conn, packet);
    
    std::cout << "Sent reliable packet seq=" << sequence << " size=" << packet.size() << std::endl;
}

void RakNetServer::sendRawPacket(const Connection& conn, const std::vector<uint8_t>& data) {
    sockaddr_in client_addr;
    client_addr.sin_family = AF_INET;
    client_addr.sin_port = htons(conn.port);
    inet_pton(AF_INET, conn.address.c_str(), &client_addr.sin_addr);
    
    sendto(socket_, data.data(), data.size(), 0,
           (sockaddr*)&client_addr, sizeof(client_addr));
    
    std::cout << "Sent raw packet to " << conn.address << ":" << conn.port 
              << " (size: " << data.size() << " bytes)" << std::endl;
    logHex(data.data(), std::min(data.size(), size_t(16)), "Sent: ");
}

// ==================== IMPROVED PACKET HANDLING ====================

void RakNetServer::handleIncomingPackets() {
    if (!running_) return;

    sockaddr_in client_addr{};
    socklen_t addr_len = sizeof(client_addr);
    uint8_t buffer[4096];

    // Non-blocking receive
    ssize_t received = recvfrom(
        socket_,
        buffer,
        sizeof(buffer),
        MSG_DONTWAIT,
        (sockaddr*)&client_addr,
        &addr_len
    );

    if (received <= 0) {
        return;
    }

    uint8_t packet_id = buffer[0];

    std::cout << "Received packet: 0x" << std::hex << (int)packet_id
              << " from " << inet_ntoa(client_addr.sin_addr)
              << ":" << ntohs(client_addr.sin_port)
              << " (" << std::dec << received << " bytes)\n";

    logHex(buffer, std::min(received, ssize_t(16)), "Data: ");

    // Track / create connection
    auto conn = findOrCreateConnection(client_addr);
    startConnectionTimer(*conn);

    // -------------------------
    //   UNCONNECTED RAKNET
    // -------------------------
    switch (packet_id) {
        case 0x01: // UNCONNECTED_PING
            handleUnconnectedPing(client_addr, buffer, received);
            return;

        case 0x05: // OPEN_CONNECTION_REQUEST_1
            handleOpenConnectionRequest1(client_addr, buffer, received);
            return;

        case 0x07: // OPEN_CONNECTION_REQUEST_2
            handleOpenConnectionRequest2(client_addr, buffer, received);
            return;

        case 0x09: // CONNECTION_REQUEST
            handleConnectionRequest(client_addr, buffer, received);
            return;

        case 0x13: // NEW_INCOMING_CONNECTION
            handleNewIncomingConnection(client_addr, buffer, received);
            return;

        case 0x15: // DISCONNECTION_NOTIFICATION
            handleDisconnection(client_addr, buffer, received);
            return;

        case 0xC0: // ACK
            handleAckPacket(client_addr, buffer, received);
            return;

        case 0xA0: // NACK
            handleNackPacket(client_addr, buffer, received);
            return;

        case 0x84: // CONNECTED_PING (RakNet connected layer)
            handleConnectedPing(client_addr, buffer, received);
            return;

        case 0x90: // CLIENT_CONNECT (BEDROCK PACKET, NOT RAKNET)
            std::cout << "=== CLIENT_CONNECT PACKET ===\n";
            handleClientConnect(client_addr, buffer, received);
            return;
    }

    // -------------------------
    //    CONNECTED PACKETS
    // -------------------------
    // Bit 7 harus 1 => paket RakNet yang sudah "connected" (0b1xxxxxxx)
    if (packet_id & 0x80) {
        handleConnectedPacket(client_addr, buffer, received);
        return;
    }

    // -------------------------
    //       UNKNOWN
    // -------------------------
    std::cout << "Unknown RakNet/Bedrock packet: 0x" 
              << std::hex << (int)packet_id << std::dec << "\n";

    handleResends();
    cleanupStaleConnections();
}

void RakNetServer::handleConnectedPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    // Jika ini CONNECTED_PING (0x84) atau CONNECTED_PONG (0x85), pindah ke handler khusus
    if (buffer[0] == 0x84) {
        handleConnectedPing(client_addr, buffer, length);
        return;
    }
    if (buffer[0] == 0x85) {
        handleConnectedPong(client_addr, buffer, length);
        return;
    }

    auto conn = findOrCreateConnection(client_addr);
    RakNetHeader header = RakNetHeader::parse(buffer, length);

    std::cout << "[Connected] Flags=0x" << std::hex << (int)header.flags
              << " Seq=" << std::dec << header.sequence << std::endl;

    // Wajib ACK setiap sequence client
    sendAckPacket(client_addr, header.sequence);

    //
    // --- Reliability Parsing ---
    //
    uint8_t reliability = (header.flags >> 5) & 0x07;
    size_t offset = 3; // minimal header size

    // Jika reliable → add reliableIndex (2 bytes) + messageIndex (1 byte)
    if (reliability == 2 || reliability == 3 || reliability == 4 || reliability == 6 || reliability == 7) {
        offset += 3;
    }

    // Jika ordered/sequenced → add ordering channel info
    if (reliability == 1 || reliability == 3 || reliability == 4) {
        offset += 4;
    }

    if (offset >= length) return;

    uint8_t* payload = buffer + offset;
    size_t payload_len = length - offset;

    if (payload_len == 0) return;

    uint8_t pk = payload[0];

    std::cout << "[Bedrock] Packet ID: 0x" << std::hex << (int)pk << std::dec << std::endl;

    switch (pk) {
        case 0x01: // LOGIN
            handleBedrockLogin(*conn, payload, payload_len);
            break;

        case 0x90: // CLIENT_CONNECT
            std::cout << "[Bedrock] CLIENT_CONNECT" << std::endl;
            handleClientConnect(client_addr, payload, payload_len);
            break;

        default:
            std::cout << "[Bedrock] Unknown: 0x" << std::hex << (int)pk << std::dec << std::endl;
            logHex(payload, std::min(payload_len, (size_t)16), "Payload: ");
            break;
    }
}

void RakNetServer::handleBedrockLogin(Connection& conn, uint8_t* buffer, size_t length) {
    std::cout << "=== HANDLING BEDROCK LOGIN 1.21.50 ===" << std::endl;
    
    // Parse login packet header
    if (length < 50) {
        std::cout << "Login packet too short" << std::endl;
        return;
    }
    
    // Minecraft 1.21.50 login packet structure:
    // - Protocol version (4 bytes)
    // - Login data length (varint) 
    // - Login data (chain of NBT data dengan JWT token, identity data, etc.)
    
    // Untuk testing, kita extract username sederhana
    // Di real implementation, perlu parse NBT data
    
    // Cari username pattern dalam login data
    std::string username = "Player";
    for (size_t i = 10; i < length - 5; i++) {
        if (buffer[i] == 'p' && buffer[i+1] == 'l' && buffer[i+2] == 't' && buffer[i+3] == 'N') {
            // Potentially found platform data, extract username
            if (i + 20 < length) {
                username = "User_" + std::to_string(conn.guid);
                break;
            }
        }
    }
    
    conn.username = username;
    
    std::cout << "Player logging in: " << username << std::endl;
    std::cout << "Login data size: " << length << " bytes" << std::endl;
    
    // Untuk Minecraft 1.21.50, kita perlu encryption handshake dulu
    if (!conn.encrypted) {
        std::cout << "Client not encrypted, initiating handshake..." << std::endl;
        sendEncryptionRequest(conn);
    } else {
        std::cout << "Client already encrypted, continuing login..." << std::endl;
        sendLoginSuccess(conn);
    }
}

void RakNetServer::sendLoginSequence(Connection& conn) {
    std::cout << "Starting login sequence for " << conn.address << ":" << conn.port << std::endl;
    
    // 1. Send Login Success (reliable)
    auto login_success = BedrockProtocol::createPlayStatusPacket(PlayStatus::LOGIN_SUCCESS);
    sendReliablePacket(conn, login_success);
    
    // 2. Send Resource Pack Stack (reliable)  
    auto resource_stack = createResourcePackStack();
    sendReliablePacket(conn, resource_stack);
    
    // 3. Send Start Game (reliable)
    auto start_game = createStartGamePacket();
    sendReliablePacket(conn, start_game);
    
    std::cout << "Login sequence sent (waiting for ACKs)..." << std::endl;
}

// ==================== MODERN MINECRAFT AUTHENTICATION ====================

void RakNetServer::sendLoginSuccess(Connection& conn) {
    std::cout << "Sending LOGIN_SUCCESS after encryption handshake..." << std::endl;
    
    // Create PlayStatus packet dengan LOGIN_SUCCESS
    std::vector<uint8_t> play_status;
    play_status.push_back(0x02); // PlayStatus packet ID
    
    // Status: LOGIN_SUCCESS (0)
    play_status.push_back(0x00);
    play_status.push_back(0x00);
    play_status.push_back(0x00);
    play_status.push_back(0x00);
    
    sendReliablePacket(conn, play_status);
    std::cout << "Sent LOGIN_SUCCESS, continuing login sequence..." << std::endl;
    
    // Lanjutkan dengan login sequence normal
    sendLoginSequence(conn);
}

void RakNetServer::handleModernLogin(Connection& conn, uint8_t* buffer, size_t length) {
    std::cout << "=== MODERN LOGIN HANDLER ===" << std::endl;
    
    // Parse modern login packet (contains JWT, identity data, etc.)
    // For now, we'll use simplified authentication
    conn.authenticated = true;
    
    // Send encryption request if needed
    if (!conn.encrypted) {
        sendEncryptionRequest(conn);
    } else {
        completeModernLogin(conn);
    }
}

void RakNetServer::handleEncryptionHandshake(Connection& conn, uint8_t* buffer, size_t length) {
    std::cout << "=== HANDLING CLIENT ENCRYPTION HANDSHAKE ===" << std::endl;
    
    if (length < 2) {
        std::cout << "Invalid handshake packet length" << std::endl;
        return;
    }
    
    // Parse the handshake response
    // Client mengirim ClientToServerHandshake (0x04) dengan encrypted payload
    
    logHex(buffer, std::min(length, (size_t)32), "Handshake response: ");
    
    // TRUE/FALSE
    conn.encrypted = true;
    
    std::cout << "Encryption handshake SUCCESS for " << conn.address << std::endl;
    
    sendLoginSuccess(conn);
}

void RakNetServer::sendEncryptionRequest(Connection& conn) {
    std::cout << "Sending encryption request to " << conn.address << ":" << conn.port << std::endl;
    
    // ===== TEST 1: Empty Handshake =====
    std::vector<uint8_t> handshake_packet;
    handshake_packet.push_back(0x03); // Packet ID ServerToClientHandshake
    
    handshake_packet.push_back(0x00); // Zero-length token
    
    std::cout << "TEST 1: Sending EMPTY handshake" << std::endl;
    sendReliablePacket(conn, handshake_packet);
    logHex(handshake_packet.data(), handshake_packet.size(), "Empty handshake: ");
    
    usleep(100000);
    
    // ===== TEST 2: Simple Token =====
    handshake_packet.clear();
    handshake_packet.push_back(0x03);
    
    std::string simple_token = "test";
    handshake_packet.push_back(simple_token.length()); // Length
    handshake_packet.insert(handshake_packet.end(), simple_token.begin(), simple_token.end());
    
    std::cout << "TEST 2: Sending SIMPLE handshake" << std::endl;
    sendReliablePacket(conn, handshake_packet);
    logHex(handshake_packet.data(), handshake_packet.size(), "Simple handshake: ");
    
    // ===== TEST 3: Use EncryptionHandler =====
    handshake_packet.clear();
    handshake_packet.push_back(0x03);
    
    auto handshake_data = encryption_handler_.generateHandshakeToken();
    handshake_packet.push_back(handshake_data.size()); // Length
    handshake_packet.insert(handshake_packet.end(), handshake_data.begin(), handshake_data.end());
    
    std::cout << "TEST 3: Sending ENCRYPTION HANDLER handshake" << std::endl;
    sendReliablePacket(conn, handshake_packet);
    logHex(handshake_packet.data(), handshake_packet.size(), "Encryption handler handshake: ");
    
    std::cout << "Sent 3 different handshake formats, waiting for response..." << std::endl;
}


void RakNetServer::completeModernLogin(Connection& conn) {
    std::cout << "Completing modern login for " << conn.username << std::endl;
    
    // Mark as authenticated and send game data
    conn.authenticated = true;
    
    // Send the normal login sequence
    sendLoginSequence(conn);
}

// ==================== IMPROVED ACK/NACK HANDLING ====================

void RakNetServer::handleAckPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    if (length < 4) return;
    
    // Parse ACK sequence numbers (simple implementation)
    uint16_t sequence = (buffer[2] << 8) | buffer[1];
    
    auto conn = findOrCreateConnection(client_addr);
    conn->sent_packets.erase(sequence); // Remove from resend queue
    
    std::cout << "ACK received for sequence " << sequence 
              << " from " << inet_ntoa(client_addr.sin_addr) << std::endl;
}

void RakNetServer::handleNackPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    if (length < 4) return;
    
    uint16_t sequence = (buffer[2] << 8) | buffer[1];
    
    auto conn = findOrCreateConnection(client_addr);
    auto it = conn->sent_packets.find(sequence);
    if (it != conn->sent_packets.end()) {
        std::cout << "NACK received for sequence " << sequence << ", resending..." << std::endl;
        sendRawPacket(*conn, it->second.data);
    }
}

void RakNetServer::handleResends() {
    auto now = std::chrono::steady_clock::now();
    
    for (auto& conn : connections_) {
        auto it = conn.sent_packets.begin();
        while (it != conn.sent_packets.end()) {
            auto elapsed = std::chrono::duration_cast<std::chrono::milliseconds>(now - it->second.timestamp);
            if (elapsed.count() > 1000) { // 1 second timeout
                std::cout << "Resending packet seq=" << it->first << " to " << conn.address << std::endl;
                sendRawPacket(conn, it->second.data);
                it->second.timestamp = now;
            }
            ++it;
        }
    }
}

// ==================== IMPROVED CONNECTION HANDLING ====================

void RakNetServer::handleConnectionRequest(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    std::cout << "=== HANDLING CONNECTION_REQUEST ===" << std::endl;
    
    uint8_t response[21] = {
        0x10, // CONNECTION_REQUEST_ACCEPTED
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // Server GUID
        0x00, 0x00, // Ping time
        0x00 // Use security
    };
    
    sendto(socket_, response, sizeof(response), 0,
           (sockaddr*)&client_addr, sizeof(client_addr));
    
    std::cout << "Sent CONNECTION_REQUEST_ACCEPTED" << std::endl;
    
    // Send NEW_INCOMING_CONNECTION after a short delay
    usleep(50000); // 50ms delay
    sendNewIncomingConnection(client_addr);
}

void RakNetServer::handleNewIncomingConnection(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    std::cout << "Handling NEW_INCOMING_CONNECTION" << std::endl;
    
    auto conn = findOrCreateConnection(client_addr);
    conn->connected = true;
    conn->sequence_in = 0;
    conn->sequence_out = 1;
    
    std::cout << "Client fully connected: " << conn->address << ":" << conn->port << std::endl;
    
    if (!conn->encrypted) {
        std::cout << "=== INITIATING ENCRYPTION HANDSHAKE (BACKUP) ===" << std::endl;
        sendEncryptionRequest(*conn);
    }
}

void RakNetServer::handleDisconnection(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    std::cout << "Client disconnected: " << inet_ntoa(client_addr.sin_addr) 
              << ":" << ntohs(client_addr.sin_port) << std::endl;
    
    // Remove from connections
    auto it = std::find_if(connections_.begin(), connections_.end(),
        [&](const Connection& conn) {
            return conn.address == inet_ntoa(client_addr.sin_addr) && 
                   conn.port == ntohs(client_addr.sin_port);
        });
    
    if (it != connections_.end()) {
        connections_.erase(it);
    }
}

// ==================== IMPROVED PING/PONG ====================

/*void RakNetServer::handleConnectedPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    // Jika ini CONNECTED_PING (0x84) atau CONNECTED_PONG (0x85), pindah ke handler khusus
    if (buffer[0] == 0x84) {
        handleConnectedPing(client_addr, buffer, length);
        return;
    }
    if (buffer[0] == 0x85) {
        handleConnectedPong(client_addr, buffer, length);
        return;
    }

    auto conn = findOrCreateConnection(client_addr);
    RakNetHeader header = RakNetHeader::parse(buffer, length);

    std::cout << "[Connected] Flags=0x" << std::hex << (int)header.flags
              << " Seq=" << std::dec << header.sequence << std::endl;

    // Wajib ACK setiap sequence client
    sendAckPacket(client_addr, header.sequence);

    //
    // --- Reliability Parsing ---
    //
    uint8_t reliability = (header.flags >> 5) & 0x07;
    size_t offset = 3; // minimal header size

    // Jika reliable → add reliableIndex (2 bytes) + messageIndex (1 byte)
    if (reliability == 2 || reliability == 3 || reliability == 4 || reliability == 6 || reliability == 7) {
        offset += 3;
    }

    // Jika ordered/sequenced → add ordering channel info
    if (reliability == 1 || reliability == 3 || reliability == 4) {
        offset += 4;
    }

    if (offset >= length) return;

    uint8_t* payload = buffer + offset;
    size_t payload_len = length - offset;

    if (payload_len == 0) return;

    uint8_t pk = payload[0];

    std::cout << "[Bedrock] Packet ID: 0x" << std::hex << (int)pk << std::dec << std::endl;

    switch (pk) {
        case 0x01: // LOGIN
            handleBedrockLogin(*conn, payload, payload_len);
            break;

        case 0x03: // CLIENT_TO_SERVER_HANDSHAKE  // ========== TAMBAHKAN INI ==========
            std::cout << "[Bedrock] CLIENT_TO_SERVER_HANDSHAKE" << std::endl;
            handleEncryptionHandshake(*conn, payload, payload_len);
            break;

        case 0x04: // SERVER_TO_CLIENT_HANDSHAKE  // ========== TAMBAHKAN INI ==========
            std::cout << "[Bedrock] SERVER_TO_CLIENT_HANDSHAKE RESPONSE" << std::endl;
            handleEncryptionHandshake(*conn, payload, payload_len);
            break;

        case 0x90: // CLIENT_CONNECT
            std::cout << "[Bedrock] CLIENT_CONNECT" << std::endl;
            handleClientConnect(client_addr, payload, payload_len);
            break;

        default:
            std::cout << "[Bedrock] Unknown: 0x" << std::hex << (int)pk << std::dec << std::endl;
            logHex(payload, std::min(payload_len, (size_t)16), "Payload: ");
            break;
    }
}*/

// ==================== IMPROVED PACKET CREATION ====================

std::vector<uint8_t> RakNetServer::createResourcePackStack() {
    std::vector<uint8_t> packet;
    
    // Packet ID untuk Resource Pack Stack
    packet.push_back(0x07); 
    
    // Force acceptance (0)
    packet.push_back(0x00);
    
    // Behavior pack count (0)
    packet.push_back(0x00);
    
    // Resource pack count (0)  
    packet.push_back(0x00);
    
    // Game version (1.21.50)
    const char* game_version = "1.21.50";
    packet.push_back(strlen(game_version));
    for (const char* p = game_version; *p; p++) {
        packet.push_back(static_cast<uint8_t>(*p));
    }
    
    return packet;
}

std::vector<uint8_t> RakNetServer::createStartGamePacket() {
    std::vector<uint8_t> packet;
    
    // Packet ID untuk Start Game
    packet.push_back(0x0B); 
    
    // Entity ID (0 untuk player)
    for (int i = 0; i < 8; i++) packet.push_back(0x00);
    
    // Runtime Entity ID (0 untuk player)  
    for (int i = 0; i < 8; i++) packet.push_back(0x00);
    
    // Player Game Mode (1 untuk creative)
    packet.push_back(0x01);
    
    // Player position (0, 100, 0 - spawn point)
    float x = 0.0f, y = 100.0f, z = 0.0f;
    uint8_t* x_bytes = reinterpret_cast<uint8_t*>(&x);
    uint8_t* y_bytes = reinterpret_cast<uint8_t*>(&y); 
    uint8_t* z_bytes = reinterpret_cast<uint8_t*>(&z);
    
    for (int i = 0; i < 4; i++) packet.push_back(x_bytes[i]);
    for (int i = 0; i < 4; i++) packet.push_back(y_bytes[i]);
    for (int i = 0; i < 4; i++) packet.push_back(z_bytes[i]);
    
    // Rotation (0, 0)
    float yaw = 0.0f, pitch = 0.0f;
    uint8_t* yaw_bytes = reinterpret_cast<uint8_t*>(&yaw);
    uint8_t* pitch_bytes = reinterpret_cast<uint8_t*>(&pitch);
    
    for (int i = 0; i < 4; i++) packet.push_back(yaw_bytes[i]);
    for (int i = 0; i < 4; i++) packet.push_back(pitch_bytes[i]);
    
    // Seed (0)
    for (int i = 0; i < 4; i++) packet.push_back(0x00);
    
    // Biome type (1 = plains)
    packet.push_back(0x01);
    
    // User UUID (0)
    for (int i = 0; i < 8; i++) packet.push_back(0x00);
    
    return packet;
}

// ==================== EXISTING FUNCTIONS (MAINTAINED) ====================

bool RakNetServer::start(uint16_t port) {
    socket_ = socket(AF_INET, SOCK_DGRAM, 0);
    if (socket_ < 0) {
        perror("socket() failed");
        return false;
    }
    
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    addr.sin_addr.s_addr = INADDR_ANY;
    
    if (bind(socket_, (sockaddr*)&addr, sizeof(addr)) < 0) {
        perror("bind() failed");
        close(socket_);
        return false;
    }
    
    std::cout << "GarudaMC Server listening on port " << port << std::endl;
    running_ = true;
    return true;
}

void RakNetServer::handleUnconnectedPing(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    std::cout << "Handling UNCONNECTED_PING" << std::endl;
    
    // Build response
    std::vector<uint8_t> response;
    
    // Packet ID
    response.push_back(0x1C); // UNCONNECTED_PONG
    
    // Ping time (copy from request)
    if (length >= 9) {
        for (int i = 1; i <= 8; i++) {
            response.push_back(buffer[i]);
        }
    } else {
        for (int i = 0; i < 8; i++) response.push_back(0x00);
    }
    
    // Server GUID
    uint64_t server_guid = 1234567890;
    for (int i = 0; i < 8; i++) {
        response.push_back((server_guid >> (i * 8)) & 0xFF);
    }
    
    // Magic
    uint8_t magic[] = {0x00, 0xFF, 0xFF, 0x00, 0xFE, 0xFE, 0xFE, 0xFE, 0xFD, 0xFD, 0xFD, 0xFD, 0x12, 0x34, 0x56, 0x78};
    response.insert(response.end(), magic, magic + 16);
    
    // Server info string - FORMAT BERDASARKAN server.properties
    std::string server_info =
    "MCPE;"
    "GarudaMC Server;" // Server name 
    "757;" // Protocol version (1.21.50)
    "1.21.50;" // Game version
    "0;"  // Player count  
    "250;" // Max players
    "13253860892328930865;" // Server ID (random)
    "Bedrock level;" // Level name
    "Survival;"  // Game mode
    "1;" // Default game mode
    "19132;" // IPv4 port
    "19133"; // IPv6 port
    
    // Length of server info (big endian)
    uint16_t info_len = server_info.length();
    response.push_back((info_len >> 8) & 0xFF);
    response.push_back(info_len & 0xFF);
    
    // Server info content
    response.insert(response.end(), server_info.begin(), server_info.end());
    
    sendto(socket_, response.data(), response.size(), 0,
           (sockaddr*)&client_addr, sizeof(client_addr));
    
    std::cout << "Sent UNCONNECTED_PONG (" << response.size() << " bytes)" << std::endl;
    std::cout << "Server info: " << server_info << std::endl;
    logHex(response.data(), std::min(response.size(), size_t(64)), "PONG: ");
}

void RakNetServer::handleOpenConnectionRequest1(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    std::cout << "Handling OPEN_CONNECTION_REQUEST_1" << std::endl;
    
    uint8_t response[28] = {
        0x06, // OPEN_CONNECTION_REPLY_1
        0x00, 0xFF, 0xFF, 0x00, 0xFE, 0xFE, 0xFE, 0xFE, 0xFD, 0xFD, 0xFD, 0xFD, 0x12, 0x34, 0x56, 0x78, // Magic
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // Server GUID
        0x00 // Security (no security)
    };
    
    sendto(socket_, response, sizeof(response), 0,
           (sockaddr*)&client_addr, sizeof(client_addr));
    
    std::cout << "Sent OPEN_CONNECTION_REPLY_1" << std::endl;
}

void RakNetServer::handleOpenConnectionRequest2(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    std::cout << "Handling OPEN_CONNECTION_REQUEST_2" << std::endl;
    
    uint8_t response[31] = {
        0x08, // OPEN_CONNECTION_REPLY_2
        0x00, 0xFF, 0xFF, 0x00, 0xFE, 0xFE, 0xFE, 0xFE, 0xFD, 0xFD, 0xFD, 0xFD, 0x12, 0x34, 0x56, 0x78, // Magic
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // Server GUID
        0x00, 0x00, // Client address (port)
        0x00, 0x00, 0x00, 0x00 // MTU size
    };
    
    sendto(socket_, response, sizeof(response), 0,
           (sockaddr*)&client_addr, sizeof(client_addr));
    
    std::cout << "Sent OPEN_CONNECTION_REPLY_2" << std::endl;
    
    // === FIX: SEND MISSING PACKETS ===
    std::cout << "Sending connection completion packets..." << std::endl;
    
    // 1. Send CONNECTION_REQUEST_ACCEPTED (0x10)
    uint8_t conn_accepted[21] = {
        0x10, // CONNECTION_REQUEST_ACCEPTED
        0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // Server GUID
        0x00, 0x00, // Ping time
        0x00 // Use security
    };
    
    sendto(socket_, conn_accepted, sizeof(conn_accepted), 0,
           (sockaddr*)&client_addr, sizeof(client_addr));
    std::cout << "Sent CONNECTION_REQUEST_ACCEPTED (0x10)" << std::endl;
    
    // 2. Short delay then send NEW_INCOMING_CONNECTION (0x13)
    usleep(50000); // 50ms delay
    sendNewIncomingConnection(client_addr);
    
    // Mark as connected
    auto conn = findOrCreateConnection(client_addr);
    conn->connected = true;
    
    std::cout << "=== CONNECTION HANDSHAKE COMPLETED ===" << std::endl;
    
    std::cout << "=== INITIATING ENCRYPTION HANDSHAKE ===" << std::endl;
    sendEncryptionRequest(*conn);
    std::cout << "Waiting for client handshake response..." << std::endl;
}

void RakNetServer::sendAckPacket(sockaddr_in client_addr, uint16_t sequence_number) {
    uint8_t ack_packet[4] = {
        0xc0, // ACK packet ID
        static_cast<uint8_t>(sequence_number & 0xFF),        // Sequence low
        static_cast<uint8_t>((sequence_number >> 8) & 0xFF), // Sequence high  
        0x00  // Additional flag
    };
    
    sendto(socket_, ack_packet, sizeof(ack_packet), 0,
           (sockaddr*)&client_addr, sizeof(client_addr));
    
    std::cout << "Sent ACK for sequence " << sequence_number << std::endl;
}

void RakNetServer::sendNewIncomingConnection(sockaddr_in client_addr) {
    std::vector<uint8_t> packet;
    
    packet.push_back(0x13); // NEW_INCOMING_CONNECTION
    
    // Server GUID (8 bytes)
    for (int i = 0; i < 7; i++) packet.push_back(0x00);
    packet.push_back(0x01);
    
    // Client port (2 bytes)
    packet.push_back((client_addr.sin_port >> 8) & 0xFF);
    packet.push_back(client_addr.sin_port & 0xFF);
    
    // Client IP (4 bytes)
    packet.push_back(client_addr.sin_addr.s_addr & 0xFF);
    packet.push_back((client_addr.sin_addr.s_addr >> 8) & 0xFF);
    packet.push_back((client_addr.sin_addr.s_addr >> 16) & 0xFF);
    packet.push_back((client_addr.sin_addr.s_addr >> 24) & 0xFF);
    
    // Timestamp (4 bytes)
    for (int i = 0; i < 4; i++) packet.push_back(0x00);
    
    sendto(socket_, packet.data(), packet.size(), 0,
           (sockaddr*)&client_addr, sizeof(client_addr));
    
    std::cout << "Sent NEW_INCOMING_CONNECTION (" << packet.size() << " bytes)" << std::endl;
}

void RakNetServer::handleClientConnect(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    std::cout << "=== HANDLING CLIENT_CONNECT ===" << std::endl;
    
    auto conn = findOrCreateConnection(client_addr);
    
    // 1. Send Login Success 
    std::cout << "Sending LOGIN_SUCCESS..." << std::endl;
    std::vector<uint8_t> play_status;
    play_status.push_back(0x02); // PlayStatus packet ID
    play_status.push_back(0x00); // LOGIN_SUCCESS (4 byte little endian)
    play_status.push_back(0x00);
    play_status.push_back(0x00); 
    play_status.push_back(0x00);
    sendPacket(*conn, play_status);
    
    // 2. Send Resource Pack Stack (SIMPLE VERSION)
    std::cout << "Sending RESOURCE_PACK_STACK..." << std::endl;
    std::vector<uint8_t> resource_pack;
    resource_pack.push_back(0x07); // ResourcePackStack packet ID
    resource_pack.push_back(0x00); // Force acceptance
    resource_pack.push_back(0x00); // Behavior pack count
    resource_pack.push_back(0x00); // Resource pack count
    // Game version
    const char* game_ver = "1.21.50";
    resource_pack.push_back(strlen(game_ver));
    for (const char* p = game_ver; *p; p++) {
        resource_pack.push_back(static_cast<uint8_t>(*p));
    }
    sendPacket(*conn, resource_pack);
    
    // 3. Send Start Game Packet (SIMPLE VERSION)
    std::cout << "Sending START_GAME..." << std::endl;
    std::vector<uint8_t> start_game;
    start_game.push_back(0x0B); // StartGame packet ID
    
    // Entity ID & Runtime Entity ID (0 untuk player)
    for (int i = 0; i < 16; i++) start_game.push_back(0x00);
    
    // Game mode (1 = creative)
    start_game.push_back(0x01);
    
    // Position (0, 100, 0)
    float pos[3] = {0.0f, 100.0f, 0.0f};
    for (int i = 0; i < 3; i++) {
        uint8_t* bytes = reinterpret_cast<uint8_t*>(&pos[i]);
        for (int j = 0; j < 4; j++) start_game.push_back(bytes[j]);
    }
    
    // Rotation (0, 0)
    float rot[2] = {0.0f, 0.0f};
    for (int i = 0; i < 2; i++) {
        uint8_t* bytes = reinterpret_cast<uint8_t*>(&rot[i]);
        for (int j = 0; j < 4; j++) start_game.push_back(bytes[j]);
    }
    
    // Seed (0)
    for (int i = 0; i < 4; i++) start_game.push_back(0x00);
    
    // Biome type (1 = plains)
    start_game.push_back(0x01);
    
    sendPacket(*conn, start_game);
    
    std::cout << "=== COMPLETE LOGIN SEQUENCE SENT ===" << std::endl;
}

Connection* RakNetServer::findOrCreateConnection(sockaddr_in client_addr) {
    std::string address = inet_ntoa(client_addr.sin_addr);
    uint16_t port = ntohs(client_addr.sin_port);
    
    for (auto& conn : connections_) {
        if (conn.address == address && conn.port == port) {
            return &conn;
        }
    }
    
    Connection new_conn;
    new_conn.address = address;
    new_conn.port = port;
    new_conn.guid = connections_.size() + 1;
    new_conn.connected = false;
    new_conn.sequence_in = 0;
    new_conn.sequence_out = 1;
    new_conn.last_activity = std::chrono::steady_clock::now();
    
    connections_.push_back(new_conn);
    std::cout << "Created new connection: " << address << ":" << port << " (GUID: " << new_conn.guid << ")" << std::endl;
    
    return &connections_.back();
}

void RakNetServer::stop() {
    if (socket_ >= 0) {
        close(socket_);
        socket_ = -1;
    }
    running_ = false;
    connections_.clear();
    std::cout << "Server stopped" << std::endl;
}

void RakNetServer::sendPacket(const Connection& conn, const std::vector<uint8_t>& data) {
    // Use non-const reference for reliable packet sending
    Connection& nonConstConn = const_cast<Connection&>(conn);
    sendReliablePacket(nonConstConn, data);
}

// ==================== IMPROVED PING/PONG ====================

void RakNetServer::handleConnectedPing(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    std::cout << "=== HANDLING CONNECTED_PING (0x84) ===" << std::endl;
    
    if (length < 5) return;
    
    // Extract ping time (4 bytes little endian) - dimulai dari offset 1
    uint32_t ping_time = 0;
    memcpy(&ping_time, &buffer[1], 4);
    
    std::cout << "Ping time: " << ping_time << std::endl;
    
    // Send CONNECTED_PONG response (0x85)
    std::vector<uint8_t> response;
    response.push_back(0x85); // CONNECTED_PONG packet ID
    
    // Ping time (copy from request)
    for (int i = 0; i < 4; i++) {
        response.push_back((ping_time >> (i * 8)) & 0xFF);
    }
    
    // Pong time (same as ping time for now)
    for (int i = 0; i < 4; i++) {
        response.push_back((ping_time >> (i * 8)) & 0xFF);
    }
    
    sendto(socket_, response.data(), response.size(), 0,
           (sockaddr*)&client_addr, sizeof(client_addr));
    
    std::cout << "Sent CONNECTED_PONG (0x85) response" << std::endl;
    logHex(response.data(), response.size(), "PONG: ");
}

void RakNetServer::handleConnectedPong(sockaddr_in client_addr, uint8_t* buffer, size_t length) {
    std::cout << "Received CONNECTED_PONG" << std::endl;
    // Optional: Handle pong response if needed
}