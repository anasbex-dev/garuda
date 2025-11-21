package api

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/gorilla/mux"
    "github.com/anabex-dev/garuda/internal/server"
    "github.com/anabex-dev/garuda/internal/world"
)

// APIServer represents the REST API server
type APIServer struct {
    garudaServer *server.GarudaServer
    router       *mux.Router
    httpServer   *http.Server
    running      bool
}

// APIResponse standard API response format
type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// NewAPIServer creates a new API server instance
func NewAPIServer(garudaServer *server.GarudaServer) *APIServer {
    api := &APIServer{
        garudaServer: garudaServer,
        router:       mux.NewRouter(),
        running:      false,
    }
    
    api.setupRoutes()
    return api
}

// setupRoutes defines all API routes
func (a *APIServer) setupRoutes() {
    // API middleware
    a.router.Use(a.loggingMiddleware)
    a.router.Use(a.authMiddleware)

    // Health check
    a.router.HandleFunc("/api/health", a.healthHandler).Methods("GET")
    
    // Server information
    a.router.HandleFunc("/api/server/info", a.serverInfoHandler).Methods("GET")
    a.router.HandleFunc("/api/server/stats", a.serverStatsHandler).Methods("GET")
    
    // Player management
    a.router.HandleFunc("/api/players", a.listPlayersHandler).Methods("GET")
    a.router.HandleFunc("/api/players/{username}", a.playerInfoHandler).Methods("GET")
    a.router.HandleFunc("/api/players/{username}/kick", a.kickPlayerHandler).Methods("POST")
    a.router.HandleFunc("/api/players/{username}/ban", a.banPlayerHandler).Methods("POST")
    
    // World management
    a.router.HandleFunc("/api/world/info", a.worldInfoHandler).Methods("GET")
    a.router.HandleFunc("/api/world/time", a.worldTimeHandler).Methods("GET", "POST")
    a.router.HandleFunc("/api/world/weather", a.worldWeatherHandler).Methods("GET", "POST")
    
    // Command execution
    a.router.HandleFunc("/api/command", a.executeCommandHandler).Methods("POST")
    
    // Plugin management
    a.router.HandleFunc("/api/plugins", a.listPluginsHandler).Methods("GET")
    a.router.HandleFunc("/api/plugins/{name}/enable", a.enablePluginHandler).Methods("POST")
    a.router.HandleFunc("/api/plugins/{name}/disable", a.disablePluginHandler).Methods("POST")
    
    // Server control
    a.router.HandleFunc("/api/server/stop", a.stopServerHandler).Methods("POST")
    a.router.HandleFunc("/api/server/restart", a.restartServerHandler).Methods("POST")
    a.router.HandleFunc("/api/server/save", a.saveWorldHandler).Methods("POST")
    
    // Console log streaming (WebSocket)
    a.router.HandleFunc("/api/console/stream", a.consoleStreamHandler)
    
    // Serve static files for web dashboard
    a.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))
}

// Start starts the API server
func (a *APIServer) Start(port int) error {
    if a.running {
        return fmt.Errorf("API server is already running")
    }

    a.httpServer = &http.Server{
        Addr:         fmt.Sprintf(":%d", port),
        Handler:      a.router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    a.running = true
    
    log.Printf("REST API server starting on port %d", port)
    
    go func() {
        if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Printf("API server error: %v", err)
        }
    }()
    
    return nil
}

// Stop stops the API server
func (a *APIServer) Stop() error {
    if !a.running {
        return nil
    }
    
    a.running = false
    if a.httpServer != nil {
        return a.httpServer.Close()
    }
    return nil
}

// ===== HANDLER IMPLEMENTATIONS =====

// healthHandler returns server health status
func (a *APIServer) healthHandler(w http.ResponseWriter, r *http.Request) {
    a.jsonResponse(w, APIResponse{
        Success: true,
        Message: "Server is healthy",
        Data: map[string]interface{}{
            "status":    "online",
            "timestamp": time.Now().Unix(),
        },
    })
}

