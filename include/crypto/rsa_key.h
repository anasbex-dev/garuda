#pragma once
#include <vector>
#include <string>
#include <memory>

class RSAKey {
private:
    struct Impl;
    std::unique_ptr<Impl> pimpl;
    
public:
    RSAKey();
    ~RSAKey();
    
    // Prevent copying
    RSAKey(const RSAKey&) = delete;
    RSAKey& operator=(const RSAKey&) = delete;
    
    // Allow moving
    RSAKey(RSAKey&&) noexcept;
    RSAKey& operator=(RSAKey&&) noexcept;
    
    bool generateKeyPair(int bits = 2048);
    bool loadFromPEM(const std::string& pem_data);
    std::string getPublicKeyPEM() const;
    std::string getPrivateKeyPEM() const;
    
    std::vector<uint8_t> encrypt(const std::vector<uint8_t>& data) const;
    std::vector<uint8_t> decrypt(const std::vector<uint8_t>& data) const;
    std::vector<uint8_t> sign(const std::vector<uint8_t>& data) const;
    bool verify(const std::vector<uint8_t>& data, const std::vector<uint8_t>& signature) const;
};