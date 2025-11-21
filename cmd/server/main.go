package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "path/filepath"
    "runtime"
    "runtime/pprof"
    "syscall"
    "time"

    "github.com/anabex-dev/garuda/internal/config"
    "github.com/anabex-dev/garuda/internal/network/raknet"
    "github.com/anabex-dev/garuda/internal/server"
    "github.com/anabex-dev/garuda/internal/world"
    "github.com/anabex-dev/garuda/pkg/plugin"
    "github.com/anabex-dev/garuda/api"
)

var (
    version   = "1.0.0"
    buildTime = "unknown"
    gitCommit = "unknown"
)

// Server instance global variable
var garudaServer *raknet.RakNetServer

func main() {
    // Parse command line flags
    
     if err := apiServer.Start(8080); err != nil {
        log.Printf("Failed to start API server: %v", err)
    } else {
        log.Printf("REST API available at http://localhost:8080")
    }
    var (
        configFile    = flag.String("config", "config.json", "Configuration file path")
        versionFlag   = flag.Bool("version", false, "Show version information")
        helpFlag      = flag.Bool("help", false, "Show help message")
        cpuProfile    = flag.String("cpuprofile", "", "Write CPU profile to file")
        memProfile    = flag.String("memprofile", "", "Write memory profile to file")
        debugFlag     = flag.Bool("debug", false, "Enable debug mode")
        offlineMode   = flag.Bool("offline", false, "Run in offline mode")
        worldName     = flag.String("world", "", "World name to load")
        port          = flag.Int("port", 0, "Server port")
        maxPlayers    = flag.Int("maxplayers", 0, "Maximum players")
    )
    flag.Parse()

    // Show version information
    if *versionFlag {
        showVersion()
        return
    }

    // Show help
    if *helpFlag {
        showHelp()
        return
    }

    // Start CPU profiling if enabled
    if *cpuProfile != "" {
        f, err := os.Create(*cpuProfile)
        if err != nil {
            log.Fatal("Could not create CPU profile: ", err)
        }
        defer f.Close()
        
        if err := pprof.StartCPUProfile(f); err != nil {
            log.Fatal("Could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
        
        log.Printf("CPU profiling enabled, output to: %s", *cpuProfile)
    }

    // Setup signal handling early
    setupSignalHandling()

    // Initialize logging
    initLogging()

    // Show startup banner
    showBanner()

    // Load configuration
    cfg := loadConfiguration(*configFile, *debugFlag, *offlineMode, *port, *maxPlayers, *worldName)

    // Validate configuration
    if !cfg.IsValid() {
        log.Fatalf("Invalid configuration detected. Please check your configuration file.")
    }

    // Create necessary directories
    createDirectories(cfg)

    // Initialize server components
    worldInstance := initializeWorld(cfg)
    serverAPI := initializeServerAPI(worldInstance)
    pluginManager := initializePluginManager(cfg, serverAPI)

    // Create and start RakNet server
    garudaServer = initializeRakNetServer(cfg, worldInstance, pluginManager, serverAPI)

    // Start the server
    if err := startServer(garudaServer, cfg); err != nil {
        log.Fatalf("Failed to start Garuda server: %v", err)
    }

    // Write memory profile if enabled
    if *memProfile != "" {
        f, err := os.Create(*memProfile)
        if err != nil {
            log.Fatal("Could not create memory profile: ", err)
        }
        defer f.Close()
        
        runtime.GC() // Run garbage collection to get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            log.Fatal("Could not write memory profile: ", err)
        }
        
        log.Printf("Memory profiling completed, output to: %s", *memProfile)
    }

    log.Println("Garuda server has shut down successfully")
}

func showVersion() {
    fmt.Printf("Garuda Minecraft Server\n")
    fmt.Printf("Version: %s\n", version)
    fmt.Printf("Build Time: %s\n", buildTime)
    fmt.Printf("Git Commit: %s\n", gitCommit)
    fmt.Printf("Go Version: %s\n", runtime.Version())
    fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func showHelp() {
    fmt.Printf("Garuda Minecraft Server - Usage\n\n")
    fmt.Printf("Flags:\n")
    flag.PrintDefaults()
    fmt.Printf("\nExamples:\n")
    fmt.Printf("  %s                            # Start with default config\n", os.Args[0])
    fmt.Printf("  %s -config custom.json        # Use custom config file\n", os.Args[0])
    fmt.Printf("  %s -port 25565 -offline       # Custom port in offline mode\n", os.Args[0])
    fmt.Printf("  %s -debug -cpuprofile cpu.pprof  # Debug mode with profiling\n", os.Args[0])
}

func showBanner() {
    banner := `
    ╔═══════════════════════════════════════╗
    ║            • GARUDA MC •           ║
    ║    High-Performance Minecraft Server 
    ║   © AnasBex Development - 2025
    ║           Version: %-10s       ║
    ╚═══════════════════════════════════════╝
    `
    fmt.Printf(banner+"\n", version)
    log.Printf("Starting Garuda Minecraft Server v%s...", version)
    log.Printf("Build: %s (%s)", gitCommit, buildTime)
    log.Printf("Runtime: %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func initLogging() {
    // Create logs directory if it doesn't exist
    if err := os.MkdirAll("logs", 0755); err != nil {
        log.Printf("Warning: Could not create logs directory: %v", err)
        return
    }

    // Setup log file with timestamp
    logFileName := fmt.Sprintf("logs/garuda-%s.log", time.Now().Format("2006-01-02-15-04-05"))
    logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Printf("Warning: Could not open log file: %v", err)
        return
    }

    // Set output to both console and file
    log.SetOutput(logFile)
    log.Printf("Logging initialized, output to: %s", logFileName)
}

func loadConfiguration(configFile string, debug, offline bool, port, maxPlayers int, worldName string) *config.Config {
    var cfg *config.Config

    // Check if config file exists
    if _, err := os.Stat(configFile); err == nil {
        // Load from specified config file
        cfg = config.LoadConfig()
        log.Printf("Configuration loaded from: %s", configFile)
    } else {
        // Use default configuration
        cfg = config.DefaultConfig()
        log.Printf("Using default configuration")
    }

    // Override with command line flags
    if debug {
        cfg.Debug = true
        log.Printf("Debug mode enabled via command line")
    }
    if offline {
        cfg.OnlineMode = false
        log.Printf("Offline mode enabled via command line")
    }
    if port > 0 {
        cfg.Port = port
        log.Printf("Port set to %d via command line", port)
    }
    if maxPlayers > 0 {
        cfg.MaxPlayers = maxPlayers
        log.Printf("Max players set to %d via command line", maxPlayers)
    }
    if worldName != "" {
        cfg.WorldName = worldName
        log.Printf("World name set to %s via command line", worldName)
    }

    return cfg
}

func createDirectories(cfg *config.Config) {
    directories := []string{
        "worlds",
        cfg.GetWorldPath(),
        cfg.GetPluginPath(),
        "logs",
        "backups",
    }

    for _, dir := range directories {
        if err := os.MkdirAll(dir, 0755); err != nil {
            log.Printf("Warning: Could not create directory %s: %v", dir, err)
        } else {
            log.Printf("Directory ensured: %s", dir)
        }
    }
}

func initializeWorld(cfg *config.Config) *world.World {
    worldPath := cfg.GetWorldPath()
    
    // Check if world exists
    if _, err := os.Stat(worldPath); os.IsNotExist(err) {
        log.Printf("Creating new world: %s (Seed: %d)", cfg.WorldName, cfg.Seed)
    } else {
        log.Printf("Loading existing world: %s", cfg.WorldName)
    }

    worldInstance := world.NewWorld(cfg.WorldName, cfg.Seed)
    
    // Apply configuration to world
    worldInstance.SetSpawnPoint(cfg.GetSpawnPoint())
    worldInstance.SetDifficulty(cfg.Difficulty)
    
    // Set game rules from config
    for name, value := range cfg.GameRules {
        worldInstance.SetGameRule(name, value)
    }

    log.Printf("World initialized: %s", cfg.WorldName)
    return worldInstance
}

func initializeServerAPI(worldInstance *world.World) *server.GarudaServer {
    serverAPI := server.NewGarudaServer(worldInstance)
    log.Printf("Server API initialized")
    return serverAPI
}

func initializePluginManager(cfg *config.Config, serverAPI *server.GarudaServer) *plugin.PluginManager {
    if !cfg.EnablePlugins {
        log.Printf("Plugins are disabled")
        return nil
    }

    pluginManager := plugin.NewPluginManager(cfg.GetPluginPath(), serverAPI)
    
    // Load plugins
    if err := pluginManager.LoadAllPlugins(); err != nil {
        log.Printf("Warning: Could not load plugins: %v", err)
    } else {
        log.Printf("Plugin manager initialized, loaded %d plugins", len(pluginManager.GetPlugins()))
    }

    return pluginManager
}

func initializeRakNetServer(cfg *config.Config, worldInstance *world.World, pluginManager *plugin.PluginManager, serverAPI *server.GarudaServer) *raknet.RakNetServer {
    server := raknet.NewServer(cfg.Address, cfg.Port)
    
    // Set server properties
    server.SetWorld(worldInstance)
    server.SetPluginManager(pluginManager)
    server.SetServerAPI(serverAPI)
    server.SetMaxPlayers(cfg.MaxPlayers)
    
    log.Printf("RakNet server initialized on %s:%d", cfg.Address, cfg.Port)
    return server
}

func startServer(server *raknet.RakNetServer, cfg *config.Config) error {
    // Display server information
    log.Printf("╔═══════════════════════════════════════╗")
    log.Printf("║           Server Starting...         ║")
    log.Printf("╠═══════════════════════════════════════╣")
    log.Printf("║ Address: %-28s ║", cfg.GetFullAddress())
    log.Printf("║ MOTD: %-30s ║", cfg.Motd)
    log.Printf("║ Version: %-27s ║", cfg.Version)
    log.Printf("║ Max Players: %-23d ║", cfg.MaxPlayers)
    log.Printf("║ World: %-29s ║", cfg.WorldName)
    log.Printf("║ Gamemode: %-26s ║", cfg.GetGamemodeName())
    log.Printf("║ Difficulty: %-24s ║", cfg.GetDifficultyName())
    log.Printf("║ Online Mode: %-23t ║", cfg.OnlineMode)
    log.Printf("║ View Distance: %-21d ║", cfg.ViewDistance)
    log.Printf("╚═══════════════════════════════════════╝")

    // Performance information
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    log.Printf("Memory: Alloc=%.2fMB, Sys=%.2fMB", 
        float64(m.Alloc)/1024/1024, 
        float64(m.Sys)/1024/1024)

    // Start the server
    startTime := time.Now()
    if err := server.Start(); err != nil {
        return fmt.Errorf("failed to start server: %v", err)
    }

    log.Printf("Server started successfully in %v", time.Since(startTime))
    log.Printf("Ready for connections! Use 'minecraft://%s' to connect", cfg.GetFullAddress())

    // Server main loop
    serverMainLoop(server)

    return nil
}

func serverMainLoop(server *raknet.RakNetServer) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Periodic server maintenance
            serverMaintenance(server)
            
        case <-shutdownChan:
            // Graceful shutdown
            log.Println("Initiating graceful shutdown...")
            server.Stop()
            return
        }
    }
}

