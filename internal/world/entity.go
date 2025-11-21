package world

import (
    "math"
    "sync"
    "time"

    "garuda/internal/protocol/minecraft"
)

type EntityType int

const (
    EntityPlayer EntityType = iota
    EntityZombie
    EntitySkeleton
    EntityCreeper
    EntitySpider
    EntityCow
    EntityPig
    EntitySheep
    EntityChicken
    EntityItem
    EntityArrow
)

type Entity struct {
    EntityID      int64
    RuntimeID     uint64
    Type          EntityType
    Position      minecraft.Vector3
    Rotation      minecraft.Vector2
    Motion        minecraft.Vector3
    Health        float32
    MaxHealth     float32
    OnGround      bool
    Age           int64
    Dead          bool
    Data          interface{} // Entity-specific data
    mutex         sync.RWMutex
}

type EntityManager struct {
    entities    map[int64]*Entity
    nextID      int64
    mutex       sync.RWMutex
    world       *World
}

type MobData struct {
    AttackDamage float32
    AttackSpeed  float32
    FollowRange  float32
    MovementSpeed float32
    Hostile      bool
}

type ItemEntityData struct {
    ItemStack *ItemStack
    PickupDelay int64
}

func NewEntityManager(world *World) *EntityManager {
    return &EntityManager{
        entities: make(map[int64]*Entity),
        nextID:   1,
        world:    world,
    }
}

func (em *EntityManager) CreateEntity(entityType EntityType, position minecraft.Vector3) *Entity {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    entity := &Entity{
        EntityID:  em.nextID,
        RuntimeID: uint64(em.nextID),
        Type:      entityType,
        Position:  position,
        Rotation:  minecraft.Vector2{X: 0, Y: 0},
        Motion:    minecraft.Vector3{X: 0, Y: 0, Z: 0},
        Health:    1.0,
        MaxHealth: 1.0,
        OnGround:  false,
        Age:       0,
        Dead:      false,
    }

    // Set entity-specific data
    switch entityType {
    case EntityZombie:
        entity.Health = 20.0
        entity.MaxHealth = 20.0
        entity.Data = &MobData{
            AttackDamage: 3.0,
            AttackSpeed:  1.0,
            FollowRange:  35.0,
            MovementSpeed: 0.23,
            Hostile:      true,
        }
    case EntitySkeleton:
        entity.Health = 20.0
        entity.MaxHealth = 20.0
        entity.Data = &MobData{
            AttackDamage: 2.0,
            AttackSpeed:  1.0,
            FollowRange:  16.0,
            MovementSpeed: 0.25,
            Hostile:      true,
        }
    case EntityCreeper:
        entity.Health = 20.0
        entity.MaxHealth = 20.0
        entity.Data = &MobData{
            AttackDamage: 3.0,
            AttackSpeed:  1.0,
            FollowRange:  16.0,
            MovementSpeed: 0.2,
            Hostile:      true,
        }
    case EntityCow, EntityPig, EntitySheep:
        entity.Health = 10.0
        entity.MaxHealth = 10.0
        entity.Data = &MobData{
            AttackDamage: 0.0,
            AttackSpeed:  0.0,
            FollowRange:  16.0,
            MovementSpeed: 0.25,
            Hostile:      false,
        }
    case EntityItem:
        entity.Health = 1.0
        entity.MaxHealth = 1.0
        entity.Data = &ItemEntityData{
            ItemStack:   &ItemStack{ID: 1, Count: 1},
            PickupDelay: 40, // 2 seconds at 20 ticks/sec
        }
    }

    em.entities[entity.EntityID] = entity
    em.nextID++

    return entity
}

func (em *EntityManager) CreateItemEntity(position minecraft.Vector3, item *ItemStack) *Entity {
    entity := em.CreateEntity(EntityItem, position)
    if itemData, ok := entity.Data.(*ItemEntityData); ok {
        itemData.ItemStack = item
    }
    return entity
}

func (em *EntityManager) RemoveEntity(entityID int64) {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    delete(em.entities, entityID)
}

func (em *EntityManager) GetEntity(entityID int64) *Entity {
    em.mutex.RLock()
    defer em.mutex.RUnlock()

    return em.entities[entityID]
}

