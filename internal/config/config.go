package config

import (
    "encoding/json"
    "os"
    "runtime"
    "strings"
    "sync"
)

type ServerConfig struct {
    Address    string `json:"address"`
    MaxPlayers int    `json:"max_players"`
    MOTD       string `json:"motd"`
    Version    string `json:"version"`
}

type WorldConfig struct {
    Name          string `json:"name"`
    Seed          string `json:"seed"`
    Gamemode      string `json:"gamemode"`
    Difficulty    int    `json:"difficulty"`
    ViewDistance  int    `json:"view_distance"`
}

type PerformanceConfig struct {
    MaxEntities    int  `json:"max_entities"`
    MaxChunks      int  `json:"max_chunks"`
    EnablePhysics  bool `json:"enable_physics"`
    EnableRedstone bool `json:"enable_redstone"`
    EnableMobs     bool `json:"enable_mobs"`
    EnableWeather  bool `json:"enable_weather"`
    CompressionLevel int `json:"compression_level"`
}

type Config struct {
    Server      ServerConfig      `json:"server"`
    World       WorldConfig       `json:"world"`
    Performance PerformanceConfig `json:"performance"`
    Debug       bool              `json:"debug"`
    Platform    string            `json:"-"`
    Protocol ProtocolConfig `json: "protocol"`
}

type ProtocolConfig struct {
    Version           string `json:"version"`
    AutoNegotiate     bool   `json:"auto_negotiate"`
    StrictVersionCheck bool  `json:"strict_version_check"`
}

var (
    instance *Config
    once     sync.Once
)

func detectPlatform() string {
    // Detect operating system and environment
    os := runtime.GOOS
    arch := runtime.GOARCH
    
    // Check for Termux
    if _, err := os.Stat("/data/data/com.termux/files/usr"); err == nil {
        return "termux"
    }
    
    // Check for WSL (Windows Subsystem for Linux)
    if os == "linux" {
        if _, err := os.Stat("/mnt/c/Windows"); err == nil {
            return "wsl"
        }
        if _, err := os.Stat("/proc/version"); err == nil {
            if version, err := os.ReadFile("/proc/version"); err == nil {
                if strings.Contains(strings.ToLower(string(version)), "microsoft") {
                    return "wsl"
                }
            }
        }
    }
    
    // Check for Docker container
    if _, err := os.Stat("/.dockerenv"); err == nil {
        return "docker"
    }
    
    // Check for low-memory systems
    if isLowMemorySystem() {
        return "lowmem"
    }
    
    return os + "_" + arch
}

func isLowMemorySystem() bool {
    // Simple memory check (ini hanya estimasi)
    if runtime.GOOS == "linux" {
        if memInfo, err := os.ReadFile("/proc/meminfo"); err == nil {
            memStr := string(memInfo)
            if strings.Contains(memStr, "MemTotal:") {
                // Parse memory total, jika < 2GB consider low memory
                if strings.Contains(memStr, "MemTotal:       1024") || 
                   strings.Contains(memStr, "MemTotal:        512") {
                    return true
                }
            }
        }
    }
    return false
}

func DefaultConfig() *Config {
    platform := detectPlatform()
    
    cfg := &Config{
        Server: ServerConfig{
            Address:    "0.0.0.0:19132",
            MaxPlayers: 20,
            MOTD:       "§bGaruda§f Minecraft Server",
            Version:    "1.21.50", // Default version
        },
        World: WorldConfig{
            Name:          "world",
            Seed:          "garuda",
            Gamemode:      "survival",
            Difficulty:    2,
            ViewDistance:  8,
        },
        Performance: PerformanceConfig{
            MaxEntities:    100,
            MaxChunks:      100,
            EnablePhysics:  true,
            EnableRedstone: false,
            EnableMobs:     true,
            EnableWeather:  true,
            CompressionLevel: 1,
        },
        Debug:    true,
        Platform: platform,
        Protocol: ProtocolConfig{
            Version:           "1.21.50",
            AutoNegotiate:     true,
            StrictVersionCheck: false,
        },
    }
    
    applyPlatformOptimizations(cfg, platform)
    return cfg
}

