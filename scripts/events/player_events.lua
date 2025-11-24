local PlayerEvents = {}

function PlayerEvents.registerAll()
    -- Register event handlers
    registerEvent("player_join", PlayerEvents.onPlayerJoin)
    registerEvent("player_chat", PlayerEvents.onPlayerChat)
    registerEvent("player_move", PlayerEvents.onPlayerMove)
end

function PlayerEvents.onPlayerJoin(player)
    -- Custom join logic
    player:setHealth(100)
    player:addItem("diamond_sword", 1)
end

function PlayerEvents.onPlayerChat(player, message)
    if message == "!spawn" then
        player:teleport(0, 100, 0)
        return false -- Cancel original message
    end
end

return PlayerEvents