func (em *EntityManager) GetEntities() []*Entity {
    em.mutex.RLock()
    defer em.mutex.RUnlock()

    entities := make([]*Entity, 0, len(em.entities))
    for _, entity := range em.entities {
        entities = append(entities, entity)
    }
    return entities
}

func (em *EntityManager) GetEntitiesInRange(position minecraft.Vector3, radius float32) []*Entity {
    em.mutex.RLock()
    defer em.mutex.RUnlock()

    var nearby []*Entity
    for _, entity := range em.entities {
        dx := entity.Position.X - position.X
        dy := entity.Position.Y - position.Y
        dz := entity.Position.Z - position.Z
        distance := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
        
        if distance <= radius {
            nearby = append(nearby, entity)
        }
    }
    return nearby
}

func (em *EntityManager) Update() {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    for _, entity := range em.entities {
        if entity.Dead {
            continue
        }

        em.updateEntity(entity)
    }
}

func (em *EntityManager) updateEntity(entity *Entity) {
    entity.Age++

    switch entity.Type {
    case EntityItem:
        em.updateItemEntity(entity)
    case EntityZombie, EntitySkeleton, EntityCreeper, EntitySpider:
        em.updateHostileMob(entity)
    case EntityCow, EntityPig, EntitySheep, EntityChicken:
        em.updatePassiveMob(entity)
    }
}

func (em *EntityManager) updateItemEntity(entity *Entity) {
    if itemData, ok := entity.Data.(*ItemEntityData); ok {
        // Decrease pickup delay
        if itemData.PickupDelay > 0 {
            itemData.PickupDelay--
        }

        // Apply gravity
        if !entity.OnGround {
            entity.Motion.Y -= 0.04 // Gravity
        }

        // Apply motion
        entity.Position.X += entity.Motion.X
        entity.Position.Y += entity.Motion.Y
        entity.Position.Z += entity.Motion.Z

        // Check for ground collision
        if em.checkEntityCollision(entity) {
            entity.OnGround = true
            entity.Motion.Y = 0
        } else {
            entity.OnGround = false
        }

        // Apply friction
        entity.Motion.X *= 0.98
        entity.Motion.Z *= 0.98
        if entity.OnGround {
            entity.Motion.X *= 0.6
            entity.Motion.Z *= 0.6
        }
    }
}

func (em *EntityManager) updateHostileMob(entity *Entity) {
    mobData, ok := entity.Data.(*MobData)
    if !ok {
        return
    }

    // Simple AI: find nearest player and move towards them
    nearestPlayer := em.findNearestPlayer(entity.Position, mobData.FollowRange)
    if nearestPlayer != nil {
        // Move towards player
        dx := nearestPlayer.Position.X - entity.Position.X
        dz := nearestPlayer.Position.Z - entity.Position.Z
        distance := float32(math.Sqrt(float64(dx*dx + dz*dz)))

        if distance > 2.0 { // Stop when close enough to attack
            // Normalize direction
            if distance > 0 {
                dx /= distance
                dz /= distance
            }

            // Apply movement
            entity.Motion.X = dx * mobData.MovementSpeed
            entity.Motion.Z = dz * mobData.MovementSpeed

            // Update rotation to face player
            yaw := float32(math.Atan2(float64(dz), float64(dx)))*180/math.Pi - 90
            entity.Rotation.Y = yaw
        }

        // Apply gravity and motion
        if !entity.OnGround {
            entity.Motion.Y -= 0.04
        }

        entity.Position.X += entity.Motion.X
        entity.Position.Y += entity.Motion.Y
        entity.Position.Z += entity.Motion.Z

        // Check collision
        if em.checkEntityCollision(entity) {
            entity.OnGround = true
            entity.Motion.Y = 0
        } else {
            entity.OnGround = false
        }
    }
}