func applyPlatformOptimizations(cfg *Config, platform string) {
    switch platform {
    case "termux":
        cfg.Server.MaxPlayers = 5
        cfg.World.ViewDistance = 4
        cfg.Performance.MaxEntities = 20
        cfg.Performance.MaxChunks = 25
        cfg.Performance.EnableRedstone = false
        cfg.Performance.EnableWeather = false
        cfg.Performance.CompressionLevel = 0
        
    case "wsl":
        cfg.World.ViewDistance = 6
        cfg.Performance.MaxEntities = 50
        cfg.Performance.EnableRedstone = true
        
    case "docker":
        cfg.Server.MaxPlayers = 10
        cfg.World.ViewDistance = 6
        cfg.Performance.MaxEntities = 50
        
    case "lowmem":
        cfg.Server.MaxPlayers = 8
        cfg.World.ViewDistance = 4
        cfg.Performance.MaxEntities = 30
        cfg.Performance.MaxChunks = 50
        cfg.Performance.EnableRedstone = false
        cfg.Performance.EnableWeather = false
        
    case "linux_arm", "linux_arm64":
        // Raspberry Pi dan ARM devices
        cfg.Server.MaxPlayers = 8
        cfg.World.ViewDistance = 6
        cfg.Performance.MaxEntities = 40
        cfg.Performance.EnableRedstone = false
        
    case "windows_386", "linux_386":
        // 32-bit systems
        cfg.Server.MaxPlayers = 8
        cfg.World.ViewDistance = 6
        cfg.Performance.MaxEntities = 30
        cfg.Performance.MaxChunks = 50
    }
}

func Load(configPath string) (*Config, error) {
    var cfg *Config
    
    if configPath == "" {
        configPath = getDefaultConfigPath()
    }
    
    file, err := os.Open(configPath)
    if err != nil {
        if os.IsNotExist(err) {
            cfg = DefaultConfig()
            if err := Save(configPath, cfg); err != nil {
                return nil, err
            }
            return cfg, nil
        }
        return nil, err
    }
    defer file.Close()
    
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&cfg); err != nil {
        return nil, err
    }
    
    // Set platform detection
    cfg.Platform = detectPlatform()
    
    return cfg, nil
}

func Save(configPath string, cfg *Config) error {
    if configPath == "" {
        configPath = getDefaultConfigPath()
    }
    
    // Ensure directory exists
    if err := os.MkdirAll(getConfigDir(), 0755); err != nil {
        return err
    }
    
    file, err := os.Create(configPath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(cfg)
}

func getDefaultConfigPath() string {
    configDir := getConfigDir()
    return configDir + "/config.json"
}

func getConfigDir() string {
    // XDG Base Directory compliant
    if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
        return xdgConfig + "/garuda"
    }
    
    // Platform-specific default paths
    switch runtime.GOOS {
    case "windows":
        return os.Getenv("APPDATA") + "\\Garuda"
    case "darwin":
        return os.Getenv("HOME") + "/Library/Application Support/Garuda"
    default: // Linux, BSD, etc.
        // Check for Termux first
        if _, err := os.Stat("/data/data/com.termux"); err == nil {
            return "/data/data/com.termux/files/home/.garuda"
        }
        return os.Getenv("HOME") + "/.config/garuda"
    }
}

func (c *Config) IsLowEndDevice() bool {
    return c.Platform == "termux" || c.Platform == "lowmem" || 
           strings.Contains(c.Platform, "arm") || 
           strings.Contains(c.Platform, "386")
}

func (c *Config) GetOptimalThreadCount() int {
    // Jangan gunakan semua cores di low-end devices
    availableCores := runtime.NumCPU()
    
    if c.IsLowEndDevice() {
        if availableCores > 2 {
            return 2
        }
        return 1
    }
    
    // Untuk high-end systems, gunakan semua cores minus 1 untuk system
    if availableCores > 4 {
        return availableCores - 1
    }
    
    return availableCores
}