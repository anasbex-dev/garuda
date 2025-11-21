package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)

// worldWeatherHandler gets or sets world weather
func (a *APIServer) worldWeatherHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        a.jsonResponse(w, APIResponse{
            Success: true,
            Data: map[string]interface{}{
                "weather": "clear", // TODO: Get actual weather
            },
        })
        
    case "POST":
        var request struct {
            Weather  string `json:"weather"`
            Duration int    `json:"duration"`
        }
        
        if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
            a.errorResponse(w, "Invalid request body", http.StatusBadRequest)
            return
        }
        
        // Convert weather string to type
        weatherType := 0 // clear
        switch request.Weather {
        case "rain":
            weatherType = 1
        case "thunder":
            weatherType = 2
        }
        
        a.garudaServer.SetWeather(weatherType, request.Duration)
        
        a.jsonResponse(w, APIResponse{
            Success: true,
            Message: fmt.Sprintf("Weather set to %s for %d ticks", request.Weather, request.Duration),
        })
    }
}

// listPluginsHandler returns list of loaded plugins
func (a *APIServer) listPluginsHandler(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement plugin listing
    a.jsonResponse(w, APIResponse{
        Success: true,
        Data: map[string]interface{}{
            "plugins": []string{}, // Empty for now
        },
    })
}

// consoleStreamHandler handles WebSocket console log streaming
func (a *APIServer) consoleStreamHandler(w http.ResponseWriter, r *http.Request) {
    var upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }
    
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    defer conn.Close()
    
    // TODO: Implement console log streaming
    for {
        message := map[string]interface{}{
            "type":    "log",
            "message": "Console streaming not yet implemented",
            "time":    time.Now().Unix(),
        }
        
        if err := conn.WriteJSON(message); err != nil {
            break
        }
        
        time.Sleep(5 * time.Second)
    }
}

// Add these missing methods to GarudaServer if not exists
func (s *GarudaServer) GetMaxPlayers() int {
    // TODO: Implement based on config
    return 20
}