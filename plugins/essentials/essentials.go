package essentials

import (
    "fmt"
    "garuda/pkg/plugin"
    "garuda/world"
    "strconv"
)

type EssentialsPlugin struct {
    manager plugin.PluginManager
    api     *plugin.PluginAPI
}

func NewEssentialsPlugin() *EssentialsPlugin {
    return &EssentialsPlugin{}
}

func (p *EssentialsPlugin) GetName() string {
    return "Essentials"
}

func (p *EssentialsPlugin) GetVersion() string {
    return "1.0.0"
}

func (p *EssentialsPlugin) GetAuthor() string {
    return "Garuda Team"
}

func (p *EssentialsPlugin) GetDescription() string {
    return "Essential commands and features for Garuda server"
}

func (p *EssentialsPlugin) OnEnable(manager plugin.PluginManager) {
    p.manager = manager
    p.api = plugin.NewPluginAPI(manager)
    
    p.registerCommands()
    p.registerEvents()
    
    p.api.BroadcastMessage("Essentials plugin enabled!")
}

func (p *EssentialsPlugin) OnDisable() {
    p.api.BroadcastMessage("Essentials plugin disabled!")
}

func (p *EssentialsPlugin) registerCommands() {
    p.manager.RegisterCommand("gamemode", p.handleGameModeCommand)
    p.manager.RegisterCommand("tp", p.handleTeleportCommand)
    p.manager.RegisterCommand("give", p.handleGiveCommand)
    p.manager.RegisterCommand("time", p.handleTimeCommand)
    p.manager.RegisterCommand("heal", p.handleHealCommand)
}

func (p *EssentialsPlugin) registerEvents() {
    p.manager.RegisterEvent(plugin.EventPlayerJoin, p.handlePlayerJoin)
    p.manager.RegisterEvent(plugin.EventPlayerChat, p.handlePlayerChat)
}

func (p *EssentialsPlugin) handleGameModeCommand(sender plugin.CommandSender, command string, args []string) bool {
    if !sender.HasPermission("essentials.gamemode") {
        sender.SendMessage("You don't have permission to use this command")
        return true
    }
    
    if len(args) < 1 {
        sender.SendMessage("Usage: /gamemode <0|1|2|3> [player]")
        return true
    }
    
    gamemode, err := strconv.Atoi(args[0])
    if err != nil || gamemode < 0 || gamemode > 3 {
        sender.SendMessage("Invalid gamemode. Use 0 (Survival), 1 (Creative), 2 (Adventure), or 3 (Spectator)")
        return true
    }
    
    var target world.Player
    if sender.IsPlayer() {
        target = p.manager.GetPlayer(sender.GetName())
    }
    
    if len(args) > 1 {
        target = p.manager.GetPlayer(args[1])
        if target == nil {
            sender.SendMessage("Player not found: " + args[1])
            return true
        }
    }
    
    if target == nil {
        sender.SendMessage("You must specify a player when using from console")
        return true
    }
    
    // This would actually change the player's gamemode in a real implementation
    sender.SendMessage(fmt.Sprintf("Set %s's gamemode to %d", target.Username, gamemode))
    return true
}

func (p *EssentialsPlugin) handleTeleportCommand(sender plugin.CommandSender, command string, args []string) bool {
    if !sender.HasPermission("essentials.teleport") {
        sender.SendMessage("You don't have permission to use this command")
        return true
    }
    
    if len(args) < 2 {
        sender.SendMessage("Usage: /tp <x> <y> <z> OR /tp <player>")
        return true
    }
    
    if sender.IsPlayer() {
        // This would actually teleport the player in a real implementation
        sender.SendMessage(fmt.Sprintf("Teleported to %s, %s, %s", args[0], args[1], args[2]))
    } else {
        sender.SendMessage("Console cannot teleport itself")
    }
    
    return true
}

func (p *EssentialsPlugin) handleGiveCommand(sender plugin.CommandSender, command string, args []string) bool {
    if !sender.HasPermission("essentials.give") {
        sender.SendMessage("You don't have permission to use this command")
        return true
    }
    
    if len(args) < 2 {
        sender.SendMessage("Usage: /give <player> <item> [amount]")
        return true
    }
    
    target := p.manager.GetPlayer(args[0])
    if target == nil {
        sender.SendMessage("Player not found: " + args[0])
        return true
    }
    
    itemID, err := strconv.Atoi(args[1])
    if err != nil {
        sender.SendMessage("Invalid item ID: " + args[1])
        return true
    }
    
    amount := 1
    if len(args) > 2 {
        amount, err = strconv.Atoi(args[2])
        if err != nil {
            sender.SendMessage("Invalid amount: " + args[2])
            return true
        }
    }
    
    // This would actually give the item to the player in a real implementation
    sender.SendMessage(fmt.Sprintf("Gave %d of item %d to %s", amount, itemID, target.Username))
    return true
}

func (p *EssentialsPlugin) handleTimeCommand(sender plugin.CommandSender, command string, args []string) bool {
    if !sender.HasPermission("essentials.time") {
        sender.SendMessage("You don't have permission to use this command")
        return true
    }
    
    if len(args) < 1 {
        sender.SendMessage("Usage: /time <set|add> <value>")
        return true
    }
    
    switch args[0] {
    case "set":
        if len(args) < 2 {
            sender.SendMessage("Usage: /time set <value>")
            return true
        }
        sender.SendMessage("Time set to " + args[1])
    case "add":
        if len(args) < 2 {
            sender.SendMessage("Usage: /time add <value>")
            return true
        }
        sender.SendMessage("Added " + args[1] + " to time")
    default:
        sender.SendMessage("Unknown time operation: " + args[0])
    }
    
    return true
}

func (p *EssentialsPlugin) handleHealCommand(sender plugin.CommandSender, command string, args []string) bool {
    if !sender.HasPermission("essentials.heal") {
        sender.SendMessage("You don't have permission to use this command")
        return true
    }
    
    var target world.Player
    if sender.IsPlayer() {
        target = p.manager.GetPlayer(sender.GetName())
    }
    
    if len(args) > 0 {
        target = p.manager.GetPlayer(args[0])
        if target == nil {
            sender.SendMessage("Player not found: " + args[0])
            return true
        }
    }
    
    if target == nil {
        sender.SendMessage("You must specify a player when using from console")
        return true
    }
    
    // This would actually heal the player in a real implementation
    sender.SendMessage(fmt.Sprintf("Healed %s", target.Username))
    return true
}

func (p *EssentialsPlugin) handlePlayerJoin(event plugin.Event) {
    joinEvent := event.(*plugin.PlayerJoinEvent)
    p.api.BroadcastMessage("Welcome " + joinEvent.Player.Username + " to the server!")
}

func (p *EssentialsPlugin) handlePlayerChat(event plugin.Event) {
    chatEvent := event.(*plugin.PlayerChatEvent)
    
    // Example: Capitalize first letter of message
    if len(chatEvent.Message) > 0 {
        chatEvent.Message = string(chatEvent.Message[0]-32) + chatEvent.Message[1:]
    }
}