package config

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
)

type Config struct {
    Address     string `json:"address"`
    Port        int    `json:"port"`
    MaxPlayers  int    `json:"max_players"`
    Motd        string `json:"motd"`
    Version     string `json:"version"`
    WorldName   string `json:"world_name"`
    Seed        int64  `json:"seed"`
    Debug       bool   `json:"debug"`
    
    // RakNet specific settings
    RakNetTimeout int  `json:"raknet_timeout"`
    MTUSize       int  `json:"mtu_size"`
    
    // World settings
    Gamemode      string `json:"gamemode"`
    Difficulty    int    `json:"difficulty"`
    SpawnProtection int `json:"spawn_protection"`
    
    // Game settings
    ViewDistance   int    `json:"view_distance"`
    AllowCheats    bool   `json:"allow_cheats"`
    PvP            bool   `json:"pvp"`
    AllowFlight    bool   `json:"allow_flight"`
    SpawnMonsters  bool   `json:"spawn_monsters"`
    SpawnAnimals   bool   `json:"spawn_animals"`
    OnlineMode     bool   `json:"online_mode"`
    ResourcePacks  []string `json:"resource_packs"`
    Whitelist      bool   `json:"whitelist"`
    MaxWorldSize   int    `json:"max_world_size"`
}

func Load() *Config {
    // Try to load from config.json
    config, err := loadFromFile("config.json")
    if err != nil {
        log.Printf("Could not load config.json: %v", err)
        log.Printf("Using default configuration")
        return defaultConfig()
    }
    
    log.Printf("Configuration loaded from config.json")
    return config
}

func loadFromFile(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("could not read config file: %v", err)
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("could not parse config file: %v", err)
    }
    
    return &config, nil
}

func defaultConfig() *Config {
    return &Config{
        Address:       "0.0.0.0",
        Port:          19132,
        MaxPlayers:    20,
        Motd:          "Garuda Minecraft Server",
        Version:       "1.20.0",
        WorldName:     "garuda-world",
        Seed:          12345,
        Debug:         true,
        RakNetTimeout: 30,
        MTUSize:       1492,
        Gamemode:      "survival",
        Difficulty:    2,
        SpawnProtection: 16,
        ViewDistance:  8,
        AllowCheats:   false,
        PvP:           true,
        AllowFlight:   false,
        SpawnMonsters: true,
        SpawnAnimals:  true,
        OnlineMode:    false,
        ResourcePacks: []string{},
        Whitelist:     false,
        MaxWorldSize:  10000,
    }
}

func (c *Config) Save() error {
    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return fmt.Errorf("could not marshal config: %v", err)
    }
    
    if err := os.WriteFile("config.json", data, 0644); err != nil {
        return fmt.Errorf("could not write config file: %v", err)
    }
    
    return nil
}

// Helper methods
func (c *Config) GetGamemodeID() int32 {
    switch c.Gamemode {
    case "survival":
        return 0
    case "creative":
        return 1
    case "adventure":
        return 2
    case "spectator":
        return 3
    default:
        return 0 // Default to survival
    }
}

func (c *Config) IsValid() bool {
    if c.Port < 1 || c.Port > 65535 {
        return false
    }
    if c.MaxPlayers < 1 {
        return false
    }
    if c.ViewDistance < 2 || c.ViewDistance > 32 {
        return false
    }
    return true
}