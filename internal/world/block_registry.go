package world

type Block struct {
    ID          uint32
    Name        string
    DisplayName string
    Hardness    float32
    Resistance  float32
    StackSize   uint16
    Diggable    bool
    Transparent bool
    FilterLight int
    EmitLight   int
    BoundingBox string
    Material    string
    HarvestTools map[uint32]bool
}

type BlockRegistry struct {
    blocks map[uint32]*Block
    nameToID map[string]uint32
}

func NewBlockRegistry() *BlockRegistry {
    registry := &BlockRegistry{
        blocks:   make(map[uint32]*Block),
        nameToID: make(map[string]uint32),
    }
    
    registry.registerBlocks()
    return registry
}

func (br *BlockRegistry) registerBlocks() {
    blocks := []*Block{
        // Air
        {ID: 0, Name: "minecraft:air", DisplayName: "Air", Hardness: 0, Resistance: 0, StackSize: 0, Diggable: false, Transparent: true, FilterLight: 0, EmitLight: 0, BoundingBox: "empty", Material: "air"},
        
        // Stone types
        {ID: 1, Name: "minecraft:stone", DisplayName: "Stone", Hardness: 1.5, Resistance: 6.0, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "rock", HarvestTools: map[uint32]bool{270: true, 274: true, 257: true, 278: true}},
        {ID: 2, Name: "minecraft:grass", DisplayName: "Grass Block", Hardness: 0.6, Resistance: 0.6, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "grass", HarvestTools: map[uint32]bool{270: true, 274: true, 256: true, 279: true}},
        {ID: 3, Name: "minecraft:dirt", DisplayName: "Dirt", Hardness: 0.5, Resistance: 0.5, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "dirt", HarvestTools: map[uint32]bool{270: true, 274: true, 256: true, 279: true}},
        
        // Wood and plants
        {ID: 5, Name: "minecraft:oak_planks", DisplayName: "Oak Planks", Hardness: 2.0, Resistance: 3.0, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "wood", HarvestTools: map[uint32]bool{270: true, 274: true, 258: true, 279: true}},
        {ID: 6, Name: "minecraft:oak_sapling", DisplayName: "Oak Sapling", Hardness: 0.0, Resistance: 0.0, StackSize: 64, Diggable: true, Transparent: true, FilterLight: 0, EmitLight: 0, BoundingBox: "empty", Material: "plant"},
        {ID: 7, Name: "minecraft:bedrock", DisplayName: "Bedrock", Hardness: -1.0, Resistance: 18000000.0, StackSize: 64, Diggable: false, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "rock"},
        
        // Liquids
        {ID: 8, Name: "minecraft:flowing_water", DisplayName: "Water", Hardness: 100.0, Resistance: 100.0, StackSize: 0, Diggable: false, Transparent: true, FilterLight: 1, EmitLight: 0, BoundingBox: "empty", Material: "water"},
        {ID: 9, Name: "minecraft:water", DisplayName: "Water", Hardness: 100.0, Resistance: 100.0, StackSize: 0, Diggable: false, Transparent: true, FilterLight: 1, EmitLight: 0, BoundingBox: "empty", Material: "water"},
        {ID: 10, Name: "minecraft:flowing_lava", DisplayName: "Lava", Hardness: 100.0, Resistance: 100.0, StackSize: 0, Diggable: false, Transparent: true, FilterLight: 0, EmitLight: 15, BoundingBox: "empty", Material: "lava"},
        {ID: 11, Name: "minecraft:lava", DisplayName: "Lava", Hardness: 100.0, Resistance: 100.0, StackSize: 0, Diggable: false, Transparent: true, FilterLight: 0, EmitLight: 15, BoundingBox: "empty", Material: "lava"},
        
        // Ores
        {ID: 14, Name: "minecraft:gold_ore", DisplayName: "Gold Ore", Hardness: 3.0, Resistance: 3.0, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "rock", HarvestTools: map[uint32]bool{257: true, 278: true}},
        {ID: 15, Name: "minecraft:iron_ore", DisplayName: "Iron Ore", Hardness: 3.0, Resistance: 3.0, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "rock", HarvestTools: map[uint32]bool{257: true, 278: true}},
        {ID: 16, Name: "minecraft:coal_ore", DisplayName: "Coal Ore", Hardness: 3.0, Resistance: 3.0, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "rock", HarvestTools: map[uint32]bool{270: true, 274: true, 257: true, 278: true}},
        
        // Logs
        {ID: 17, Name: "minecraft:oak_log", DisplayName: "Oak Log", Hardness: 2.0, Resistance: 2.0, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "wood", HarvestTools: map[uint32]bool{270: true, 274: true, 258: true, 279: true}},
        
        // Leaves
        {ID: 18, Name: "minecraft:oak_leaves", DisplayName: "Oak Leaves", Hardness: 0.2, Resistance: 0.2, StackSize: 64, Diggable: true, Transparent: true, FilterLight: 1, EmitLight: 0, BoundingBox: "block", Material: "leaves", HarvestTools: map[uint32]bool{270: true, 274: true, 258: true, 279: true}},
        
        // Glass
        {ID: 20, Name: "minecraft:glass", DisplayName: "Glass", Hardness: 0.3, Resistance: 0.3, StackSize: 64, Diggable: true, Transparent: true, FilterLight: 0, EmitLight: 0, BoundingBox: "block", Material: "glass"},
        
        // Flowers and plants
        {ID: 31, Name: "minecraft:grass", DisplayName: "Grass", Hardness: 0.0, Resistance: 0.0, StackSize: 64, Diggable: true, Transparent: true, FilterLight: 0, EmitLight: 0, BoundingBox: "empty", Material: "plant"},
        {ID: 37, Name: "minecraft:dandelion", DisplayName: "Dandelion", Hardness: 0.0, Resistance: 0.0, StackSize: 64, Diggable: true, Transparent: true, FilterLight: 0, EmitLight: 0, BoundingBox: "empty", Material: "plant"},
        {ID: 38, Name: "minecraft:poppy", DisplayName: "Poppy", Hardness: 0.0, Resistance: 0.0, StackSize: 64, Diggable: true, Transparent: true, FilterLight: 0, EmitLight: 0, BoundingBox: "empty", Material: "plant"},
        
        // Crafting blocks
        {ID: 58, Name: "minecraft:crafting_table", DisplayName: "Crafting Table", Hardness: 2.5, Resistance: 2.5, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "wood", HarvestTools: map[uint32]bool{270: true, 274: true, 258: true, 279: true}},
        
        // Furnace
        {ID: 61, Name: "minecraft:furnace", DisplayName: "Furnace", Hardness: 3.5, Resistance: 3.5, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "rock", HarvestTools: map[uint32]bool{270: true, 274: true, 257: true, 278: true}},
        
        // Chest
        {ID: 54, Name: "minecraft:chest", DisplayName: "Chest", Hardness: 2.5, Resistance: 2.5, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "wood", HarvestTools: map[uint32]bool{270: true, 274: true, 258: true, 279: true}},
        
        // Redstone
        {ID: 73, Name: "minecraft:redstone_ore", DisplayName: "Redstone Ore", Hardness: 3.0, Resistance: 3.0, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "rock", HarvestTools: map[uint32]bool{257: true, 278: true}},
        
        // Diamond
        {ID: 56, Name: "minecraft:diamond_ore", DisplayName: "Diamond Ore", Hardness: 3.0, Resistance: 3.0, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "rock", HarvestTools: map[uint32]bool{257: true, 278: true}},
        
        // TNT
        {ID: 46, Name: "minecraft:tnt", DisplayName: "TNT", Hardness: 0.0, Resistance: 0.0, StackSize: 64, Diggable: true, Transparent: false, FilterLight: 15, EmitLight: 0, BoundingBox: "block", Material: "tnt"},
    }
    
    for _, block := range blocks {
        br.blocks[block.ID] = block
        br.nameToID[block.Name] = block.ID
    }
}

