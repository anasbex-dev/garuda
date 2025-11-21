package world

type Player struct {
    EntityID      int64
    RuntimeID     uint64
    Username      string
    Position      minecraft.Vector3
    Rotation      minecraft.Vector2
    GameMode      int32
    ChunkCoord    ChunkCoord
    Inventory     *Inventory
    Health        float32
    MaxHealth     float32
    Hunger        int32
    Saturation    float32
    Experience    int32
    Level         int32
}

func NewPlayer(entityID int64, username string) *Player {
    player := &Player{
        EntityID:   entityID,
        RuntimeID:  uint64(entityID),
        Username:   username,
        Position:   minecraft.Vector3{X: 0, Y: 70, Z: 0},
        Rotation:   minecraft.Vector2{X: 0, Y: 0},
        GameMode:   1, // Survival
        ChunkCoord: ChunkCoord{X: 0, Z: 0},
        Inventory:  NewInventory(36), // 27 main + 9 hotbar
        Health:     20.0,
        MaxHealth:  20.0,
        Hunger:     20,
        Saturation: 5.0,
        Experience: 0,
        Level:      0,
    }
    
    // Give starter items
    player.giveStarterItems()
    
    return player
}

func (p *Player) giveStarterItems() {
    // Wooden tools
    p.Inventory.SetItem(0, &ItemStack{ID: 268, Count: 1, Damage: 0}) // Wooden sword
    p.Inventory.SetItem(1, &ItemStack{ID: 269, Count: 1, Damage: 0}) // Wooden shovel
    p.Inventory.SetItem(2, &ItemStack{ID: 270, Count: 1, Damage: 0}) // Wooden pickaxe
    p.Inventory.SetItem(3, &ItemStack{ID: 271, Count: 1, Damage: 0}) // Wooden axe
    
    // Food
    p.Inventory.SetItem(8, &ItemStack{ID: 260, Count: 16, Damage: 0}) // Apples
    
    // Blocks
    p.Inventory.SetItem(4, &ItemStack{ID: 5, Count: 32, Damage: 0})   // Oak wood
    p.Inventory.SetItem(5, &ItemStack{ID: 6, Count: 64, Damage: 0})   // Oak planks
    p.Inventory.SetItem(6, &ItemStack{ID: 4, Count: 32, Damage: 0})   // Cobblestone
    p.Inventory.SetItem(7, &ItemStack{ID: 3, Count: 16, Damage: 0})   // Dirt
}

func (p *Player) Damage(amount float32) {
    p.Health -= amount
    if p.Health < 0 {
        p.Health = 0
    }
}

func (p *Player) Heal(amount float32) {
    p.Health += amount
    if p.Health > p.MaxHealth {
        p.Health = p.MaxHealth
    }
}

func (p *Player) AddExperience(amount int32) {
    p.Experience += amount
    // Simple level calculation
    p.Level = p.Experience / 100
}

func (p *Player) GetHeldItem() *ItemStack {
    return p.Inventory.GetHeldItem()
}

func (p *Player) SetHeldSlot(slot int) bool {
    return p.Inventory.SetHeldSlot(slot)
}