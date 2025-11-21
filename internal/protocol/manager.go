package protocol

import (
    "garuda/pkg/utils"
    "sort"
    "strings"
)

type ProtocolManager struct {
    logger *utils.Logger
    currentVersion ProtocolVersion
}

func NewProtocolManager(logger *utils.Logger) *ProtocolManager {
    return &ProtocolManager{
        logger: logger,
        currentVersion: GetLatestVersion(),
    }
}

func (pm *ProtocolManager) SetServerVersion(version string) bool {
    if v, exists := GetVersion(version); exists {
        pm.currentVersion = v
        pm.logger.Info("Server protocol version set to: %s (Protocol: %d)", v.Version, v.Protocol)
        return true
    }
    pm.logger.Error("Unsupported version: %s", version)
    return false
}

func (pm *ProtocolManager) GetServerVersion() ProtocolVersion {
    return pm.currentVersion
}

func (pm *ProtocolManager) HandleClientProtocol(protocol int) (ProtocolVersion, bool) {
    if v, exists := GetVersionByProtocol(protocol); exists {
        pm.logger.Info("Client connected with protocol: %d (%s)", protocol, v.Version)
        return v, true
    }
    
    // Find closest supported version
    closest := pm.findClosestVersion(protocol)
    if closest.Protocol != 0 {
        pm.logger.Warn("Client protocol %d not supported, using closest: %s (%d)", 
            protocol, closest.Version, closest.Protocol)
        return closest, true
    }
    
    pm.logger.Error("No compatible protocol found for client: %d", protocol)
    return ProtocolVersion{}, false
}

func (pm *ProtocolManager) findClosestVersion(clientProtocol int) ProtocolVersion {
    var closest ProtocolVersion
    minDiff := 1000 // Large number
    
    for _, v := range Versions {
        diff := abs(v.Protocol - clientProtocol)
        if diff < minDiff {
            minDiff = diff
            closest = v
        }
    }
    
    return closest
}

func (pm *ProtocolManager) ValidateVersionCompatibility(clientVersion, serverVersion ProtocolVersion) bool {
    // Allow clients with same or older protocol versions
    return clientVersion.Protocol <= serverVersion.Protocol
}

func (pm *ProtocolManager) GetMOTD(serverName string, playerCount, maxPlayers int) string {
    version := pm.currentVersion
    return strings.Join([]string{
        "MCPE",
        serverName,
        version.Version,
        "0",
        toString(playerCount),
        toString(maxPlayers),
        "0",
        "Garuda",
        "Survival",
        "1",
    }, ";")
}

func (pm *ProtocolManager) GetCompressionThreshold() int {
    return pm.currentVersion.Features.CompressionThreshold
}

func (pm *ProtocolManager) GetRakNetVersion() byte {
    return pm.currentVersion.Features.RakNetVersion
}

func (pm *ProtocolManager) LogSupportedVersions() {
    versions := GetSupportedVersions()
    sort.Strings(versions)
    
    pm.logger.Info("Supported Minecraft versions:")
    for i, v := range versions {
        protocol := Versions[v].Protocol
        pm.logger.Info("  %d. %s (Protocol: %d)", i+1, v, protocol)
    }
}

// Helper functions
func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}

func toString(n int) string {
    // Simple int to string conversion
    if n == 0 {
        return "0"
    }
    
    var result string
    for n > 0 {
        result = string('0'+byte(n%10)) + result
        n /= 10
    }
    return result
}