package world

import (
    "sort"
)

type InventoryType int

const (
    InventoryPlayer InventoryType = iota
    InventoryChest
    InventoryFurnace
    InventoryCrafting
    InventoryEnchanting
    InventoryEnderChest
)

type Inventory struct {
    Type     InventoryType
    Size     int
    Items    []ItemStack
    Title    string
    mutex    sync.RWMutex
}

type InventoryManager struct {
    inventories map[string]*Inventory
    mutex       sync.RWMutex
    logger      *utils.Logger
}

func NewInventoryManager(logger *utils.Logger) *InventoryManager {
    return &InventoryManager{
        inventories: make(map[string]*Inventory),
        logger:      logger,
    }
}

func NewPlayerInventory() *Inventory {
    inv := &Inventory{
        Type:  InventoryPlayer,
        Size:  36, // 27 inventory + 9 hotbar
        Items: make([]ItemStack, 36),
        Title: "Inventory",
    }
    
    // Initialize empty slots
    for i := range inv.Items {
        inv.Items[i] = ItemStack{ID: 0, Count: 0, Data: 0}
    }
    
    return inv
}

func NewChestInventory(size int, title string) *Inventory {
    if size <= 0 || size > 54 {
        size = 27 // Default chest size
    }
    
    inv := &Inventory{
        Type:  InventoryChest,
        Size:  size,
        Items: make([]ItemStack, size),
        Title: title,
    }
    
    for i := range inv.Items {
        inv.Items[i] = ItemStack{ID: 0, Count: 0, Data: 0}
    }
    
    return inv
}

func (im *InventoryManager) CreateInventory(invType InventoryType, size int, title string) *Inventory {
    var inv *Inventory
    
    switch invType {
    case InventoryPlayer:
        inv = NewPlayerInventory()
    case InventoryChest:
        inv = NewChestInventory(size, title)
    default:
        inv = &Inventory{
            Type:  invType,
            Size:  size,
            Items: make([]ItemStack, size),
            Title: title,
        }
        
        for i := range inv.Items {
            inv.Items[i] = ItemStack{ID: 0, Count: 0, Data: 0}
        }
    }
    
    // Generate unique ID
    invID := im.generateInventoryID()
    
    im.mutex.Lock()
    im.inventories[invID] = inv
    im.mutex.Unlock()
    
    return inv
}

func (im *InventoryManager) generateInventoryID() string {
    // Simple ID generation - bisa diimprove dengan UUID
    return fmt.Sprintf("inv_%d", len(im.inventories)+1)
}

func (im *InventoryManager) GetInventory(invID string) *Inventory {
    im.mutex.RLock()
    defer im.mutex.RUnlock()
    
    return im.inventories[invID]
}

func (im *InventoryManager) RemoveInventory(invID string) {
    im.mutex.Lock()
    defer im.mutex.Unlock()
    
    delete(im.inventories, invID)
}

func (inv *Inventory) GetItem(slot int) ItemStack {
    inv.mutex.RLock()
    defer inv.mutex.RUnlock()
    
    if slot < 0 || slot >= len(inv.Items) {
        return ItemStack{ID: 0, Count: 0, Data: 0}
    }
    
    return inv.Items[slot]
}

func (inv *Inventory) SetItem(slot int, item ItemStack) bool {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    if slot < 0 || slot >= len(inv.Items) {
        return false
    }
    
    inv.Items[slot] = item
    return true
}

func (inv *Inventory) AddItem(item ItemStack) (remaining ItemStack, added bool) {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    remaining = item
    added = false
    
    // First try to stack with existing items
    for i := range inv.Items {
        if inv.Items[i].ID == item.ID && inv.Items[i].Data == item.Data {
            space := 64 - inv.Items[i].Count
            if space > 0 {
                transfer := byte(min(int(item.Count), int(space)))
                inv.Items[i].Count += transfer
                remaining.Count -= transfer
                added = true
                
                if remaining.Count <= 0 {
                    return ItemStack{ID: 0, Count: 0, Data: 0}, true
                }
            }
        }
    }
    
    // Then try to put in empty slots
    for i := range inv.Items {
        if inv.Items[i].ID == 0 {
            inv.Items[i] = remaining
            return ItemStack{ID: 0, Count: 0, Data: 0}, true
        }
    }
    
    return remaining, added
}

func (inv *Inventory) RemoveItem(slot int, count byte) (ItemStack, bool) {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    if slot < 0 || slot >= len(inv.Items) || inv.Items[slot].ID == 0 {
        return ItemStack{ID: 0, Count: 0, Data: 0}, false
    }
    
    item := inv.Items[slot]
    if item.Count <= count {
        inv.Items[slot] = ItemStack{ID: 0, Count: 0, Data: 0}
        return item, true
    } else {
        removed := ItemStack{ID: item.ID, Count: count, Data: item.Data}
        inv.Items[slot].Count -= count
        return removed, true
    }
}

