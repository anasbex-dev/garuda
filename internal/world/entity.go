package world

import (
    "garuda/pkg/utils"
    "sync"
    "time"
)

const (
    EntityPlayer = iota
    EntityItem
    EntityZombie
    EntitySkeleton
    EntityCreeper
)

type Entity struct {
    ID        int64
    Type      int
    Position  [3]float32
    Rotation  [2]float32
    Velocity  [3]float32
    Health    float32
    MaxHealth float32
    Metadata  map[string]interface{}
    World     *World
    mutex     sync.RWMutex
}

type EntityManager struct {
    entities map[int64]*Entity
    nextID   int64
    mutex    sync.RWMutex
    logger   *utils.Logger
}

func NewEntityManager(logger *utils.Logger) *EntityManager {
    return &EntityManager{
        entities: make(map[int64]*Entity),
        nextID:   1,
        logger:   logger,
    }
}

func (em *EntityManager) NewEntity(entityType int, world *World, position [3]float32) *Entity {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    entity := &Entity{
        ID:        em.nextID,
        Type:      entityType,
        Position:  position,
        Rotation:  [2]float32{0, 0},
        Velocity:  [3]float32{0, 0, 0},
        Health:    20.0,
        MaxHealth: 20.0,
        Metadata:  make(map[string]interface{}),
        World:     world,
    }

    em.setDefaultMetadata(entity)
    em.entities[em.nextID] = entity
    em.nextID++

    return entity
}

func (em *EntityManager) setDefaultMetadata(entity *Entity) {
    switch entity.Type {
    case EntityZombie:
        entity.Metadata["name"] = "Zombie"
        entity.Metadata["scale"] = float32(1.0)
        entity.Health = 20.0
        entity.MaxHealth = 20.0
    case EntitySkeleton:
        entity.Metadata["name"] = "Skeleton"
        entity.Metadata["scale"] = float32(1.0)
        entity.Health = 20.0
        entity.MaxHealth = 20.0
    case EntityCreeper:
        entity.Metadata["name"] = "Creeper"
        entity.Metadata["scale"] = float32(1.0)
        entity.Health = 20.0
        entity.MaxHealth = 20.0
    case EntityItem:
        entity.Metadata["name"] = "Item"
        entity.Metadata["scale"] = float32(0.5)
        entity.Health = 1.0
        entity.MaxHealth = 1.0
    }
}

func (em *EntityManager) GetEntity(id int64) *Entity {
    em.mutex.RLock()
    defer em.mutex.RUnlock()
    return em.entities[id]
}

func (em *EntityManager) RemoveEntity(id int64) {
    em.mutex.Lock()
    defer em.mutex.Unlock()
    delete(em.entities, id)
}

func (em *EntityManager) UpdateEntities() {
    em.mutex.RLock()
    entities := make([]*Entity, 0, len(em.entities))
    for _, entity := range em.entities {
        entities = append(entities, entity)
    }
    em.mutex.RUnlock()

    for _, entity := range entities {
        em.updateEntity(entity)
    }
}

func (em *EntityManager) updateEntity(entity *Entity) {
    entity.mutex.Lock()
    defer entity.mutex.Unlock()

    if entity.Type == EntityItem {
        return
    }

    entity.Position[0] += entity.Velocity[0]
    entity.Position[1] += entity.Velocity[1]
    entity.Position[2] += entity.Velocity[2]

    entity.Velocity[0] *= 0.98
    entity.Velocity[1] *= 0.98
    entity.Velocity[2] *= 0.98

    if entity.Velocity[1] > -0.08 {
        entity.Velocity[1] -= 0.08
    }

    onGround := em.checkOnGround(entity)
    if onGround {
        entity.Velocity[1] = 0
    }
}

func (em *EntityManager) checkOnGround(entity *Entity) bool {
    checkX := int(entity.Position[0])
    checkY := int(entity.Position[1]) - 1
    checkZ := int(entity.Position[2])

    block := entity.World.GetBlock(checkX, checkY, checkZ)
    return block.ID != 0
}

func (em *EntityManager) SpawnMob(mobType int, world *World, position [3]float32) *Entity {
    entity := em.NewEntity(mobType, world, position)
    em.logger.Debug("Spawned mob %d at %.1f,%.1f,%.1f", mobType, position[0], position[1], position[2])
    return entity
}

func (em *EntityManager) SpawnItem(itemID uint32, world *World, position [3]float32) *Entity {
    entity := em.NewEntity(EntityItem, world, position)
    entity.Metadata["item_id"] = itemID
    entity.Metadata["item_count"] = byte(1)
    em.logger.Debug("Spawned item %d at %.1f,%.1f,%.1f", itemID, position[0], position[1], position[2])
    return entity
}

func (em *EntityManager) GetEntitiesInRange(position [3]float32, radius float32) []*Entity {
    em.mutex.RLock()
    defer em.mutex.RUnlock()

    var result []*Entity
    for _, entity := range em.entities {
        dx := entity.Position[0] - position[0]
        dy := entity.Position[1] - position[1]
        dz := entity.Position[2] - position[2]
        distance := dx*dx + dy*dy + dz*dz

        if distance <= radius*radius {
            result = append(result, entity)
        }
    }
    return result
}

func (e *Entity) GetPosition() [3]float32 {
    e.mutex.RLock()
    defer e.mutex.RUnlock()
    return e.Position
}

func (e *Entity) SetPosition(position [3]float32) {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.Position = position
}

func (e *Entity) GetRotation() [2]float32 {
    e.mutex.RLock()
    defer e.mutex.RUnlock()
    return e.Rotation
}

func (e *Entity) SetRotation(rotation [2]float32) {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.Rotation = rotation
}

func (e *Entity) GetVelocity() [3]float32 {
    e.mutex.RLock()
    defer e.mutex.RUnlock()
    return e.Velocity
}

func (e *Entity) SetVelocity(velocity [3]float32) {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.Velocity = velocity
}

func (e *Entity) Damage(amount float32) {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.Health -= amount
    if e.Health < 0 {
        e.Health = 0
    }
}

func (e *Entity) Heal(amount float32) {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.Health += amount
    if e.Health > e.MaxHealth {
        e.Health = e.MaxHealth
    }
}

func (e *Entity) IsAlive() bool {
    e.mutex.RLock()
    defer e.mutex.RUnlock()
    return e.Health > 0
}