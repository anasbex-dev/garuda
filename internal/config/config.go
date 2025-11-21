package config

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

type Config struct {
    // Network settings
    Address     string `json:"address"`
    Port        int    `json:"port"`
    MaxPlayers  int    `json:"max_players"`
    Motd        string `json:"motd"`
    Version     string `json:"version"`
    
    // World settings
    WorldName   string `json:"world_name"`
    Seed        int64  `json:"seed"`
    Gamemode    string `json:"gamemode"`
    Difficulty  int    `json:"difficulty"`
    ViewDistance int   `json:"view_distance"`
    SpawnProtection int `json:"spawn_protection"`
    MaxWorldSize int   `json:"max_world_size"`
    
    // Server settings
    Debug          bool `json:"debug"`
    OnlineMode     bool `json:"online_mode"`
    Whitelist      bool `json:"whitelist"`
    AllowCheats    bool `json:"allow_cheats"`
    PvP            bool `json:"pvp"`
    AllowFlight    bool `json:"allow_flight"`
    SpawnMonsters  bool `json:"spawn_monsters"`
    SpawnAnimals   bool `json:"spawn_animals"`
    Hardcore       bool `json:"hardcore"`
    
    // RakNet specific settings
    RakNetTimeout int  `json:"raknet_timeout"`
    MTUSize       int  `json:"mtu_size"`
    CompressionThreshold int `json:"compression_threshold"`
    
    // Resource packs
    ResourcePacks  []string `json:"resource_packs"`
    ForceResources bool     `json:"force_resources"`
    
    // Performance settings
    ChunkTickRadius int `json:"chunk_tick_radius"`
    MaxChunksPerTick int `json:"max_chunks_per_tick"`
    PlayerIdleTimeout int `json:"player_idle_timeout"`
    
    // Plugin settings
    PluginFolder string `json:"plugin_folder"`
    EnablePlugins bool  `json:"enable_plugins"`
    
    // Advanced settings
    LevelType      string `json:"level_type"`
    Generator      string `json:"generator"`
    AutoSave       bool   `json:"auto_save"`
    AutoSaveInterval int  `json:"auto_save_interval"`
    QueryEnabled   bool   `json:"query_enabled"`
    QueryPort      int    `json:"query_port"`
    
    // Game rules
    GameRules map[string]interface{} `json:"game_rules"`
}

type ServerProperties struct {
    Config
    WhitelistPlayers []string `json:"whitelist_players"`
    Ops              []string `json:"ops"`
    BannedPlayers    []string `json:"banned_players"`
    BannedIPs        []string `json:"banned_ips"`
}

func Load() *Config {
    // Try to load from config.json first
    config, err := loadFromFile("config.json")
    if err != nil {
        log.Printf("Could not load config.json: %v", err)
        
        // Try to load from server.properties (Minecraft compatible)
        config, err = loadFromProperties("server.properties")
        if err != nil {
            log.Printf("Could not load server.properties: %v", err)
            log.Printf("Using default configuration")
            config = defaultConfig()
            
            // Create default config file
            if err := config.Save(); err != nil {
                log.Printf("Warning: Could not create config file: %v", err)
            }
        }
    }
    
    // Validate configuration
    if !config.IsValid() {
        log.Printf("Invalid configuration detected, using defaults")
        config = defaultConfig()
    }
    
    log.Printf("Configuration loaded successfully")
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

func loadFromProperties(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("could not read properties file: %v", err)
    }
    
    props := parseProperties(string(data))
    config := propertiesToConfig(props)
    
    return config, nil
}

func parseProperties(data string) map[string]string {
    props := make(map[string]string)
    lines := strings.Split(data, "\n")
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        
        parts := strings.SplitN(line, "=", 2)
        if len(parts) == 2 {
            key := strings.TrimSpace(parts[0])
            value := strings.TrimSpace(parts[1])
            props[key] = value
        }
    }
    
    return props
}

