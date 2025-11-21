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
    
    cfg := config.Load()
    
    // Hanya pakai RakNet server
    server := raknet.NewServer(cfg.Address, cfg.Port)
    
    setupSignalHandling(server)
    
    log.Printf("Garuda RakNet server starting on %s:%d", cfg.Address, cfg.Port)
    log.Printf("Version: %s | Max Players: %d", cfg.Version, cfg.MaxPlayers)
    log.Printf("World: %s", cfg.WorldName)
    
    if err := server.Start(); err != nil {
        log.Fatalf("Failed to start Garuda server: %v", err)
    }
}

func setupSignalHandling(server *raknet.RakNetServer) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Received shutdown signal...")
        server.Stop()
        os.Exit(0)
    }()
}