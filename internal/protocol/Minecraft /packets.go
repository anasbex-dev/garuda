package minecraft

import (
    "time"
)

// Packet IDs for Minecraft Bedrock Edition
const (
    IDLogin                          = 0x01
    IDPlayStatus                     = 0x02
    IDServerToClientHandshake        = 0x03
    IDClientToServerHandshake        = 0x04
    IDDisconnect                     = 0x05
    IDResourcePacksInfo              = 0x06
    IDResourcePackStack              = 0x07
    IDResourcePackClientResponse     = 0x08
    IDText                           = 0x09
    IDSetTime                        = 0x0a
    IDStartGame                      = 0x0b
    IDAddPlayer                      = 0x0c
    IDAddActor                       = 0x0d
    IDRemoveActor                    = 0x0e
    IDAddItemActor                   = 0x0f
    IDTakeItemActor                  = 0x11
    IDMoveActorAbsolute              = 0x12
    IDMovePlayer                     = 0x13
    IDRiderJump                      = 0x14
    IDUpdateBlock                    = 0x15
    IDAddPainting                    = 0x16
    IDTickSync                       = 0x17
    IDLevelSoundEventV1              = 0x18
    IDLevelEvent                     = 0x19
    IDBlockEvent                     = 0x1a
    IDActorEvent                     = 0x1b
    IDMobEffect                      = 0x1c
    IDUpdateAttributes               = 0x1d
    IDInventoryTransaction           = 0x1e
    IDMobEquipment                   = 0x1f
    IDMobArmorEquipment              = 0x20
    IDInteract                       = 0x21
    IDBlockPickRequest               = 0x22
    IDActorPickRequest               = 0x23
    IDPlayerAction                   = 0x24
    IDHurtArmor                      = 0x26
    IDSetActorData                   = 0x27
    IDSetActorMotion                 = 0x28
    IDSetActorLink                   = 0x29
    IDSetHealth                      = 0x2a
    IDSetSpawnPosition               = 0x2b
    IDAnimate                        = 0x2c
    IDRespawn                        = 0x2d
    IDContainerOpen                  = 0x2e
    IDContainerClose                 = 0x2f
    IDPlayerHotbar                   = 0x30
    IDInventoryContent               = 0x31
    IDInventorySlot                  = 0x32
    IDContainerSetData               = 0x33
    IDCraftingData                   = 0x34
    IDCraftingEvent                  = 0x35
    IDGUIDataPickItem                = 0x36
    IDAdventureSettings              = 0x37
    IDBlockActorData                 = 0x38
    IDPlayerInput                    = 0x39
    IDLevelChunk                     = 0x3a
    IDSetCommandsEnabled             = 0x3b
    IDSetDifficulty                  = 0x3c
    IDChangeDimension                = 0x3d
    IDSetPlayerGameType              = 0x3e
    IDPlayerList                     = 0x3f
    IDSimpleEvent                    = 0x40
    IDEvent                          = 0x41
    IDSpawnExperienceOrb             = 0x42
    IDClientboundMapItemData         = 0x43
    IDMapInfoRequest                 = 0x44
    IDRequestChunkRadius             = 0x45
    IDChunkRadiusUpdated             = 0x46
    IDItemFrameDropItem              = 0x47
    IDGameRulesChanged               = 0x48
    IDCamera                         = 0x49
    IDBossEvent                      = 0x4a
    IDShowCredits                    = 0x4b
    IDAvailableCommands              = 0x4c
    IDCommandRequest                 = 0x4d
    IDCommandBlockUpdate             = 0x4e
    IDCommandOutput                  = 0x4f
    IDUpdateTrade                    = 0x50
    IDUpdateEquipment                = 0x51
    IDResourcePackDataInfo           = 0x52
    IDResourcePackChunkData          = 0x53
    IDResourcePackChunkRequest       = 0x54
    IDTransfer                       = 0x55
    IDPlaySound                      = 0x56
    IDStopSound                      = 0x57
    IDSetTitle                       = 0x58
    IDAddBehaviorTree                = 0x59
    IDStructureBlockUpdate           = 0x5a
    IDShowStoreOffer                 = 0x5b
    IDPurchaseReceipt                = 0x5c
    IDPlayerSkin                     = 0x5d
    IDSubClientLogin                 = 0x5e
    IDAutomationClientConnect        = 0x5f
    IDSetLastHurtBy                  = 0x60
    IDBookEdit                       = 0x61
    IDNPCRequest                     = 0x62
    IDPhotoTransfer                  = 0x63
    IDModalFormRequest               = 0x64
    IDModalFormResponse              = 0x65
    IDServerSettingsRequest          = 0x66
    IDServerSettingsResponse         = 0x67
    IDShowProfile                    = 0x68
    IDSetDefaultGameType             = 0x69
    IDRemoveObjective                = 0x6a
    IDSetDisplayObjective            = 0x6b
    IDSetScore                       = 0x6c
    IDLabTable                       = 0x6d
    IDUpdateBlockSynced              = 0x6e
    IDMoveActorDelta                 = 0x6f
    IDSetScoreboardIdentity          = 0x70
    IDSetLocalPlayerAsInitialized    = 0x71
    IDUpdateSoftEnum                 = 0x72
    IDNetworkStackSettings           = 0x73
    IDScriptCustomEvent              = 0x75
    IDSpawnParticleEffect            = 0x76
    IDAvailableActorIdentifiers      = 0x77
    IDLevelSoundEventV2              = 0x78
    IDNetworkChunkPublisherUpdate    = 0x79
    IDBiomeDefinitionList            = 0x7a
    IDLevelSoundEvent                = 0x7b
    IDLevelEventGeneric              = 0x7c
    IDLecternUpdate                  = 0x7d
    IDVideoStreamConnect             = 0x7e
    IDAddEntity                      = 0x7f
    IDRemoveEntity                   = 0x80
    IDClientCacheStatus              = 0x81
    IDOnScreenTextureAnimation       = 0x82
    IDMapCreateLockedCopy            = 0x83
    IDStructureTemplateDataRequest   = 0x84
    IDStructureTemplateDataResponse  = 0x85
    IDClientCacheBlobStatus          = 0x86
    IDClientCacheMissResponse        = 0x87
    IDEducationSettings              = 0x88
    IDEmote                          = 0x89
    IDMultiplayerSettings            = 0x8a
    IDSettingsCommand                = 0x8b
    IDAnvilDamage                    = 0x8c
    IDCompletedUsingItem             = 0x8d
    IDNetworkSettings                = 0x8e
    IDPlayerAuthInput                = 0x8f
    IDCreativeContent                = 0x90
    IDPlayerEnchantOptions           = 0x91
    IDItemStackRequest               = 0x92
    IDItemStackResponse              = 0x93
    IDPlayerArmorDamage              = 0x94
    IDCodeBuilder                    = 0x95
    IDUpdatePlayerGameType           = 0x96
    IDEmoteList                      = 0x97
    IDPositionTrackingDBServerBroadcast = 0x98
    IDPositionTrackingDBClientRequest   = 0x99
    IDDebugInfo                      = 0x9a
    IDPacketViolationWarning         = 0x9b
    IDMotionPredictionHints          = 0x9c
    IDAnimateEntity                  = 0x9d
    IDCameraShake                    = 0x9e
    IDPlayerFog                      = 0x9f
    IDCorrectPlayerMovePrediction    = 0xa0
    IDItemComponent                  = 0xa1
    IDFilterText                     = 0xa2
    IDClientboundDebugRenderer       = 0xa3
    IDSyncActorProperty              = 0xa4
    IDAddVolumeEntity                = 0xa5
    IDRemoveVolumeEntity             = 0xa6
    IDSimulationType                 = 0xa7
    IDNPCDialogue                    = 0xa8
    IDEducationResourceURI           = 0xa9
    IDCreatePhoto                    = 0xaa
    IDUpdateSubChunkBlocks           = 0xab
    IDPhotoInfoRequest               = 0xac
    IDSubChunk                       = 0xad
    IDSubChunkRequest                = 0xae
    IDPlayerStartItemCooldown        = 0xaf
    IDScriptMessage                  = 0xb0
    IDCodeBuilderSource              = 0xb1
    IDTickingAreasLoad               = 0xb2
    IDDimensionData                  = 0xb3
    IDAgentAction                    = 0xb4
    IDChangeMobProperty              = 0xb5
    IDLessonProgress                 = 0xb6
    IDRequestAbility                 = 0xb7
    IDRequestPermissions             = 0xb8
    IDToastRequest                   = 0xb9
    IDUpdateAbilities                = 0xba
    IDUpdateAdventureSettings        = 0xbb
    IDDeathInfo                      = 0xbc
    IDEditorNetwork                  = 0xbd
)