func (br *BlockRegistry) GetBlock(id uint32) *Block {
    return br.blocks[id]
}

func (br *BlockRegistry) GetBlockByName(name string) *Block {
    id, exists := br.nameToID[name]
    if !exists {
        return nil
    }
    return br.blocks[id]
}

func (br *BlockRegistry) IsSolid(id uint32) bool {
    block := br.GetBlock(id)
    if block == nil {
        return false
    }
    return block.BoundingBox == "block" && !block.Transparent
}

func (br *BlockRegistry) IsTransparent(id uint32) bool {
    block := br.GetBlock(id)
    if block == nil {
        return true
    }
    return block.Transparent
}

func (br *BlockRegistry) IsLiquid(id uint32) bool {
    block := br.GetBlock(id)
    if block == nil {
        return false
    }
    return block.Material == "water" || block.Material == "lava"
}

func (br *BlockRegistry) CanHarvestWith(blockID uint32, toolID uint32) bool {
    block := br.GetBlock(blockID)
    if block == nil || block.HarvestTools == nil {
        return true
    }
    
    if toolID == 0 {
        return false
    }
    
    return block.HarvestTools[toolID]
}

func (br *BlockRegistry) GetDigTime(blockID uint32, toolID uint32, correctTool bool) float32 {
    block := br.GetBlock(blockID)
    if block == nil || !block.Diggable {
        return -1.0
    }
    
    baseTime := block.Hardness * 1.5
    
    if correctTool {
        baseTime = block.Hardness * 0.75
    }
    
    if toolID == 0 {
        baseTime *= 3.33
    }
    
    return baseTime
}