func (inv *Inventory) RemoveItems(itemID uint32, count int) ([]ItemStack, bool) {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    removed := make([]ItemStack, 0)
    remaining := count
    
    for i := range inv.Items {
        if inv.Items[i].ID == itemID && remaining > 0 {
            available := int(inv.Items[i].Count)
            take := min(available, remaining)
            
            removed = append(removed, ItemStack{
                ID:    itemID,
                Count: byte(take),
                Data:  inv.Items[i].Data,
            })
            
            if take == available {
                inv.Items[i] = ItemStack{ID: 0, Count: 0, Data: 0}
            } else {
                inv.Items[i].Count -= byte(take)
            }
            
            remaining -= take
        }
    }
    
    return removed, remaining == 0
}

func (inv *Inventory) HasItem(itemID uint32, count int) bool {
    inv.mutex.RLock()
    defer inv.mutex.RUnlock()
    
    total := 0
    for _, item := range inv.Items {
        if item.ID == itemID {
            total += int(item.Count)
            if total >= count {
                return true
            }
        }
    }
    
    return total >= count
}

func (inv *Inventory) CountItem(itemID uint32) int {
    inv.mutex.RLock()
    defer inv.mutex.RUnlock()
    
    total := 0
    for _, item := range inv.Items {
        if item.ID == itemID {
            total += int(item.Count)
        }
    }
    
    return total
}

func (inv *Inventory) FindItem(itemID uint32) []int {
    inv.mutex.RLock()
    defer inv.mutex.RUnlock()
    
    slots := make([]int, 0)
    for i, item := range inv.Items {
        if item.ID == itemID && item.Count > 0 {
            slots = append(slots, i)
        }
    }
    
    return slots
}

func (inv *Inventory) FindEmptySlot() int {
    inv.mutex.RLock()
    defer inv.mutex.RUnlock()
    
    for i, item := range inv.Items {
        if item.ID == 0 {
            return i
        }
    }
    
    return -1
}

func (inv *Inventory) IsEmpty() bool {
    inv.mutex.RLock()
    defer inv.mutex.RUnlock()
    
    for _, item := range inv.Items {
        if item.ID != 0 {
            return false
        }
    }
    
    return true
}

func (inv *Inventory) GetContents() []ItemStack {
    inv.mutex.RLock()
    defer inv.mutex.RUnlock()
    
    items := make([]ItemStack, len(inv.Items))
    copy(items, inv.Items)
    return items
}

func (inv *Inventory) SetContents(items []ItemStack) {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    copy(inv.Items, items)
}

func (inv *Inventory) SwapSlots(slot1, slot2 int) bool {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    if slot1 < 0 || slot1 >= len(inv.Items) || slot2 < 0 || slot2 >= len(inv.Items) {
        return false
    }
    
    inv.Items[slot1], inv.Items[slot2] = inv.Items[slot2], inv.Items[slot1]
    return true
}

func (inv *Inventory) Sort() {
    inv.mutex.Lock()
    defer inv.mutex.Unlock()
    
    // Sort by item ID, then by data value
    sort.Slice(inv.Items, func(i, j int) bool {
        if inv.Items[i].ID == inv.Items[j].ID {
            return inv.Items[i].Data < inv.Items[j].Data
        }
        return inv.Items[i].ID < inv.Items[j].ID
    })
    
    // Compact stacks
    im.compactStacks(inv)
}

func (im *InventoryManager) compactStacks(inv *Inventory) {
    // Group items by type
    itemsByType := make(map[uint32]map[uint16][]*ItemStack)
    
    for i := range inv.Items {
        item := &inv.Items[i]
        if item.ID == 0 {
            continue
        }
        
        if itemsByType[item.ID] == nil {
            itemsByType[item.ID] = make(map[uint16][]*ItemStack)
        }
        
        itemsByType[item.ID][item.Data] = append(itemsByType[item.ID][item.Data], item)
    }
    
    // Rebuild inventory dengan stacks yang compact
    newItems := make([]ItemStack, len(inv.Items))
    slot := 0
    
    for itemID, dataMap := range itemsByType {
        for dataValue, itemList := range dataMap {
            totalCount := 0
            for _, item := range itemList {
                totalCount += int(item.Count)
            }
            
            // Distribute across stacks
            for totalCount > 0 && slot < len(newItems) {
                stackSize := min(totalCount, 64)
                newItems[slot] = ItemStack{
                    ID:    itemID,
                    Count: byte(stackSize),
                    Data:  dataValue,
                }
                totalCount -= stackSize
                slot++
            }
        }
    }
    
    // Fill remaining slots with air
    for i := slot; i < len(newItems); i++ {
        newItems[i] = ItemStack{ID: 0, Count: 0, Data: 0}
    }
    
    inv.Items = newItems
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}