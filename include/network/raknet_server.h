#pragma once
#include <cstdint>
#include <vector>
#include <string>
#include <netinet/in.h>
#include <chrono>
#include <unordered_map>
#include <cstdio>
#include <iostream>


struct SentPacket {
    std::vector<uint8_t> data;
    std::chrono::steady_clock::time_point timestamp;
    uint16_t sequence;
};

class Connection {
public:
    std::string address;
    uint16_t port;
    uint64_t guid;
    bool connected;
    
    // Improved connection management
    uint16_t sequence_in;
    uint16_t sequence_out;
    std::chrono::steady_clock::time_point last_activity;
    std::unordered_map<uint16_t, SentPacket> sent_packets;
    
    Connection() : connected(false), sequence_in(0), sequence_out(1) {}
};

class RakNetServer {
private:

    void logHex(const uint8_t* data, size_t length, const std::string& prefix = "") {
    std::cout << prefix;
    for (size_t i = 0; i < std::min(length, size_t(32)); i++) {
        printf("%02X ", data[i]);
    }
    if (length > 32) std::cout << "...";
    std::cout << std::endl;
}

    int socket_;
    bool running_;
    std::vector<Connection> connections_;
    
    // Existing functions
    // void sendStartGamePacket(Connection& conn);
    void sendNewIncomingConnection(sockaddr_in client_addr);
    void sendAckPacket(sockaddr_in client_addr, uint16_t sequence_number);
    void handleUnconnectedPing(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleOpenConnectionRequest1(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleOpenConnectionRequest2(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleConnectionRequest(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleNewIncomingConnection(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleClientConnect(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    // void sendResourcePackStack(Connection& conn);
    
    Connection* findOrCreateConnection(sockaddr_in client_addr);
    
    // NEW IMPROVED FUNCTIONS
    void handleConnectedPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleBedrockLogin(Connection& conn, uint8_t* buffer, size_t length);
    void handleAckPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleNackPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleConnectedPing(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleConnectedPong(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleDisconnection(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    
    void sendLoginSequence(Connection& conn);
    void sendReliablePacket(Connection& conn, const std::vector<uint8_t>& payload);
    void sendRawPacket(const Connection& conn, const std::vector<uint8_t>& data);
    std::vector<uint8_t> addRakNetHeader(const std::vector<uint8_t>& payload, uint16_t sequence, uint8_t reliability);
    
    std::vector<uint8_t> createResourcePackStack();
    std::vector<uint8_t> createStartGamePacket();
    
    void handleResends();
    void cleanupStaleConnections();
    void startConnectionTimer(Connection& conn);
    bool isConnectionTimedOut(const Connection& conn);
    
public:
    RakNetServer() : socket_(-1), running_(false) {}
    bool start(uint16_t port);
    void stop();
    void handleIncomingPackets();
    void sendPacket(const Connection& conn, const std::vector<uint8_t>& data);
    int getSocket() const { return socket_; }
};