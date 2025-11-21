package main

import (
    "fmt"
    "garuda/internal/config"
    "garuda/internal/network/raknet"
    "garuda/minecraft/server"
    "garuda/pkg/utils"
    "log"
    "os"
    "os/signal"
    "runtime"
    "runtime/debug"
    "syscall"
    "time"
)

func main() {
    // Display awesome banner first
    displayGarudaBanner()
    
    // Small delay untuk effect
    time.Sleep(100 * time.Millisecond)
    
    // Apply system optimizations
    optimizeSystem()
    
    cfg, err := config.Load("")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    logger := utils.NewLogger(cfg.Debug)
    
    // Log platform info
    logger.Info("Platform: %s", cfg.Platform)
    logger.Info("CPU Cores: %d (Using: %d)", runtime.NumCPU(), cfg.GetOptimalThreadCount())
    logger.Info("GOOS: %s, GOARCH: %s", runtime.GOOS, runtime.GOARCH)
    
    // Apply runtime optimizations
    applyRuntimeOptimizations(cfg, logger)
    
    raknetServer := raknet.NewServer(cfg.Server.Address, logger)
    mcServer := server.NewServer(raknetServer, cfg, logger)
    
    logger.Info("Starting Garuda Minecraft Server...")
    logger.Info("Version: %s | Players: %d/%d", 
        cfg.Server.Version, 0, cfg.Server.MaxPlayers)
    logger.Info("View Distance: %d | World: %s", 
        cfg.World.ViewDistance, cfg.World.Name)
    
    setupSignalHandling(mcServer, logger)
    
    if err := mcServer.Start(); err != nil {
        logger.Fatal("Failed to start server: %v", err)
    }
}

func displayGarudaBanner() {
    // Clear screen first (optional)
    fmt.Print("\033[2J\033[H")
    
    banner := `
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║    ██████╗  █████╗ ██████╗ ██╗   ██╗██████╗  █████╗     ███████╗██████╗     ║
║    ██╔════╝ ██╔══██╗██╔══██╗██║   ██║██╔══██╗██╔══██╗    ██╔════╝██╔══██╗    ║
║    ██║  ███╗███████║██████╔╝██║   ██║██████╔╝███████║    █████╗  ██████╔╝    ║
║    ██║   ██║██╔══██║██╔══██╗██║   ██║██╔══██╗██╔══██║    ██╔══╝  ██╔══██╗    ║
║    ╚██████╔╝██║  ██║██║  ██║╚██████╔╝██║  ██║██║  ██║    ███████╗██║  ██║    ║
║     ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝    ╚══════╝╚═╝  ╚═╝    ║
║                                                                              ║
║                        ███████╗██████╗ ██╗     ██╗██████╗                   ║
║                        ██╔════╝██╔══██╗██║     ██║██╔══██╗                  ║
║                        █████╗  ██████╔╝██║     ██║██║  ██║                  ║
║                        ██╔══╝  ██╔══██╗██║     ██║██║  ██║                  ║
║                        ██║     ██║  ██║███████╗██║██████╔╝                  ║
║                        ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝╚═════╝                   ║
║                                                                              ║
║                  ███╗   ███╗██╗   ██╗ ██████╗███████╗██████╗                 ║
║                  ████╗ ████║██║   ██║██╔════╝██╔════╝██╔══██╗                ║
║                  ██╔████╔██║██║   ██║██║     █████╗  ██║  ██║                ║
║                  ██║╚██╔╝██║██║   ██║██║     ██╔══╝  ██║  ██║                ║
║                  ██║ ╚═╝ ██║╚██████╔╝╚██████╗███████╗██████╔╝                ║
║                  ╚═╝     ╚═╝ ╚═════╝  ╚═════╝╚══════╝╚═════╝                 ║
║                                                                              ║
║                                                                              ║
║                    G A R U D A   F R A M E W O R K   M C                     ║
║                                                                              ║
║                   Copyright © 2025 - AnasBex - v1.0.0                       ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
`
    
    // Print banner dengan color codes
    fmt.Print("\033[1;36m") // Cyan color
    fmt.Println(banner)
    fmt.Print("\033[0m") // Reset color
    
    // Additional info line
    fmt.Print("\033[1;33m") // Yellow color
    fmt.Println("               [ Garuda Minecraft Server Framework ]")
    fmt.Print("\033[0m") // Reset color
    
    fmt.Println()
}