// LoginPacket represents a client login request
type LoginPacket struct {
    ProtocolVersion int32
    ConnectionRequestData []byte
    ClientNetworkVersion int32
    PlatformID int32
    PlatformOfflineID string
    PlatformOnlineID string
    SelfSignedID string
    ServerAddress string
    LanguageCode string
    SkinData SkinData
}

// PlayStatusPacket represents the status of player login
type PlayStatusPacket struct {
    Status int32
}

// Play status constants
const (
    PlayStatusLoginSuccess = iota
    PlayStatusLoginFailedClient
    PlayStatusLoginFailedServer
    PlayStatusPlayerSpawn
    PlayStatusLoginFailedInvalidTenant
    PlayStatusLoginFailedVanillaEdu
    PlayStatusLoginFailedEduVanilla
    PlayStatusLoginFailedServerFull
)

// DisconnectPacket represents a disconnection notification
type DisconnectPacket struct {
    HideDisconnectionScreen bool
    Message string
}

// TextPacket represents a chat or system message
type TextPacket struct {
    TextType byte
    NeedsTranslation bool
    SourceName string
    Message string
    Parameters []string
    XUID string
    PlatformChatID string
    FilteredMessage string
}

// Text types
const (
    TextTypeRaw = iota
    TextTypeChat
    TextTypeTranslation
    TextTypePopup
    TextTypeJukeboxPopup
    TextTypeTip
    TextTypeSystem
    TextTypeWhisper
    TextTypeAnnouncement
    TextTypeJsonWhisper
    TextTypeJson
)

