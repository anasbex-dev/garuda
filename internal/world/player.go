package world

import (
    "garuda/pkg/utils"
)

type PlayerInventory struct {
    Items    [36]ItemStack
    Selected int
}

type ItemStack struct {
    ID    uint32
    Count byte
    Data  uint16
}

type Player struct {
    Entity
    Username     string
    UUID         string
    Gamemode     int32
    Inventory    PlayerInventory
    Experience   int32
    Level        int32
    FoodLevel    int32
    IsSprinting  bool
    IsSneaking   bool
    IsFlying     bool
    Abilities    PlayerAbilities
}

type PlayerAbilities struct {
    FlySpeed    float32
    WalkSpeed   float32
    MayFly      bool
    Flying      bool
    Invulnerable bool
    MayBuild    bool
    Instabuild  bool
}

func NewPlayer(username, uuid string, world *World, position [3]float32) *Player {
    player := &Player{
        Entity: Entity{
            ID:       -1,
            Type:     EntityPlayer,
            Position: position,
            Rotation: [2]float32{0, 0},
            Velocity: [3]float32{0, 0, 0},
            Health:   20.0,
            MaxHealth: 20.0,
            Metadata: make(map[string]interface{}),
            World:    world,
        },
        Username:   username,
        UUID:       uuid,
        Gamemode:   1,
        Experience: 0,
        Level:      0,
        FoodLevel:  20,
        Abilities: PlayerAbilities{
            FlySpeed:    0.05,
            WalkSpeed:   0.1,
            MayFly:      true,
            Flying:      false,
            Invulnerable: false,
            MayBuild:    true,
            Instabuild:  false,
        },
    }

    player.Inventory = PlayerInventory{
        Selected: 0,
    }

    player.initializeInventory()

    return player
}

func (p *Player) initializeInventory() {
    for i := range p.Inventory.Items {
        p.Inventory.Items[i] = ItemStack{ID: 0, Count: 0, Data: 0}
    }

    p.Inventory.Items[0] = ItemStack{ID: 270, Count: 1, Data: 0}
    p.Inventory.Items[1] = ItemStack{ID: 273, Count: 1, Data: 0}
    p.Inventory.Items[2] = ItemStack{ID: 274, Count: 1, Data: 0}
}

func (p *Player) GetSelectedItem() ItemStack {
    if p.Inventory.Selected >= 0 && p.Inventory.Selected < len(p.Inventory.Items) {
        return p.Inventory.Items[p.Inventory.Selected]
    }
    return ItemStack{ID: 0, Count: 0, Data: 0}
}

func (p *Player) SetSelectedSlot(slot int) {
    if slot >= 0 && slot < len(p.Inventory.Items) {
        p.Inventory.Selected = slot
    }
}

func (p *Player) AddItem(item ItemStack) bool {
    for i := range p.Inventory.Items {
        if p.Inventory.Items[i].ID == item.ID && p.Inventory.Items[i].Data == item.Data {
            if p.Inventory.Items[i].Count+item.Count <= 64 {
                p.Inventory.Items[i].Count += item.Count
                return true
            }
        }
    }

    for i := range p.Inventory.Items {
        if p.Inventory.Items[i].ID == 0 {
            p.Inventory.Items[i] = item
            return true
        }
    }

    return false
}

func (p *Player) RemoveItem(slot int, count byte) bool {
    if slot < 0 || slot >= len(p.Inventory.Items) {
        return false
    }

    if p.Inventory.Items[slot].Count <= count {
        p.Inventory.Items[slot] = ItemStack{ID: 0, Count: 0, Data: 0}
    } else {
        p.Inventory.Items[slot].Count -= count
    }

    return true
}

func (p *Player) FindItem(itemID uint32) int {
    for i, item := range p.Inventory.Items {
        if item.ID == itemID && item.Count > 0 {
            return i
        }
    }
    return -1
}

func (p *Player) HasItem(itemID uint32) bool {
    return p.FindItem(itemID) != -1
}

func (p *Player) GetInventorySize() int {
    return len(p.Inventory.Items)
}

func (p *Player) GetItemInSlot(slot int) ItemStack {
    if slot >= 0 && slot < len(p.Inventory.Items) {
        return p.Inventory.Items[slot]
    }
    return ItemStack{ID: 0, Count: 0, Data: 0}
}

func (p *Player) SetItemInSlot(slot int, item ItemStack) {
    if slot >= 0 && slot < len(p.Inventory.Items) {
        p.Inventory.Items[slot] = item
    }
}

func (p *Player) CanFly() bool {
    return p.Gamemode == 1 || p.Gamemode == 2 || p.Abilities.MayFly
}

func (p *Player) IsCreative() bool {
    return p.Gamemode == 1
}

func (p *Player) IsSurvival() bool {
    return p.Gamemode == 0
}

func (p *Player) IsAdventure() bool {
    return p.Gamemode == 2
}

func (p *Player) IsSpectator() bool {
    return p.Gamemode == 3
}