// serverInfoHandler returns server information
func (a *APIServer) serverInfoHandler(w http.ResponseWriter, r *http.Request) {
    stats := a.garudaServer.GetServerStats()
    
    a.jsonResponse(w, APIResponse{
        Success: true,
        Data: map[string]interface{}{
            "name":          "Garuda Server",
            "version":       "1.0.0",
            "motd":          "Garuda Minecraft Server",
            "players_online": a.garudaServer.GetPlayerCount(),
            "max_players":   a.garudaServer.GetMaxPlayers(),
            "world_name":    a.garudaServer.GetWorldName(),
            "uptime":        stats["uptime"],
            "start_time":    stats["start_time"],
        },
    })
}

// serverStatsHandler returns detailed server statistics
func (a *APIServer) serverStatsHandler(w http.ResponseWriter, r *http.Request) {
    stats := a.garudaServer.GetServerStats()
    
    a.jsonResponse(w, APIResponse{
        Success: true,
        Data:    stats,
    })
}

// listPlayersHandler returns list of online players
func (a *APIServer) listPlayersHandler(w http.ResponseWriter, r *http.Request) {
    players := a.garudaServer.GetOnlinePlayers()
    
    playerList := make([]map[string]interface{}, len(players))
    for i, player := range players {
        playerList[i] = map[string]interface{}{
            "username":   player.Username,
            "entity_id":  player.EntityID,
            "position":   player.Position,
            "health":     player.Health,
            "game_mode":  player.GameMode,
            "op":         a.garudaServer.IsOp(player.Username),
        }
    }
    
    a.jsonResponse(w, APIResponse{
        Success: true,
        Data: map[string]interface{}{
            "players": playerList,
            "count":   len(players),
        },
    })
}

// playerInfoHandler returns information about a specific player
func (a *APIServer) playerInfoHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    username := vars["username"]
    
    player := a.garudaServer.GetPlayer(username)
    if player == nil {
        a.errorResponse(w, "Player not found", http.StatusNotFound)
        return
    }
    
    a.jsonResponse(w, APIResponse{
        Success: true,
        Data: map[string]interface{}{
            "username":    player.Username,
            "entity_id":   player.EntityID,
            "position":    player.Position,
            "rotation":    player.Rotation,
            "health":      player.Health,
            "max_health":  player.MaxHealth,
            "hunger":      player.Hunger,
            "game_mode":   player.GameMode,
            "op":          a.garudaServer.IsOp(player.Username),
            "whitelisted": a.garudaServer.IsWhitelisted(player.Username),
            "banned":      a.garudaServer.IsPlayerBanned(player.Username),
        },
    })
}