// StartGamePacket contains game initialization data
type StartGamePacket struct {
    EntityID int64
    RuntimeEntityID uint64
    PlayerGameType int32
    PlayerPosition Vector3
    Rotation Vector2
    Seed int32
    BiomeType int16
    BiomeName string
    Dimension int32
    Generator int32
    WorldGameMode int32
    Difficulty int32
    SpawnPosition BlockPos
    AchievementsDisabled bool
    Time int32
    EduMode bool
    RainLevel float32
    LightningLevel float32
    CommandsEnabled bool
    TexturePacksRequired bool
    GameRules []GameRule
    Experiments []Experiment
    ExperimentsPreviouslyUsed bool
    BonusChestEnabled bool
    StartWithMapEnabled bool
    PlayerPermissions int32
    ChunkRadius int32
    ServerChunkTickRange int32
    BroadcastToLAN bool
    XBLBroadcastMode int32
    PlatformBroadcastMode int32
    XBLBroadcastIntent bool
    PlatformBroadcastIntent bool
    CommandsEnabledOnFirstJoin bool
}

// GameRule represents a game rule
type GameRule struct {
    Name string
    Value interface{}
    Type GameRuleType
}

// GameRuleType defines the type of game rule value
type GameRuleType int

const (
    GameRuleBool GameRuleType = iota
    GameRuleInt
    GameRuleFloat
)

// Experiment represents a game experiment
type Experiment struct {
    Name string
    Enabled bool
}

// MovePlayerPacket represents player movement
type MovePlayerPacket struct {
    RuntimeID uint64
    Position Vector3
    Rotation Vector2
    Mode byte
    OnGround bool
    Tick uint64
    VehicleRuntimeID uint64
}

// Move modes
const (
    MoveModeNormal = iota
    MoveModeReset
    MoveModeTeleport
    MoveModeRotation
)

// PlayerActionPacket represents player actions
type PlayerActionPacket struct {
    RuntimeID uint64
    Action int32
    Position BlockPos
    Face int32
}

// Player actions
const (
    ActionStartBreak = iota
    ActionAbortBreak
    ActionStopBreak
    ActionGetUpdatedBlock
    ActionDropItem
    ActionStartSleeping
    ActionStopSleeping
    ActionRespawn
    ActionJump
    ActionStartSprint
    ActionStopSprint
    ActionStartSneak
    ActionStopSneak
    ActionCreativeDestroy
    ActionDimensionChange
    ActionDimensionChangeAck
    ActionStartGlide
    ActionStopGlide
    ActionBuildDenied
    ActionCrackBreak
    ActionChangeSkin
    ActionSetEnchantmentSeed
    ActionStartSwimming
    ActionStopSwimming
    ActionStartSpinAttack
    ActionStopSpinAttack
    ActionInteractBlock
)

// InventoryTransactionPacket represents inventory changes
type InventoryTransactionPacket struct {
    TransactionType uint32
    Actions []InventoryAction
    TransactionData TransactionData
}

// InventoryAction represents a single inventory action
type InventoryAction struct {
    SourceType uint32
    WindowID uint32
    SourceFlags uint32
    InventorySlot uint32
    OldItem ItemStack
    NewItem ItemStack
}

// TransactionData contains transaction-specific data
type TransactionData struct {
    RequestID int32
    RequestChanged []int32
    TransactionType int32
}

