package world

type BlockRegistry struct {
    blocks map[uint16]*BlockType
    mutex  sync.RWMutex
}

type BlockType struct {
    ID          uint16
    Name        string
    Solid       bool
    Transparent bool
    LightLevel  byte
}

func NewBlockRegistry() *BlockRegistry {
    registry := &BlockRegistry{
        blocks: make(map[uint16]*BlockType),
    }
    
    registry.RegisterDefaultBlocks()
    return registry
}

func (r *BlockRegistry) RegisterBlock(block *BlockType) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    r.blocks[block.ID] = block
}

func (r *BlockRegistry) GetBlock(id uint16) *BlockType {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    return r.blocks[id]
}

func (r *BlockRegistry) RegisterDefaultBlocks() {
    // Air
    r.RegisterBlock(&BlockType{ID: 0, Name: "air", Solid: false, Transparent: true, LightLevel: 0})
    
    // Stone and ores
    r.RegisterBlock(&BlockType{ID: 1, Name: "stone", Solid: true, Transparent: false, LightLevel: 0})
    r.RegisterBlock(&BlockType{ID: 2, Name: "grass", Solid: true, Transparent: false, LightLevel: 0})
    r.RegisterBlock(&BlockType{ID: 3, Name: "dirt", Solid: true, Transparent: false, LightLevel: 0})
    r.RegisterBlock(&BlockType{ID: 4, Name: "cobblestone", Solid: true, Transparent: false, LightLevel: 0})
    
    // Wood and planks
    r.RegisterBlock(&BlockType{ID: 5, Name: "oak_wood", Solid: true, Transparent: false, LightLevel: 0})
    r.RegisterBlock(&BlockType{ID: 6, Name: "oak_planks", Solid: true, Transparent: false, LightLevel: 0})
    
    // Liquid
    r.RegisterBlock(&BlockType{ID: 8, Name: "water", Solid: false, Transparent: true, LightLevel: 0})
    r.RegisterBlock(&BlockType{ID: 9, Name: "water_stationary", Solid: false, Transparent: true, LightLevel: 0})
    r.RegisterBlock(&BlockType{ID: 10, Name: "lava", Solid: false, Transparent: true, LightLevel: 15})
    r.RegisterBlock(&BlockType{ID: 11, Name: "lava_stationary", Solid: false, Transparent: true, LightLevel: 15})
    
    // Sand and gravel
    r.RegisterBlock(&BlockType{ID: 12, Name: "sand", Solid: true, Transparent: false, LightLevel: 0})
    r.RegisterBlock(&BlockType{ID: 13, Name: "gravel", Solid: true, Transparent: false, LightLevel: 0})
    
    // Glass
    r.RegisterBlock(&BlockType{ID: 20, Name: "glass", Solid: true, Transparent: true, LightLevel: 0})
    
    // Bedrock
    r.RegisterBlock(&BlockType{ID: 7, Name: "bedrock", Solid: true, Transparent: false, LightLevel: 0})
}

var DefaultBlockRegistry = NewBlockRegistry()

func (b Block) IsSolid() bool {
    blockType := DefaultBlockRegistry.GetBlock(b.ID)
    if blockType == nil {
        return false
    }
    return blockType.Solid
}

func (b Block) IsTransparent() bool {
    blockType := DefaultBlockRegistry.GetBlock(b.ID)
    if blockType == nil {
        return true
    }
    return blockType.Transparent
}