func serverMaintenance(server *raknet.RakNetServer) {
    // Performance monitoring
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    playerCount := server.GetPlayerCount()
    log.Printf("Status: %d players online | Memory: %.2fMB | Goroutines: %d",
        playerCount, float64(m.Alloc)/1024/1024, runtime.NumGoroutine())

    // Auto-save if enabled
    if time.Now().Unix()%300 == 0 { // Every 5 minutes
        log.Println("Performing auto-save...")
        // TODO: Implement world auto-save
    }
}

// Signal handling
var shutdownChan = make(chan os.Signal, 1)

func setupSignalHandling() {
    signal.Notify(shutdownChan, 
        syscall.SIGINT,  // Ctrl+C
        syscall.SIGTERM, // Kubernetes/Systemd stop
        syscall.SIGQUIT, // Ctrl+\
        syscall.SIGHUP,  // Terminal closed
    )

    go func() {
        sig := <-shutdownChan
        log.Printf("Received signal: %v", sig)
        
        switch sig {
        case syscall.SIGINT:
            log.Println("Shutdown initiated by Ctrl+C")
        case syscall.SIGTERM:
            log.Println("Shutdown initiated by system")
        case syscall.SIGQUIT:
            log.Println("Shutdown initiated by Ctrl+\\")
        case syscall.SIGHUP:
            log.Println("Terminal disconnected, shutting down...")
        }
        
        // Perform graceful shutdown
        gracefulShutdown()
    }()
}

func gracefulShutdown() {
    log.Println("Starting graceful shutdown sequence...")
    
    // Stop accepting new connections
    if garudaServer != nil {
        garudaServer.Stop()
    }
    
    // Save world state
    log.Println("Saving world data...")
    // TODO: Implement world saving
    
    // Close plugin manager
    log.Println("Closing plugins...")
    // TODO: Close plugin manager
    
    log.Println("Graceful shutdown completed")
    os.Exit(0)
}

// Utility function to get executable name
func getExecutableName() string {
    exe, err := os.Executable()
    if err != nil {
        return "garuda"
    }
    return filepath.Base(exe)
}

// DefaultConfig provides default configuration (helper function)
func DefaultConfig() *config.Config {
    return config.DefaultConfig()
}