// MobEquipmentPacket represents equipment changes
type MobEquipmentPacket struct {
    RuntimeID uint64
    Item ItemStack
    Slot byte
    SelectedSlot byte
    WindowID byte
}

// MobArmorEquipmentPacket represents armor changes
type MobArmorEquipmentPacket struct {
    RuntimeID uint64
    Helmet ItemStack
    Chestplate ItemStack
    Leggings ItemStack
    Boots ItemStack
}

// LevelChunkPacket represents a chunk data packet
type LevelChunkPacket struct {
    ChunkX int32
    ChunkZ int32
    SubChunkCount uint32
    Data []byte
    CacheEnabled bool
    BlobHashes []uint64
}

// UpdateBlockPacket represents a block update
type UpdateBlockPacket struct {
    Position BlockPos
    BlockID uint32
    Flags uint32
    Layer uint32
}

// Block update flags
const (
    BlockUpdateNeighbors = 1 << iota
    BlockUpdateNetwork
    BlockUpdateNoGraphics
    BlockUpdatePriority
)

// AnimatePacket represents animations
type AnimatePacket struct {
    Action byte
    RuntimeID uint64
    RowdingTime float32
}

// Animate actions
const (
    AnimateActionSwingArm = iota + 1
    AnimateActionStopSleep
    AnimateActionCriticalHit
    AnimateActionMagicCriticalHit
    AnimateActionRowRight
    AnimateActionRowLeft
)

// SetTimePacket represents time updates
type SetTimePacket struct {
    Time int32
}

// RespawnPacket represents respawn data
type RespawnPacket struct {
    Position Vector3
    State byte
    RuntimeID uint64
}

// AdventureSettingsPacket represents player settings
type AdventureSettingsPacket struct {
    Flags uint32
    CommandPermission uint32
    ActionPermissions uint32
    PermissionLevel uint32
    CustomStoredPermissions uint32
    PlayerUniqueID int64
}

// Adventure flags
const (
    AdventureFlagWorldImmutable = 1 << iota
    AdventureFlagNoPvM
    AdventureFlagNoMvP
    AdventureFlagUnused
    AdventureFlagShowNameTags
    AdventureFlagAutoJump
    AdventureFlagAllowFlight
    AdventureFlagNoClip
    AdventureFlagWorldBuilder
    AdventureFlagFlying
    AdventureFlagMuted
)

// CommandRequestPacket represents command execution
type CommandRequestPacket struct {
    Command string
    Type byte
    RequestID string
    Internal bool
    Version int32
}

// AvailableCommandsPacket represents available commands
type AvailableCommandsPacket struct {
    Commands []CommandData
    Constraints []ConstraintData
}

// CommandData represents a command definition
type CommandData struct {
    Name string
    Description string
    Flags byte
    Permission byte
    Aliases []string
    Overloads []CommandOverload
}

// CommandOverload represents command parameters
type CommandOverload struct {
    Parameters []CommandParameter
}

// CommandParameter represents a command parameter
type CommandParameter struct {
    Name string
    Type int32
    Optional bool
    Options byte
}

// ConstraintData represents command constraints
type ConstraintData struct {
    Constraints []Constraint
}

// Constraint represents a command constraint
type Constraint struct {
    Value int32
    Constraints []byte
}

// Vector3 represents a 3D vector
type Vector3 struct {
    X float32
    Y float32
    Z float32
}

// Vector2 represents a 2D vector
type Vector2 struct {
    X float32
    Y float32
}

// BlockPos represents a block position
type BlockPos struct {
    X int32
    Y int32
    Z int32
}

// ItemStack represents an item stack
type ItemStack struct {
    ID uint16
    Count byte
    Damage uint16
    NBT map[string]interface{}
}

// SkinData represents player skin data
type SkinData struct {
    SkinID string
    SkinResourcePatch []byte
    SkinImageWidth uint32
    SkinImageHeight uint32
    SkinData []byte
    CapeImageWidth uint32
    CapeImageHeight uint32
    CapeData []byte
    SkinGeometry []byte
    AnimationData []byte
    PremiumSkin bool
    PersonaSkin bool
    CapeOnClassicSkin bool
    CapeID string
    FullSkinID string
    ArmSize string
    SkinColor string
    PersonaPieces []PersonaPiece
    PieceTintColors []PieceTintColor
}

// PersonaPiece represents a persona skin piece
type PersonaPiece struct {
    PieceID string
    PieceType string
    PackID string
    IsDefault bool
    ProductID string
}

// PieceTintColor represents tint color for persona pieces
type PieceTintColor struct {
    PieceType string
    Colors []string
}

