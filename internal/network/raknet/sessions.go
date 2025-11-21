package raknet

func (s *Session) sendStartGamePacket() {
    // Generate or get entity ID
    s.entityID = generateEntityID()
    
    startGamePacket := &minecraft.StartGamePacket{
        EntityID:              s.entityID,
        RuntimeEntityID:       uint64(s.entityID),
        PlayerGameType:        1, // Survival
        PlayerPosition:        minecraft.Vector3{X: 0, Y: 70, Z: 0},
        Rotation:             minecraft.Vector2{X: 0, Y: 0},
        Seed:                 s.world.GetSeed(),
        BiomeType:            1,
        BiomeName:            "plains", 
        Dimension:            0, // Overworld
        Generator:            1, // Flat
        WorldGameMode:        1, // Survival
        Difficulty:           2, // Normal
        SpawnPosition:        minecraft.BlockPos{X: 0, Y: 70, Z: 0},
        AchievementsDisabled: true,
        Time:                 0,
        EduMode:              false,
        CommandsEnabled:      true,
        TexturePacksRequired: false,
        GameRules:           []minecraft.GameRule{},
        Experiments:         []minecraft.Experiment{},
        ChunkRadius:         4,
    }
    
    packetData, err := minecraft.EncodeStartGamePacket(startGamePacket)
    if err != nil {
        log.Printf("Error encoding start game packet: %v", err)
        return
    }
    
    s.SendGamePacket(packetData)
    
    // Create player entity in world
    s.createPlayerInWorld()
    
    // Send existing entities to player
    s.sendExistingEntities()
}

func (s *Session) sendExistingEntities() {
    entities := s.world.GetEntityManager().GetEntities()
    for _, entity := range entities {
        if entity.Type == world.EntityPlayer && entity.EntityID == s.entityID {
            continue // Skip self
        }
        
        s.sendEntityToClient(entity)
    }
}

func (s *Session) sendEntityToClient(entity *world.Entity) {
    var packetData []byte
    var err error
    
    switch entity.Type {
    case world.EntityItem:
        packetData, err = s.encodeItemEntityPacket(entity)
    default:
        packetData, err = s.encodeMobEntityPacket(entity)
    }
    
    if err != nil {
        log.Printf("Error encoding entity packet: %v", err)
        return
    }
    
    s.SendGamePacket(packetData)
}

func (s *Session) encodeMobEntityPacket(entity *world.Entity) ([]byte, error) {
    entityType := s.getEntityTypeString(entity.Type)
    
    packet := &minecraft.AddActorPacket{
        RuntimeID: entity.RuntimeID,
        Type:      entityType,
        Position:  entity.Position,
        Motion:    entity.Motion,
        Rotation:  entity.Rotation,
        Attributes: []minecraft.EntityAttribute{
            {
                Name:      "minecraft:health",
                MinValue:  0.0,
                MaxValue:  entity.MaxHealth,
                Value:     entity.Health,
                Default:   entity.MaxHealth,
            },
        },
    }
    
    return minecraft.EncodeAddActorPacket(packet)
}

func (s *Session) encodeItemEntityPacket(entity *world.Entity) ([]byte, error) {
    itemData, ok := entity.Data.(*world.ItemEntityData)
    if !ok {
        return nil, fmt.Errorf("entity is not an item entity")
    }
    
    packet := &minecraft.AddItemActorPacket{
        RuntimeID: entity.RuntimeID,
        Item: &minecraft.ItemStack{
            ID:     uint16(itemData.ItemStack.ID),
            Count:  itemData.ItemStack.Count,
            Damage: uint16(itemData.ItemStack.Damage),
        },
        Position: entity.Position,
        Motion:   entity.Motion,
    }
    
    // Use AddActor packet for item entities for now
    // TODO: Implement proper AddItemActor packet encoding
    return s.encodeMobEntityPacket(entity)
}

func (s *Session) getEntityTypeString(entityType world.EntityType) string {
    switch entityType {
    case world.EntityZombie:
        return "minecraft:zombie"
    case world.EntitySkeleton:
        return "minecraft:skeleton"
    case world.EntityCreeper:
        return "minecraft:creeper"
    case world.EntitySpider:
        return "minecraft:spider"
    case world.EntityCow:
        return "minecraft:cow"
    case world.EntityPig:
        return "minecraft:pig"
    case world.EntitySheep:
        return "minecraft:sheep"
    case world.EntityChicken:
        return "minecraft:chicken"
    case world.EntityItem:
        return "minecraft:item"
    default:
        return "minecraft:zombie"
    }
}