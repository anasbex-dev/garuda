package world

import (
    "sync"
)

type Inventory struct {
    items      []*ItemStack
    heldSlot   int
    size       int
    mutex      sync.RWMutex
}

type ItemStack struct {
    ID     uint16
    Count  byte
    Damage uint16
    NBT    map[string]interface{}
}

type ItemRegistry struct {
    items map[uint16]*ItemType
    mutex sync.RWMutex
}

type ItemType struct {
    ID          uint16
    Name        string
    MaxStackSize byte
    Durability  uint16
}

func NewInventory(size int) *Inventory {
    items := make([]*ItemStack, size)
    for i := range items {
        items[i] = &ItemStack{ID: 0, Count: 0} // Empty slots
    }
    
    return &Inventory{
        items:    items,
        size:     size,
        heldSlot: 0,
    }
}

func (inv *Inventory) GetItem(slot int) *ItemStack {
    inv.mutex.RLock()
    defer inv.mutex.RUnlock()
    
    if slot < 0 || slot >= inv.size {
        return nil
    }
    return inv.items[slot]
}

func (inv *Inventory) SetItem(slot int, item *ItemStack) bool {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    if slot < 0 || slot >= inv.size {
        return false
    }
    
    inv.items[slot] = item
    return true
}

func (inv *Inventory) AddItem(item *ItemStack) bool {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    // Try to stack with existing items first
    for i, existing := range inv.items {
        if existing.ID == item.ID && existing.Damage == item.Damage && existing.Count < 64 {
            space := 64 - existing.Count
            if item.Count <= space {
                existing.Count += item.Count
                return true
            } else {
                existing.Count = 64
                item.Count -= space
            }
        }
    }
    
    // Find empty slot
    for i, existing := range inv.items {
        if existing.ID == 0 {
            inv.items[i] = item
            return true
        }
    }
    
    return false
}

func (inv *Inventory) RemoveItem(slot int, count byte) *ItemStack {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    if slot < 0 || slot >= inv.size {
        return nil
    }
    
    item := inv.items[slot]
    if item.ID == 0 || item.Count == 0 {
        return nil
    }
    
    if count >= item.Count {
        // Remove entire stack
        removed := &ItemStack{
            ID:     item.ID,
            Count:  item.Count,
            Damage: item.Damage,
            NBT:    item.NBT,
        }
        inv.items[slot] = &ItemStack{ID: 0, Count: 0}
        return removed
    } else {
        // Remove partial stack
        item.Count -= count
        return &ItemStack{
            ID:     item.ID,
            Count:  count,
            Damage: item.Damage,
            NBT:    item.NBT,
        }
    }
}

func (inv *Inventory) GetHeldItem() *ItemStack {
    return inv.GetItem(inv.heldSlot)
}

func (inv *Inventory) SetHeldSlot(slot int) bool {
    if slot < 0 || slot >= 9 {
        return false
    }
    inv.heldSlot = slot
    return true
}

func (inv *Inventory) SwapSlots(slot1, slot2 int) bool {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    if slot1 < 0 || slot1 >= inv.size || slot2 < 0 || slot2 >= inv.size {
        return false
    }
    
    inv.items[slot1], inv.items[slot2] = inv.items[slot2], inv.items[slot1]
    return true
}

func NewItemRegistry() *ItemRegistry {
    registry := &ItemRegistry{
        items: make(map[uint16]*ItemType),
    }
    
    registry.RegisterDefaultItems()
    return registry
}

func (r *ItemRegistry) RegisterItem(item *ItemType) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    r.items[item.ID] = item
}

func (r *ItemRegistry) GetItem(id uint16) *ItemType {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    return r.items[id]
}

func (r *ItemRegistry) RegisterDefaultItems() {
    // Tools
    r.RegisterItem(&ItemType{ID: 256, Name: "iron_shovel", MaxStackSize: 1, Durability: 251})
    r.RegisterItem(&ItemType{ID: 257, Name: "iron_pickaxe", MaxStackSize: 1, Durability: 251})
    r.RegisterItem(&ItemType{ID: 258, Name: "iron_axe", MaxStackSize: 1, Durability: 251})
    
    // Weapons
    r.RegisterItem(&ItemType{ID: 267, Name: "iron_sword", MaxStackSize: 1, Durability: 251})
    r.RegisterItem(&ItemType{ID: 268, Name: "wooden_sword", MaxStackSize: 1, Durability: 59})
    
    // Blocks
    r.RegisterItem(&ItemType{ID: 1, Name: "stone", MaxStackSize: 64, Durability: 0})
    r.RegisterItem(&ItemType{ID: 2, Name: "grass", MaxStackSize: 64, Durability: 0})
    r.RegisterItem(&ItemType{ID: 3, Name: "dirt", MaxStackSize: 64, Durability: 0})
    r.RegisterItem(&ItemType{ID: 4, Name: "cobblestone", MaxStackSize: 64, Durability: 0})
    r.RegisterItem(&ItemType{ID: 5, Name: "oak_wood", MaxStackSize: 64, Durability: 0})
    r.RegisterItem(&ItemType{ID: 6, Name: "oak_planks", MaxStackSize: 64, Durability: 0})
    
    // Resources
    r.RegisterItem(&ItemType{ID: 263, Name: "coal", MaxStackSize: 64, Durability: 0})
    r.RegisterItem(&ItemType{ID: 265, Name: "iron_ingot", MaxStackSize: 64, Durability: 0})
    r.RegisterItem(&ItemType{ID: 266, Name: "gold_ingot", MaxStackSize: 64, Durability: 0})
    r.RegisterItem(&ItemType{ID: 264, Name: "diamond", MaxStackSize: 64, Durability: 0})
}

var DefaultItemRegistry = NewItemRegistry()