// Packet represents a generic Minecraft packet
type Packet struct {
    ID byte
    Data []byte
    Timestamp time.Time
}

// NewPacket creates a new packet
func NewPacket(id byte, data []byte) *Packet {
    return &Packet{
        ID: id,
        Data: data,
        Timestamp: time.Now(),
    }
}

// GetID returns the packet ID
func (p *Packet) GetID() byte {
    return p.ID
}

// GetData returns the packet data
func (p *Packet) GetData() []byte {
    return p.Data
}

// GetTimestamp returns the packet timestamp
func (p *Packet) GetTimestamp() time.Time {
    return p.Timestamp
}

// SetData sets the packet data
func (p *Packet) SetData(data []byte) {
    p.Data = data
}

// String returns a string representation of the packet
func (p *Packet) String() string {
    return string(p.Data)
}

// Utility functions for common operations
func (v Vector3) DistanceTo(other Vector3) float32 {
    dx := v.X - other.X
    dy := v.Y - other.Y
    dz := v.Z - other.Z
    return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

func (v Vector3) Add(other Vector3) Vector3 {
    return Vector3{
        X: v.X + other.X,
        Y: v.Y + other.Y,
        Z: v.Z + other.Z,
    }
}

func (v Vector3) Subtract(other Vector3) Vector3 {
    return Vector3{
        X: v.X - other.X,
        Y: v.Y - other.Y,
        Z: v.Z - other.Z,
    }
}

func (v Vector3) Multiply(scalar float32) Vector3 {
    return Vector3{
        X: v.X * scalar,
        Y: v.Y * scalar,
        Z: v.Z * scalar,
    }
}

func (pos BlockPos) ToVector3() Vector3 {
    return Vector3{
        X: float32(pos.X),
        Y: float32(pos.Y),
        Z: float32(pos.Z),
    }
}

func Vector3ToBlockPos(v Vector3) BlockPos {
    return BlockPos{
        X: int32(math.Floor(float64(v.X))),
        Y: int32(math.Floor(float64(v.Y))),
        Z: int32(math.Floor(float64(v.Z))),
    }
}

// ItemStack utility methods
func (i *ItemStack) IsEmpty() bool {
    return i.ID == 0 || i.Count == 0
}

func (i *ItemStack) Clone() *ItemStack {
    return &ItemStack{
        ID:    i.ID,
        Count: i.Count,
        Damage: i.Damage,
        NBT:   i.NBT, // Note: shallow copy, implement deep copy if needed
    }
}

// Game rule constants
const (
    GameRuleCommandBlockOutput = "commandblockoutput"
    GameRuleDoDaylightCycle = "dodaylightcycle"
    GameRuleDoEntityDrops = "doentitydrops"
    GameRuleDoFireTick = "dofiretick"
    GameRuleDoMobLoot = "domobloot"
    GameRuleDoMobSpawning = "domobspawning"
    GameRuleDoTileDrops = "dotiledrops"
    GameRuleDoWeatherCycle = "doweathercycle"
    GameRuleDrowningDamage = "drowningdamage"
    GameRuleFallDamage = "falldamage"
    GameRuleFireDamage = "firedamage"
    GameRuleKeepInventory = "keepinventory"
    GameRuleMobGriefing = "mobgriefing"
    GameRuleNaturalRegeneration = "naturalregeneration"
    GameRulePvp = "pvp"
    GameRuleSendCommandFeedback = "sendcommandfeedback"
    GameRuleShowCoordinates = "showcoordinates"
    GameRuleRandomTickSpeed = "randomtickspeed"
    GameRuleTntExplodes = "tntexplodes"
)

// Default game rules
var DefaultGameRules = map[string]interface{}{
    GameRuleCommandBlockOutput:   true,
    GameRuleDoDaylightCycle:      true,
    GameRuleDoEntityDrops:        true,
    GameRuleDoFireTick:           true,
    GameRuleDoMobLoot:            true,
    GameRuleDoMobSpawning:        true,
    GameRuleDoTileDrops:          true,
    GameRuleDoWeatherCycle:       true,
    GameRuleDrowningDamage:       true,
    GameRuleFallDamage:           true,
    GameRuleFireDamage:           true,
    GameRuleKeepInventory:        false,
    GameRuleMobGriefing:          true,
    GameRuleNaturalRegeneration:  true,
    GameRulePvp:                  true,
    GameRuleSendCommandFeedback:  true,
    GameRuleShowCoordinates:      true,
    GameRuleRandomTickSpeed:      3,
    GameRuleTntExplodes:          true,
}