func (em *EntityManager) updatePassiveMob(entity *Entity) {
    mobData, ok := entity.Data.(*MobData)
    if !ok {
        return
    }

    // Simple wandering behavior
    if entity.Age%100 == 0 { // Change direction every 5 seconds
        angle := float64(entity.Age) * 0.1
        entity.Motion.X = float32(math.Cos(angle)) * mobData.MovementSpeed * 0.5
        entity.Motion.Z = float32(math.Sin(angle)) * mobData.MovementSpeed * 0.5
        entity.Rotation.Y = float32(angle) * 180 / math.Pi
    }

    // Apply gravity and motion
    if !entity.OnGround {
        entity.Motion.Y -= 0.04
    }

    entity.Position.X += entity.Motion.X
    entity.Position.Y += entity.Motion.Y
    entity.Position.Z += entity.Motion.Z

    // Check collision
    if em.checkEntityCollision(entity) {
        entity.OnGround = true
        entity.Motion.Y = 0
    } else {
        entity.OnGround = false
    }
}

func (em *EntityManager) findNearestPlayer(position minecraft.Vector3, maxRange float32) *Player {
    var nearest *Player
    var nearestDistance float32 = maxRange

    for _, player := range em.world.players {
        dx := player.Position.X - position.X
        dy := player.Position.Y - position.Y
        dz := player.Position.Z - position.Z
        distance := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
        
        if distance < nearestDistance {
            nearest = player
            nearestDistance = distance
        }
    }

    return nearest
}

func (em *EntityManager) checkEntityCollision(entity *Entity) bool {
    // Simple AABB collision check
    aabb := em.getEntityAABB(entity)
    return em.world.CheckCollision(aabb)
}

func (em *EntityManager) getEntityAABB(entity *Entity) AABB {
    var width, height float64

    switch entity.Type {
    case EntityPlayer:
        width, height = 0.6, 1.8
    case EntityZombie, EntitySkeleton:
        width, height = 0.6, 1.95
    case EntityCreeper:
        width, height = 0.6, 1.7
    case EntitySpider:
        width, height = 1.4, 0.9
    case EntityCow, EntityPig, EntitySheep:
        width, height = 0.9, 1.4
    case EntityChicken:
        width, height = 0.4, 0.7
    case EntityItem:
        width, height = 0.25, 0.25
    default:
        width, height = 0.6, 1.8
    }

    return AABB{
        MinX: float64(entity.Position.X) - width/2,
        MinY: float64(entity.Position.Y),
        MinZ: float64(entity.Position.Z) - width/2,
        MaxX: float64(entity.Position.X) + width/2,
        MaxY: float64(entity.Position.Y) + height,
        MaxZ: float64(entity.Position.Z) + width/2,
    }
}

func (em *EntityManager) DamageEntity(entity *Entity, damage float32, source *Entity) {
    entity.mutex.Lock()
    defer entity.mutex.Unlock()

    entity.Health -= damage
    if entity.Health <= 0 {
        entity.Dead = true
        em.onEntityDeath(entity, source)
    }
}

func (em *EntityManager) onEntityDeath(entity *Entity, killer *Entity) {
    switch entity.Type {
    case EntityZombie, EntitySkeleton, EntityCreeper, EntitySpider:
        // Drop experience and possibly items
        em.createExperienceOrbs(entity.Position, 5)
        
        // Random item drops
        if em.world.random.Float32() < 0.5 {
            em.CreateItemEntity(entity.Position, &ItemStack{ID: 265, Count: 1}) // Iron ingot
        }
        
    case EntityCow:
        em.CreateItemEntity(entity.Position, &ItemStack{ID: 334, Count: 1}) // Leather
        em.CreateItemEntity(entity.Position, &ItemStack{ID: 363, Count: 2}) // Raw beef
        
    case EntityPig:
        em.CreateItemEntity(entity.Position, &ItemStack{ID: 319, Count: 1}) // Raw porkchop
        
    case EntitySheep:
        em.CreateItemEntity(entity.Position, &ItemStack{ID: 35, Count: 1}) // Wool
        
    case EntityChicken:
        em.CreateItemEntity(entity.Position, &ItemStack{ID: 365, Count: 1}) // Raw chicken
        if em.world.random.Float32() < 0.3 {
            em.CreateItemEntity(entity.Position, &ItemStack{ID: 344, Count: 1}) // Feather
        }
        
    case EntityItem:
        // Item entities don't drop anything on death
    }
}

func (em *EntityManager) createExperienceOrbs(position minecraft.Vector3, amount int) {
    // Create experience orb entities
    for i := 0; i < amount; i++ {
        // TODO: Implement experience orb entities
    }
}