// kickPlayerHandler kicks a player from the server
func (a *APIServer) kickPlayerHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    username := vars["username"]
    
    var request struct {
        Reason string `json:"reason"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        a.errorResponse(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    player := a.garudaServer.GetPlayer(username)
    if player == nil {
        a.errorResponse(w, "Player not found", http.StatusNotFound)
        return
    }
    
    reason := request.Reason
    if reason == "" {
        reason = "Kicked via API"
    }
    
    a.garudaServer.KickPlayer(player, reason)
    
    a.jsonResponse(w, APIResponse{
        Success: true,
        Message: fmt.Sprintf("Player %s kicked: %s", username, reason),
    })
}

// banPlayerHandler bans a player from the server
func (a *APIServer) banPlayerHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    username := vars["username"]
    
    var request struct {
        Reason string `json:"reason"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        a.errorResponse(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    reason := request.Reason
    if reason == "" {
        reason = "Banned via API"
    }
    
    if a.garudaServer.BanPlayer(username, reason, "API") {
        // Kick player if online
        if player := a.garudaServer.GetPlayer(username); player != nil {
            a.garudaServer.KickPlayer(player, reason)
        }
        
        a.jsonResponse(w, APIResponse{
            Success: true,
            Message: fmt.Sprintf("Player %s banned: %s", username, reason),
        })
    } else {
        a.errorResponse(w, "Failed to ban player", http.StatusInternalServerError)
    }
}

// worldInfoHandler returns world information
func (a *APIServer) worldInfoHandler(w http.ResponseWriter, r *http.Request) {
    a.jsonResponse(w, APIResponse{
        Success: true,
        Data: map[string]interface{}{
            "name":       a.garudaServer.GetWorldName(),
            "seed":       a.garudaServer.GetWorldSeed(),
            "time":       a.garudaServer.GetWorldTime(),
            "spawn_point": a.garudaServer.GetSpawnPoint(),
        },
    })
}

// worldTimeHandler gets or sets world time
func (a *APIServer) worldTimeHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        a.jsonResponse(w, APIResponse{
            Success: true,
            Data: map[string]interface{}{
                "time": a.garudaServer.GetWorldTime(),
            },
        })
        
    case "POST":
        var request struct {
            Time int32 `json:"time"`
        }
        
        if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
            a.errorResponse(w, "Invalid request body", http.StatusBadRequest)
            return
        }
        
        a.garudaServer.SetTime(request.Time)
        
        a.jsonResponse(w, APIResponse{
            Success: true,
            Message: fmt.Sprintf("World time set to %d", request.Time),
        })
    }
}

// executeCommandHandler executes a server command
func (a *APIServer) executeCommandHandler(w http.ResponseWriter, r *http.Request) {
    var request struct {
        Command string `json:"command"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        a.errorResponse(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if request.Command == "" {
        a.errorResponse(w, "Command cannot be empty", http.StatusBadRequest)
        return
    }
    
    if a.garudaServer.ExecuteCommand(request.Command) {
        a.jsonResponse(w, APIResponse{
            Success: true,
            Message: fmt.Sprintf("Command executed: %s", request.Command),
        })
    } else {
        a.errorResponse(w, "Failed to execute command", http.StatusInternalServerError)
    }
}

// stopServerHandler stops the server
func (a *APIServer) stopServerHandler(w http.ResponseWriter, r *http.Request) {
    a.jsonResponse(w, APIResponse{
        Success: true,
        Message: "Server shutdown initiated",
    })
    
    // Stop server in goroutine to allow response to be sent
    go func() {
        time.Sleep(1 * time.Second)
        a.garudaServer.Stop()
    }()
}

// saveWorldHandler saves the world
func (a *APIServer) saveWorldHandler(w http.ResponseWriter, r *http.Request) {
    if a.garudaServer.SaveWorld() {
        a.jsonResponse(w, APIResponse{
            Success: true,
            Message: "World saved successfully",
        })
    } else {
        a.errorResponse(w, "Failed to save world", http.StatusInternalServerError)
    }
}

// ===== MIDDLEWARE =====

// loggingMiddleware logs all API requests
func (a *APIServer) loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Create response wrapper to capture status code
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next.ServeHTTP(rw, r)
        
        log.Printf("API %s %s %d %v", r.Method, r.URL.Path, rw.statusCode, time.Since(start))
    })
}

// authMiddleware handles API authentication
func (a *APIServer) authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip auth for health check
        if r.URL.Path == "/api/health" {
            next.ServeHTTP(w, r)
            return
        }
        
        // Simple API key authentication
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            apiKey = r.URL.Query().Get("api_key")
        }
        
        // TODO: Implement proper API key validation
        if apiKey == "" {
            a.errorResponse(w, "API key required", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// ===== HELPER FUNCTIONS =====

// jsonResponse sends a JSON response
func (a *APIServer) jsonResponse(w http.ResponseWriter, response APIResponse) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
        log.Printf("Error encoding JSON response: %v", err)
    }
}

// errorResponse sends an error response
func (a *APIServer) errorResponse(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    response := APIResponse{
        Success: false,
        Error:   message,
    }
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
        log.Printf("Error encoding error response: %v", err)
    }
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}