#pragma once
#include <string>

class Player {
private:
    std::string name;
    uint64_t guid;
    
public:
    Player();
    void sendMessage(const std::string& message);
};