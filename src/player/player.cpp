#include "player/player.h"
#include <iostream>

// Temporary empty implementations
Player::Player() {
    std::cout << "Player created" << std::endl;
}

void Player::sendMessage(const std::string& message) {
    std::cout << "Send to player: " << message << std::endl;
}