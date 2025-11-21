package config

type Config struct {
    Address     string `json:"address"`
    Port        int    `json:"port"`
    MaxPlayers  int    `json:"max_players"`
    Motd        string `json:"motd"`
    Version     string `json:"version"`
    WorldName   string `json:"world_name"`
    Debug       bool   `json:"debug"`
    // RakNet specific settings
    RakNetTimeout int  `json:"raknet_timeout"`
    MTUSize       int  `json:"mtu_size"`
}

func Load() *Config {
    return &Config{
        Address:       "0.0.0.0",
        Port:          19132,
        MaxPlayers:    20,
        Motd:          "Garuda Minecraft Server",
        Version:       "1.0.0",
        WorldName:     "garuda-world",
        Debug:         true,
        RakNetTimeout: 30,
        MTUSize:       1492,
    }
}