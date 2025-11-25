#pragma once
#include <string>
#include <vector>
#include <unordered_map>

class JWTVerifier {
public:
    struct JWTPayload {
        std::string username;
        std::string xuid;
        std::string identity_public_key;
        std::string title_id;
        std::unordered_map<std::string, std::string> extra_claims;
    };
    
    static bool verifyJWTChain(const std::vector<uint8_t>& jwtData);
    static JWTPayload parseJWTPayload(const std::string& jwtToken);
    static std::string extractUsernameFromChain(const std::vector<uint8_t>& jwtData);
    
private:
    static std::vector<std::string> splitJWTChain(const std::vector<uint8_t>& jwtData);
    static bool verifyJWTSignature(const std::string& token, const std::string& publicKey);
};