func propertiesToConfig(props map[string]string) *Config {
    config := defaultConfig()
    
    // Network settings
    if val, ok := props["server-ip"]; ok {
        config.Address = val
    }
    if val, ok := props["server-port"]; ok {
        if port, err := strconv.Atoi(val); err == nil {
            config.Port = port
        }
    }
    if val, ok := props["max-players"]; ok {
        if max, err := strconv.Atoi(val); err == nil {
            config.MaxPlayers = max
        }
    }
    if val, ok := props["motd"]; ok {
        config.Motd = val
    }
    
    // World settings
    if val, ok := props["level-name"]; ok {
        config.WorldName = val
    }
    if val, ok := props["level-seed"]; ok {
        if seed, err := strconv.ParseInt(val, 10, 64); err == nil {
            config.Seed = seed
        }
    }
    if val, ok := props["gamemode"]; ok {
        config.Gamemode = val
    }
    if val, ok := props["difficulty"]; ok {
        if diff, err := strconv.Atoi(val); err == nil {
            config.Difficulty = diff
        }
    }
    if val, ok := props["view-distance"]; ok {
        if dist, err := strconv.Atoi(val); err == nil {
            config.ViewDistance = dist
        }
    }
    if val, ok := props["spawn-protection"]; ok {
        if prot, err := strconv.Atoi(val); err == nil {
            config.SpawnProtection = prot
        }
    }
    
    // Server settings
    if val, ok := props["online-mode"]; ok {
        config.OnlineMode = strings.ToLower(val) == "true"
    }
    if val, ok := props["white-list"]; ok {
        config.Whitelist = strings.ToLower(val) == "true"
    }
    if val, ok := props["allow-cheats"]; ok {
        config.AllowCheats = strings.ToLower(val) == "true"
    }
    if val, ok := props["pvp"]; ok {
        config.PvP = strings.ToLower(val) == "true"
    }
    if val, ok := props["allow-flight"]; ok {
        config.AllowFlight = strings.ToLower(val) == "true"
    }
    if val, ok := props["spawn-monsters"]; ok {
        config.SpawnMonsters = strings.ToLower(val) == "true"
    }
    if val, ok := props["spawn-animals"]; ok {
        config.SpawnAnimals = strings.ToLower(val) == "true"
    }
    if val, ok := props["hardcore"]; ok {
        config.Hardcore = strings.ToLower(val) == "true"
    }
    
    // Advanced settings
    if val, ok := props["level-type"]; ok {
        config.LevelType = val
    }
    if val, ok := props["generator-settings"]; ok {
        config.Generator = val
    }
    if val, ok := props["enable-query"]; ok {
        config.QueryEnabled = strings.ToLower(val) == "true"
    }
    if val, ok := props["query.port"]; ok {
        if port, err := strconv.Atoi(val); err == nil {
            config.QueryPort = port
        }
    }
    
    return config
}

func defaultConfig() *Config {
    return &Config{
        // Network settings
        Address:    "0.0.0.0",
        Port:       19132,
        MaxPlayers: 20,
        Motd:       "Garuda Minecraft Server",
        Version:    "1.20.0",
        
        // World settings
        WorldName:       "garuda-world",
        Seed:            12345,
        Gamemode:        "survival",
        Difficulty:      2, // Normal
        ViewDistance:    8,
        SpawnProtection: 16,
        MaxWorldSize:    10000,
        
        // Server settings
        Debug:         false,
        OnlineMode:    false,
        Whitelist:     false,
        AllowCheats:   false,
        PvP:           true,
        AllowFlight:   false,
        SpawnMonsters: true,
        SpawnAnimals:  true,
        Hardcore:      false,
        
        // RakNet settings
        RakNetTimeout:       30,
        MTUSize:            1492,
        CompressionThreshold: 256,
        
        // Resource packs
        ResourcePacks:  []string{},
        ForceResources: false,
        
        // Performance settings
        ChunkTickRadius:     4,
        MaxChunksPerTick:    10,
        PlayerIdleTimeout:   600, // 10 minutes
        
        // Plugin settings
        PluginFolder:  "plugins",
        EnablePlugins: true,
        
        // Advanced settings
        LevelType:        "DEFAULT",
        Generator:        "flat",
        AutoSave:         true,
        AutoSaveInterval: 600, // 10 minutes
        QueryEnabled:     false,
        QueryPort:        19133,
        
        // Game rules
        GameRules: map[string]interface{}{
            "doDaylightCycle":        true,
            "doWeatherCycle":         true,
            "doMobSpawning":          true,
            "doFireTick":             true,
            "mobGriefing":            true,
            "keepInventory":          false,
            "naturalRegeneration":    true,
            "doTileDrops":            true,
            "doEntityDrops":          true,
            "commandBlockOutput":     true,
            "sendCommandFeedback":    true,
            "maxCommandChainLength":  65536,
            "doInsomnia":             true,
            "commandBlocksEnabled":   true,
            "randomTickSpeed":        3,
            "doImmediateRespawn":     false,
            "showDeathMessages":      true,
            "functionCommandLimit":   10000,
            "spawnRadius":            10,
        },
    }
}

