#pragma once
#include <vector>
#include <cstdint>
#include <iostream>
#include <random>

class RSAKey {
public:
    RSAKey() {
        std::cout << "RSAKey constructor" << std::endl;
    }
    
    ~RSAKey() {
        std::cout << "RSAKey destructor" << std::endl;
    }
    
    bool generateKeyPair() {
        std::cout << "Generating RSA key pair (SIMULATED)..." << std::endl;
        return true;
    }
    
    std::vector<uint8_t> getPublicKey() const {
        std::vector<uint8_t> dummy_key(162, 0xAA);
        std::cout << "Returning simulated RSA public key (" << dummy_key.size() << " bytes)" << std::endl;
        return dummy_key;
    }
};

class EncryptionHandler {
public:
    EncryptionHandler() {
        std::cout << "EncryptionHandler constructor" << std::endl;
        initialize();
    }
    
    ~EncryptionHandler() {
        std::cout << "EncryptionHandler destructor" << std::endl;
    }
    
    bool initialize() {
        std::cout << "Initializing encryption handler..." << std::endl;
        return server_key_.generateKeyPair();
    }
    
    std::vector<uint8_t> getServerPublicKey() const {
        return server_key_.getPublicKey();
    }
    
    bool verifyClientToken(const std::vector<uint8_t>& token) {
        std::cout << "Verifying client token (" << token.size() << " bytes)..." << std::endl;
        std::cout << "Client token verification SUCCESS (SIMULATED)" << std::endl;
        return true;
    }
    
    std::vector<uint8_t> generateHandshakeToken() {
        std::cout << "Generating handshake token..." << std::endl;
        
        std::vector<uint8_t> token(16);
        std::random_device rd;
        std::mt19937 gen(rd());
        std::uniform_int_distribution<uint8_t> dis(0, 255);
        
        for(auto& byte : token) {
            byte = dis(gen);
        }
        
        std::cout << "Generated handshake token: ";
        for(size_t i = 0; i < std::min(token.size(), size_t(8)); i++) {
            printf("%02X ", token[i]);
        }
        std::cout << "..." << std::endl;
        
        return token;
    }
    
private:
    RSAKey server_key_;
};