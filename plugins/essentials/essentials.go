package main

import (
    "log"
    "strings"

    "garuda/internal/protocol/minecraft"
    "garuda/internal/world"
    "garuda/pkg/plugin"
)

// EssentialsPlugin provides basic server commands and features
type EssentialsPlugin struct {
    plugin.BasePlugin
    manager *plugin.PluginManager
}

// PluginInstance is the exported symbol that Garuda will look for
var PluginInstance plugin.Plugin = &EssentialsPlugin{}

func (p *EssentialsPlugin) GetName() string {
    return "Essentials"
}

func (p *EssentialsPlugin) GetVersion() string {
    return "1.0.0"
}

func (p *EssentialsPlugin) GetAuthor() string {
    return "Garuda Team"
}

func (p *EssentialsPlugin) OnEnable(manager *plugin.PluginManager) {
    p.manager = manager
    log.Printf("Essentials plugin enabled!")
}

func (p *EssentialsPlugin) OnDisable() {
    log.Printf("Essentials plugin disabled!")
}

func (p *EssentialsPlugin) OnPlayerJoin(player *world.Player) {
    api := plugin.NewGarudaAPI(p.manager)
    api.BroadcastMessage("§e" + player.Username + " joined the game")
}

func (p *EssentialsPlugin) OnPlayerQuit(player *world.Player) {
    api := plugin.NewGarudaAPI(p.manager)
    api.BroadcastMessage("§e" + player.Username + " left the game")
}

func (p *EssentialsPlugin) OnPlayerChat(player *world.Player, message string) bool {
    // Handle commands
    if strings.HasPrefix(message, "/") {
        return p.handleCommand(player, message)
    }
    
    // Format chat message
    api := plugin.NewGarudaAPI(p.manager)
    api.BroadcastMessage("§7" + player.Username + ": " + message)
    return false // Cancel original chat
}

func (p *EssentialsPlugin) handleCommand(player *world.Player, message string) bool {
    parts := strings.Split(message[1:], " ") // Remove leading slash
    command := strings.ToLower(parts[0])
    args := parts[1:]

    api := plugin.NewGarudaAPI(p.manager)

    switch command {
    case "help":
        api.SendMessageToPlayer(player, "§6--- Help ---")
        api.SendMessageToPlayer(player, "§a/help §7- Show this help")
        api.SendMessageToPlayer(player, "§a/gamemode <mode> §7- Change gamemode")
        api.SendMessageToPlayer(player, "§a/tp <player> §7- Teleport to player")
        return false
        
    case "gamemode", "gm":
        if len(args) < 1 {
            api.SendMessageToPlayer(player, "§cUsage: /gamemode <survival|creative|adventure>")
            return false
        }
        
        switch strings.ToLower(args[0]) {
        case "0", "survival":
            player.GameMode = 0
            api.SendMessageToPlayer(player, "§aGamemode set to Survival")
        case "1", "creative":
            player.GameMode = 1
            api.SendMessageToPlayer(player, "§aGamemode set to Creative")
        case "2", "adventure":
            player.GameMode = 2
            api.SendMessageToPlayer(player, "§aGamemode set to Adventure")
        default:
            api.SendMessageToPlayer(player, "§cInvalid gamemode")
        }
        return false
        
    case "tp":
        if len(args) < 1 {
            api.SendMessageToPlayer(player, "§cUsage: /tp <player>")
            return false
        }
        
        target := api.GetPlayer(args[0])
        if target == nil {
            api.SendMessageToPlayer(player, "§cPlayer not found")
            return false
        }
        
        player.Position = target.Position
        api.SendMessageToPlayer(player, "§aTeleported to " + target.Username)
        return false
        
    default:
        api.SendMessageToPlayer(player, "§cUnknown command. Type /help for help.")
        return false
    }
}

func (p *EssentialsPlugin) OnPlayerCommand(player *world.Player, command string, args []string) bool {
    // We handle commands in OnPlayerChat for now
    return true
}