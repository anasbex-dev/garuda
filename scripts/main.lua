-- GarudaMC Main Script
local Garuda = {}

-- Load modules
local Events = require("events.player_events")
local Database = require("modules.database")

function Garuda.onServerStart()
    print("GarudaMC Server Starting...")
    
    -- Initialize systems
    Database.initialize()
    Events.registerAll()
    
    -- Load plugins
    Garuda.loadPlugin("economy")
    Garuda.loadPlugin("combat")
end

function Garuda.onPlayerJoin(player)
    print("Player joined: " .. player:getName())
    
    -- Custom welcome message
    player:sendMessage("Welcome to GarudaMC Server!")
    player:setGamemode("creative")
    
    -- Teleport to spawn
    player:teleport(0, 100, 0)
end

function Garuda.loadPlugin(pluginName)
    local success, plugin = pcall(require, "plugins." .. pluginName .. ".init")
    if success then
        plugin.initialize()
        print("Loaded plugin: " .. pluginName)
    else
        print("Failed to load plugin: " .. pluginName)
    end
end

return Garuda