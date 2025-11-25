#pragma once
#include <cstdint>
#include <vector>
#include <string>
#include <netinet/in.h>
#include <chrono>
#include <unordered_map>
#include <iostream>

#ifndef RAKNET_SERVER_H
#define RAKNET_SERVER_H

#include "auth/encryption_handler.h"

class Connection {
public:
    std::string address;
    uint16_t port;
    uint64_t guid;
    bool connected;
    bool encrypted;
    
    // Sequence numbers
    uint16_t sequence_in;
    uint16_t sequence_out;
    
    // Timing
    std::chrono::steady_clock::time_point last_activity;
    
    // Sent packets for reliability
    struct SentPacket {
        std::vector<uint8_t> data;
        std::chrono::steady_clock::time_point timestamp;
        uint16_t sequence;
    };
    std::unordered_map<uint16_t, SentPacket> sent_packets;
    
    // Authentication
    std::string username;
    bool authenticated;
    
    Connection() : connected(false), encrypted(false), sequence_in(0), 
                  sequence_out(1), authenticated(false) {}
};

class RakNetServer {
private:
    int socket_;
    bool running_;
    std::vector<Connection> connections_;
    
    // Modern Authentication
    EncryptionHandler encryption_handler_;
    void sendLoginSuccess(Connection& conn);
    
    // Existing functions...
    void handleUnconnectedPing(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleOpenConnectionRequest1(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleOpenConnectionRequest2(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleConnectionRequest(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleNewIncomingConnection(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleClientConnect(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void sendResourcePackStack(Connection& conn);
    void sendStartGamePacket(Connection& conn);
    
    // Improved packet handling
    void handleConnectedPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleBedrockLogin(Connection& conn, uint8_t* buffer, size_t length);
    void handleConnectedPing(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleConnectedPong(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleAckPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleNackPacket(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    void handleDisconnection(sockaddr_in client_addr, uint8_t* buffer, size_t length);
    
    // Modern authentication handlers
    void handleModernLogin(Connection& conn, uint8_t* buffer, size_t length);
    void handleEncryptionHandshake(Connection& conn, uint8_t* buffer, size_t length);
    void sendEncryptionRequest(Connection& conn);
    void completeModernLogin(Connection& conn);
    
    // Utility functions
    Connection* findOrCreateConnection(sockaddr_in client_addr);
    void startConnectionTimer(Connection& conn);
    bool isConnectionTimedOut(const Connection& conn);
    void cleanupStaleConnections();
    void handleResends();
    
    // Packet sending
    void sendRawPacket(const Connection& conn, const std::vector<uint8_t>& data);
    void sendReliablePacket(Connection& conn, const std::vector<uint8_t>& payload);
    std::vector<uint8_t> addRakNetHeader(const std::vector<uint8_t>& payload, uint16_t sequence, uint8_t reliability);
    void sendAckPacket(sockaddr_in client_addr, uint16_t sequence_number);
    void sendNewIncomingConnection(sockaddr_in client_addr);
    void sendLoginSequence(Connection& conn);

    // Packet creation
    std::vector<uint8_t> createResourcePackStack();
    std::vector<uint8_t> createStartGamePacket();
    void logHex(const uint8_t* data, size_t length, const std::string& description = "");
    
public:
    RakNetServer();
    bool start(uint16_t port);
    void stop();
    void handleIncomingPackets();
    void sendPacket(const Connection& conn, const std::vector<uint8_t>& data);
    int getSocket() const { return socket_; }
};

#endif // RAKNET_SERVER_H