func (c *Config) Save() error {
    // Create config directory if it doesn't exist
    if err := os.MkdirAll(filepath.Dir("config.json"), 0755); err != nil {
        return fmt.Errorf("could not create config directory: %v", err)
    }
    
    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return fmt.Errorf("could not marshal config: %v", err)
    }
    
    if err := os.WriteFile("config.json", data, 0644); err != nil {
        return fmt.Errorf("could not write config file: %v", err)
    }
    
    // Also create server.properties for compatibility
    if err := c.SaveProperties(); err != nil {
        log.Printf("Warning: Could not create server.properties: %v", err)
    }
    
    log.Printf("Configuration saved to config.json")
    return nil
}

func (c *Config) SaveProperties() error {
    var props strings.Builder
    
    props.WriteString("#Minecraft server properties\n")
    props.WriteString("#Garuda Server - " + c.Version + "\n")
    props.WriteString("#" + time.Now().Format("Mon Jan 2 15:04:05 MST 2006") + "\n\n")
    
    // Network settings
    props.WriteString(fmt.Sprintf("server-ip=%s\n", c.Address))
    props.WriteString(fmt.Sprintf("server-port=%d\n", c.Port))
    props.WriteString(fmt.Sprintf("max-players=%d\n", c.MaxPlayers))
    props.WriteString(fmt.Sprintf("motd=%s\n", c.Motd))
    
    // World settings
    props.WriteString(fmt.Sprintf("level-name=%s\n", c.WorldName))
    props.WriteString(fmt.Sprintf("level-seed=%d\n", c.Seed))
    props.WriteString(fmt.Sprintf("gamemode=%s\n", c.Gamemode))
    props.WriteString(fmt.Sprintf("difficulty=%d\n", c.Difficulty))
    props.WriteString(fmt.Sprintf("view-distance=%d\n", c.ViewDistance))
    props.WriteString(fmt.Sprintf("spawn-protection=%d\n", c.SpawnProtection))
    
    // Server settings
    props.WriteString(fmt.Sprintf("online-mode=%t\n", c.OnlineMode))
    props.WriteString(fmt.Sprintf("white-list=%t\n", c.Whitelist))
    props.WriteString(fmt.Sprintf("allow-cheats=%t\n", c.AllowCheats))
    props.WriteString(fmt.Sprintf("pvp=%t\n", c.PvP))
    props.WriteString(fmt.Sprintf("allow-flight=%t\n", c.AllowFlight))
    props.WriteString(fmt.Sprintf("spawn-monsters=%t\n", c.SpawnMonsters))
    props.WriteString(fmt.Sprintf("spawn-animals=%t\n", c.SpawnAnimals))
    props.WriteString(fmt.Sprintf("hardcore=%t\n", c.Hardcore))
    
    // Advanced settings
    props.WriteString(fmt.Sprintf("level-type=%s\n", c.LevelType))
    props.WriteString(fmt.Sprintf("generator-settings=%s\n", c.Generator))
    props.WriteString(fmt.Sprintf("enable-query=%t\n", c.QueryEnabled))
    props.WriteString(fmt.Sprintf("query.port=%d\n", c.QueryPort))
    
    if err := os.WriteFile("server.properties", []byte(props.String()), 0644); err != nil {
        return fmt.Errorf("could not write properties file: %v", err)
    }
    
    return nil
}

