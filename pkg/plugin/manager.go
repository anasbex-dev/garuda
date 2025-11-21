package plugin

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "plugin"
    "sync"

    "github.com/anabex-dev/garuda/internal/protocol/minecraft"
    "github.com/anabex-dev/garuda/internal/world"
)

type PluginManager struct {
    plugins     map[string]Plugin
    enabled     map[string]bool
    mutex       sync.RWMutex
    pluginsDir  string
    server      ServerAPI
}

type ServerAPI interface {
    BroadcastMessage(message string)
    GetPlayer(name string) *world.Player
    GetOnlinePlayers() []*world.Player
    ExecuteCommand(command string) bool
    GetWorld() *world.World
}

func NewPluginManager(pluginsDir string, server ServerAPI) *PluginManager {
    return &PluginManager{
        plugins:    make(map[string]Plugin),
        enabled:    make(map[string]bool),
        pluginsDir: pluginsDir,
        server:     server,
    }
}

func (pm *PluginManager) LoadAllPlugins() error {
    // Create plugins directory if it doesn't exist
    if err := os.MkdirAll(pm.pluginsDir, 0755); err != nil {
        return fmt.Errorf("could not create plugins directory: %v", err)
    }

    // Load from compiled .so files
    return filepath.Walk(pm.pluginsDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() || filepath.Ext(path) != ".so" {
            return nil
        }

        return pm.LoadPlugin(path)
    })
}

func (pm *PluginManager) LoadPlugin(path string) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    // Open the plugin file
    plug, err := plugin.Open(path)
    if err != nil {
        return fmt.Errorf("could not open plugin %s: %v", path, err)
    }

    // Look up the symbol "PluginInstance"
    symPlugin, err := plug.Lookup("PluginInstance")
    if err != nil {
        return fmt.Errorf("plugin %s does not export PluginInstance: %v", path, err)
    }

    // Assert that it implements the Plugin interface
    p, ok := symPlugin.(Plugin)
    if !ok {
        return fmt.Errorf("plugin %s does not implement Plugin interface", path)
    }

    pluginName := p.GetName()
    
    // Check if plugin already loaded
    if _, exists := pm.plugins[pluginName]; exists {
        return fmt.Errorf("plugin %s already loaded", pluginName)
    }

    // Store the plugin
    pm.plugins[pluginName] = p
    log.Printf("Loaded plugin: %s v%s by %s", pluginName, p.GetVersion(), p.GetAuthor())

    return nil
}

func (pm *PluginManager) EnablePlugin(name string) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    p, exists := pm.plugins[name]
    if !exists {
        return fmt.Errorf("plugin %s not found", name)
    }

    if pm.enabled[name] {
        return fmt.Errorf("plugin %s already enabled", name)
    }

    // Enable the plugin
    p.OnEnable(pm)
    pm.enabled[name] = true

    log.Printf("Enabled plugin: %s", name)
    return nil
}

func (pm *PluginManager) EnableAllPlugins() {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    for name, p := range pm.plugins {
        if !pm.enabled[name] {
            p.OnEnable(pm)
            pm.enabled[name] = true
            log.Printf("Enabled plugin: %s", name)
        }
    }
}

func (pm *PluginManager) DisablePlugin(name string) error {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    p, exists := pm.plugins[name]
    if !exists {
        return fmt.Errorf("plugin %s not found", name)
    }

    if !pm.enabled[name] {
        return fmt.Errorf("plugin %s not enabled", name)
    }

    // Disable the plugin
    p.OnDisable()
    delete(pm.enabled, name)

    log.Printf("Disabled plugin: %s", name)
    return nil
}

func (pm *PluginManager) DisableAllPlugins() {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    for name, p := range pm.plugins {
        if pm.enabled[name] {
            p.OnDisable()
            delete(pm.enabled, name)
            log.Printf("Disabled plugin: %s", name)
        }
    }
}

func (pm *PluginManager) UnloadPlugin(name string) error {
    if err := pm.DisablePlugin(name); err != nil {
        return err
    }

    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    delete(pm.plugins, name)
    log.Printf("Unloaded plugin: %s", name)
    return nil
}

func (pm *PluginManager) GetPlugin(name string) Plugin {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    return pm.plugins[name]
}

func (pm *PluginManager) GetPlugins() []Plugin {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    plugins := make([]Plugin, 0, len(pm.plugins))
    for _, p := range pm.plugins {
        plugins = append(plugins, p)
    }
    return plugins
}

func (pm *PluginManager) IsEnabled(name string) bool {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    return pm.enabled[name]
}

// Event dispatching methods
func (pm *PluginManager) DispatchPlayerJoin(player *world.Player) {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    for name, p := range pm.plugins {
        if pm.enabled[name] {
            if handler, ok := p.(EventHandler); ok {
                handler.OnPlayerJoin(player)
            }
        }
    }
}

func (pm *PluginManager) DispatchPlayerQuit(player *world.Player) {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    for name, p := range pm.plugins {
        if pm.enabled[name] {
            if handler, ok := p.(EventHandler); ok {
                handler.OnPlayerQuit(player)
            }
        }
    }
}

func (pm *PluginManager) DispatchPlayerChat(player *world.Player, message string) bool {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    allowChat := true
    for name, p := range pm.plugins {
        if pm.enabled[name] {
            if handler, ok := p.(EventHandler); ok {
                if !handler.OnPlayerChat(player, message) {
                    allowChat = false
                }
            }
        }
    }
    return allowChat
}

func (pm *PluginManager) DispatchPlayerMove(player *world.Player, from, to minecraft.Vector3) bool {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    allowMove := true
    for name, p := range pm.plugins {
        if pm.enabled[name] {
            if handler, ok := p.(EventHandler); ok {
                if !handler.OnPlayerMove(player, from, to) {
                    allowMove = false
                }
            }
        }
    }
    return allowMove
}

func (pm *PluginManager) DispatchBlockBreak(player *world.Player, pos minecraft.BlockPos, block world.Block) bool {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    allowBreak := true
    for name, p := range pm.plugins {
        if pm.enabled[name] {
            if handler, ok := p.(EventHandler); ok {
                if !handler.OnBlockBreak(player, pos, block) {
                    allowBreak = false
                }
            }
        }
    }
    return allowBreak
}

func (pm *PluginManager) DispatchBlockPlace(player *world.Player, pos minecraft.BlockPos, block world.Block) bool {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    allowPlace := true
    for name, p := range pm.plugins {
        if pm.enabled[name] {
            if handler, ok := p.(EventHandler); ok {
                if !handler.OnBlockPlace(player, pos, block) {
                    allowPlace = false
                }
            }
        }
    }
    return allowPlace
}

func (pm *PluginManager) DispatchPlayerCommand(player *world.Player, command string, args []string) bool {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    allowCommand := true
    for name, p := range pm.plugins {
        if pm.enabled[name] {
            if handler, ok := p.(EventHandler); ok {
                if !handler.OnPlayerCommand(player, command, args) {
                    allowCommand = false
                }
            }
        }
    }
    return allowCommand
}