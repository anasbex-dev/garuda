package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/anabex-dev/garuda/internal/config"
    "github.com/anabex-dev/garuda/internal/network/raknet"
)

func main() {
    log.Println("Starting Garuda Minecraft Server...")
    
    // Load configuration
    cfg := config.Load()
    
    // Validate configuration
    if !cfg.IsValid() {
        log.Fatalf("Invalid configuration detected. Please check your config.json")
    }
    
    // Create default config file if it doesn't exist
    if _, err := os.Stat("config.json"); os.IsNotExist(err) {
        log.Println("Creating default config.json...")
        if err := cfg.Save(); err != nil {
            log.Printf("Warning: Could not create config.json: %v", err)
        }
    }
    
    // Create RakNet server
    server := raknet.NewServer(cfg.Address, cfg.Port)
    
    // Setup signal handling for graceful shutdown
    setupSignalHandling(server)
    
    log.Printf("Garuda RakNet server starting on %s:%d", cfg.Address, cfg.Port)
    log.Printf("Version: %s | Max Players: %d", cfg.Version, cfg.MaxPlayers)
    log.Printf("World: %s | Seed: %d", cfg.WorldName, cfg.Seed)
    log.Printf("Gamemode: %s | Difficulty: %d", cfg.Gamemode, cfg.Difficulty)
    log.Printf("View Distance: %d", cfg.ViewDistance)
    
    if cfg.OnlineMode {
        log.Printf("Server running in ONLINE mode")
    } else {
        log.Printf("Server running in OFFLINE mode")
    }
    
    if err := server.Start(); err != nil {
        log.Fatalf("Failed to start Garuda server: %v", err)
    }
}

func setupSignalHandling(server *raknet.RakNetServer) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    
    go func() {
        sig := <-sigChan
        log.Printf("Received signal: %v", sig)
        log.Println("Shutting down Garuda server gracefully...")
        
        server.Stop()
        log.Println("Garuda server stopped successfully")
        os.Exit(0)
    }()
}