// Validation methods
func (c *Config) IsValid() bool {
    if c.Port < 1 || c.Port > 65535 {
        log.Printf("Invalid port: %d", c.Port)
        return false
    }
    if c.MaxPlayers < 1 || c.MaxPlayers > 1000 {
        log.Printf("Invalid max players: %d", c.MaxPlayers)
        return false
    }
    if c.ViewDistance < 2 || c.ViewDistance > 32 {
        log.Printf("Invalid view distance: %d", c.ViewDistance)
        return false
    }
    if c.Difficulty < 0 || c.Difficulty > 3 {
        log.Printf("Invalid difficulty: %d", c.Difficulty)
        return false
    }
    if c.SpawnProtection < 0 {
        log.Printf("Invalid spawn protection: %d", c.SpawnProtection)
        return false
    }
    
    return true
}

// Helper methods
func (c *Config) GetGamemodeID() int32 {
    switch strings.ToLower(c.Gamemode) {
    case "survival", "0":
        return 0
    case "creative", "1":
        return 1
    case "adventure", "2":
        return 2
    case "spectator", "3":
        return 3
    default:
        return 0 // Default to survival
    }
}

func (c *Config) GetGamemodeName() string {
    switch c.GetGamemodeID() {
    case 0:
        return "Survival"
    case 1:
        return "Creative"
    case 2:
        return "Adventure"
    case 3:
        return "Spectator"
    default:
        return "Survival"
    }
}

func (c *Config) GetDifficultyName() string {
    switch c.Difficulty {
    case 0:
        return "Peaceful"
    case 1:
        return "Easy"
    case 2:
        return "Normal"
    case 3:
        return "Hard"
    default:
        return "Normal"
    }
}

func (c *Config) GetGameRule(name string) interface{} {
    if val, exists := c.GameRules[name]; exists {
        return val
    }
    return nil
}

func (c *Config) SetGameRule(name string, value interface{}) {
    c.GameRules[name] = value
}

func (c *Config) GetGameRuleBool(name string) bool {
    if val := c.GetGameRule(name); val != nil {
        if b, ok := val.(bool); ok {
            return b
        }
    }
    return false
}

func (c *Config) GetGameRuleInt(name string) int {
    if val := c.GetGameRule(name); val != nil {
        if i, ok := val.(int); ok {
            return i
        }
        if f, ok := val.(float64); ok {
            return int(f)
        }
    }
    return 0
}

// Network address helper
func (c *Config) GetFullAddress() string {
    return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

// World path helper
func (c *Config) GetWorldPath() string {
    return filepath.Join("worlds", c.WorldName)
}

// Plugin path helper
func (c *Config) GetPluginPath() string {
    return c.PluginFolder
}

// Config info for logging
func (c *Config) GetInfo() string {
    return fmt.Sprintf("Server: %s | Port: %d | Players: %d | World: %s | Gamemode: %s",
        c.Motd, c.Port, c.MaxPlayers, c.WorldName, c.GetGamemodeName())
}

// Environment variable support
func (c *Config) LoadFromEnv() {
    // Override config with environment variables if set
    if val := os.Getenv("GARUDA_ADDRESS"); val != "" {
        c.Address = val
    }
    if val := os.Getenv("GARUDA_PORT"); val != "" {
        if port, err := strconv.Atoi(val); err == nil {
            c.Port = port
        }
    }
    if val := os.Getenv("GARUDA_MAX_PLAYERS"); val != "" {
        if max, err := strconv.Atoi(val); err == nil {
            c.MaxPlayers = max
        }
    }
    if val := os.Getenv("GARUDA_MOTD"); val != "" {
        c.Motd = val
    }
    if val := os.Getenv("GARUDA_WORLD_NAME"); val != "" {
        c.WorldName = val
    }
    if val := os.Getenv("GARUDA_SEED"); val != "" {
        if seed, err := strconv.ParseInt(val, 10, 64); err == nil {
            c.Seed = seed
        }
    }
    if val := os.Getenv("GARUDA_GAMEMODE"); val != "" {
        c.Gamemode = val
    }
    if val := os.Getenv("GARUDA_ONLINE_MODE"); val != "" {
        c.OnlineMode = strings.ToLower(val) == "true"
    }
}

// Update main Load function to include environment variables
func LoadConfig() *Config {
    config := Load()
    config.LoadFromEnv()
    return config
}