// Alternative simpler banner (jika yang atas terlalu besar)
func displaySimpleBanner() {
    fmt.Print("\033[2J\033[H") // Clear screen
    
    simpleBanner := `
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   ██████╗  █████╗ ██████╗ ██╗   ██╗██████╗  █████╗                         │
│   ██╔════╝ ██╔══██╗██╔══██╗██║   ██║██╔══██╗██╔══██╗                        │
│   ██║  ███╗███████║██████╔╝██║   ██║██████╔╝███████║                        │
│   ██║   ██║██╔══██║██╔══██╗██║   ██║██╔══██╗██╔══██║                        │
│   ╚██████╔╝██║  ██║██║  ██║╚██████╔╝██║  ██║██║  ██║                        │
│    ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝                        │
│                                                                              │
│                  G A R U D A   F R A M E W O R K   M C                       │
│                                                                              │
│                  Copyright © 2025 - AnasBex - v1.0.0                        │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
`
    fmt.Print("\033[1;35m") // Magenta color
    fmt.Println(simpleBanner)
    fmt.Print("\033[0m")
}

// Minimal banner untuk terminals kecil
func displayMinimalBanner() {
    fmt.Print("\033[2J\033[H")
    
    minimal := `
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║                        GARUDA FRAMEWORK MC                                   ║
║                                                                              ║
║                   Copyright © 2025 - AnasBex                                ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
`
    fmt.Print("\033[1;32m") // Green color
    fmt.Println(minimal)
    fmt.Print("\033[0m")
}

// Banner dengan ASCII art Garuda (burung)
func displayGarudaBirdBanner() {
    fmt.Print("\033[2J\033[H")
    
    garudaArt := `
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║    ...                                                                       ║
║    .@@@.                                                                     ║
║   .@@@@@.                                                                    ║
║  .@@@@@@@.                                                                   ║
║ .@@@@@@@@@.                                                                  ║
║  @@@@@@@@@                                 G A R U D A                       ║
║   @@@@@@@        @@@@@@   @@@@@@@  @@@  @@@ @@@@@@@@  @@@  @@@              ║
║    @@@@@        !@@      @@@@  @@@ @@@  @@@ @@@@  @@@ @@@@ @@@              ║
║     @@@@        !@@!@@!  @@@@  @@@ @@@  @@@ @@@@  @@@ @@@@@@@@              ║
║      @@@        !@@  @@@  @@@@@@@  @@@@@@@@ @@@@@@@@  @@@ @@@@              ║
║       @@         @@@@@@    @@@@@    @@@@@   @@@@      @@@  @@@              ║
║        @                                                                     ║
║         .                                                                   ║
║                                                                              ║
║                    F R A M E W O R K   M I N E C R A F T                    ║
║                                                                              ║
║                   Copyright © 2025 - AnasBex - v1.0.0                       ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
`
    fmt.Print("\033[1;34m") // Blue color
    fmt.Println(garudaArt)
    fmt.Print("\033[0m")
}

func optimizeSystem() {
    // Set optimal GOMAXPROCS
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Reduce memory allocation frequency
    debug.SetGCPercent(100)
    
    // Set memory limit jika di Go 1.19+
    setMemoryLimit()
}

func setMemoryLimit() {
    if runtime.GOOS == "linux" || runtime.GOOS == "android" {
        // Memory limit logic...
    }
}

func applyRuntimeOptimizations(cfg *config.Config, logger *utils.Logger) {
    optimalThreads := cfg.GetOptimalThreadCount()
    runtime.GOMAXPROCS(optimalThreads)
    logger.Info("Set thread count to: %d", optimalThreads)
    
    switch cfg.Platform {
    case "termux":
        logger.Info("Running in Termux - Mobile optimized mode")
    case "wsl":
        logger.Info("Running in WSL - Windows Subsystem for Linux")
    case "docker":
        logger.Info("Running in Docker container")
    case "lowmem":
        logger.Info("Running on low-memory system - Performance optimized")
    default:
        logger.Info("Running on standard system - Full features enabled")
    }
    
    if cfg.IsLowEndDevice() {
        debug.SetGCPercent(50)
    } else {
        debug.SetGCPercent(100)
    }
}

func setupSignalHandling(server *server.Server, logger *utils.Logger) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
    
    go func() {
        <-c
        logger.Info("Shutting down server gracefully...")
        server.Stop()
        os.Exit(0)
    }()
}