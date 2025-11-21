package world

import (
    "math"
    "math/rand"
    "time"
)

type DamageSource int

const (
    DamageSourceEntity DamageSource = iota
    DamageSourcePlayer
    DamageSourceProjectile
    DamageSourceExplosion
    DamageSourceFire
    DamageSourceFall
    DamageSourceDrowning
    DamageSourceStarvation
    DamageSourceVoid
)

type CombatEvent struct {
    Attacker    *Entity
    Target      *Entity
    Damage      float32
    Source      DamageSource
    Timestamp   time.Time
    Critical    bool
    Weapon      *ItemStack
}

type CombatManager struct {
    events      []*CombatEvent
    eventMutex  sync.RWMutex
    logger      *utils.Logger
}

func NewCombatManager(logger *utils.Logger) *CombatManager {
    return &CombatManager{
        events: make([]*CombatEvent, 0),
        logger: logger,
    }
}

func (cm *CombatManager) Attack(attacker, target *Entity, weapon *ItemStack) *CombatEvent {
    if attacker == nil || target == nil || !target.IsAlive() {
        return nil
    }

    // Calculate base damage
    baseDamage := cm.calculateBaseDamage(attacker, target, weapon)
    
    // Check for critical hit
    critical := cm.isCriticalHit(attacker)
    if critical {
        baseDamage *= 1.5
    }
    
    // Apply armor reduction
    finalDamage := cm.applyArmorReduction(target, baseDamage)
    
    // Apply enchantment effects
    finalDamage = cm.applyEnchantments(attacker, target, weapon, finalDamage)
    
    // Create combat event
    event := &CombatEvent{
        Attacker:  attacker,
        Target:    target,
        Damage:    finalDamage,
        Source:    DamageSourceEntity,
        Timestamp: time.Now(),
        Critical:  critical,
        Weapon:    weapon,
    }
    
    // Apply damage to target
    target.Damage(finalDamage)
    
    // Knockback effect
    cm.applyKnockback(attacker, target, weapon)
    
    // Add to event history
    cm.eventMutex.Lock()
    cm.events = append(cm.events, event)
    cm.eventMutex.Unlock()
    
    cm.logger.Debug("Combat: %s attacked %s for %.1f damage (critical: %v)", 
        cm.getEntityName(attacker), cm.getEntityName(target), finalDamage, critical)
    
    return event
}

func (cm *CombatManager) calculateBaseDamage(attacker, target *Entity, weapon *ItemStack) float32 {
    baseDamage := float32(1.0) // Default hand damage
    
    if weapon != nil && weapon.ID != 0 {
        // Weapon-based damage
        baseDamage = cm.getWeaponDamage(weapon.ID)
    }
    
    // Entity type modifiers
    if attacker.Type == EntityPlayer {
        baseDamage += cm.getPlayerAttackDamage(attacker)
    }
    
    // Random variation ±10%
    variation := 0.9 + rand.Float32()*0.2
    baseDamage *= variation
    
    return baseDamage
}

func (cm *CombatManager) getWeaponDamage(weaponID uint32) float32 {
    switch weaponID {
    case 268: // Wooden Sword
        return 4.0
    case 272: // Stone Sword
        return 5.0
    case 267: // Iron Sword
        return 6.0
    case 276: // Diamond Sword
        return 7.0
    case 283: // Golden Sword
        return 4.0
    case 258: // Axe (various types)
        return 9.0
    case 270: // Pickaxe (various types)
        return 2.0
    case 290: // Shovel
        return 1.5
    default:
        return 1.0 // Hand or unknown item
    }
}

func (cm *CombatManager) getPlayerAttackDamage(player *Entity) float32 {
    // Base player attack strength
    damage := float32(1.0)
    
    // TODO: Add player stats, effects, etc.
    return damage
}

func (cm *CombatManager) isCriticalHit(attacker *Entity) bool {
    // 5% base critical chance
    baseChance := 0.05
    
    // Player-specific critical logic
    if attacker.Type == EntityPlayer {
        // Check if player is falling for critical hit
        if attacker.Velocity[1] < -0.08 {
            return true
        }
    }
    
    return rand.Float64() < baseChance
}

func (cm *CombatManager) applyArmorReduction(target *Entity, damage float32) float32 {
    armorPoints := cm.getArmorPoints(target)
    if armorPoints <= 0 {
        return damage
    }
    
    // Minecraft armor reduction formula
    reduction := damage * (float32(armorPoints) * 0.04)
    finalDamage := damage - reduction
    
    if finalDamage < damage * 0.2 {
        finalDamage = damage * 0.2 // Minimum 20% damage
    }
    
    return finalDamage
}

func (cm *CombatManager) getArmorPoints(entity *Entity) int {
    points := 0
    
    // TODO: Implement armor equipment checking
    // For now, basic implementation
    if entity.Type == EntityPlayer {
        points = 10 // Default player armor
    } else if entity.Type == EntityZombie {
        points = 2
    } else if entity.Type == EntitySkeleton {
        points = 0
    }
    
    return points
}

func (cm *CombatManager) applyEnchantments(attacker, target *Entity, weapon *ItemStack, damage float32) float32 {
    if weapon == nil {
        return damage
    }
    
    finalDamage := damage
    
    // TODO: Implement enchantment system
    // Sharpness, Smite, Bane of Arthropods, etc.
    
    return finalDamage
}

func (cm *CombatManager) applyKnockback(attacker, target *Entity, weapon *ItemStack) {
    if target == nil || !target.IsAlive() {
        return
    }
    
    // Calculate knockback direction
    attackerPos := attacker.GetPosition()
    targetPos := target.GetPosition()
    
    dirX := targetPos[0] - attackerPos[0]
    dirZ := targetPos[2] - attackerPos[2]
    
    // Normalize direction
    length := float32(math.Sqrt(float64(dirX*dirX + dirZ*dirZ)))
    if length > 0 {
        dirX /= length
        dirZ /= length
    }
    
    // Base knockback strength
    strength := float32(0.4)
    
    // Weapon knockback bonus
    if weapon != nil {
        // TODO: Add weapon-specific knockback
    }
    
    // Apply knockback velocity
    currentVel := target.GetVelocity()
    target.SetVelocity([3]float32{
        currentVel[0] + dirX * strength,
        currentVel[1] + 0.4, // Upward knockback
        currentVel[2] + dirZ * strength,
    })
}

func (cm *CombatManager) ProjectileAttack(projectile, target *Entity) *CombatEvent {
    if projectile == nil || target == nil || !target.IsAlive() {
        return nil
    }
    
    // Calculate projectile damage
    baseDamage := float32(2.0) // Default arrow damage
    
    // TODO: Add projectile type and enchantment modifiers
    
    // Apply armor reduction
    finalDamage := cm.applyArmorReduction(target, baseDamage)
    
    event := &CombatEvent{
        Attacker:  projectile,
        Target:    target,
        Damage:    finalDamage,
        Source:    DamageSourceProjectile,
        Timestamp: time.Now(),
        Critical:  false,
        Weapon:    nil,
    }
    
    // Apply damage
    target.Damage(finalDamage)
    
    cm.logger.Debug("Projectile: %s hit %s for %.1f damage", 
        cm.getEntityName(projectile), cm.getEntityName(target), finalDamage)
    
    return event
}

func (cm *CombatManager) EnvironmentalDamage(target *Entity, source DamageSource, damage float32) *CombatEvent {
    if target == nil || !target.IsAlive() {
        return nil
    }
    
    event := &CombatEvent{
        Attacker:  nil,
        Target:    target,
        Damage:    damage,
        Source:    source,
        Timestamp: time.Now(),
        Critical:  false,
        Weapon:    nil,
    }
    
    target.Damage(damage)
    
    cm.logger.Debug("Environmental: %s took %.1f damage from %v", 
        cm.getEntityName(target), damage, source)
    
    return event
}

func (cm *CombatManager) getEntityName(entity *Entity) string {
    if entity == nil {
        return "unknown"
    }
    
    switch entity.Type {
    case EntityPlayer:
        if player, ok := entity.Metadata["player"].(*Player); ok {
            return player.Username
        }
        return "Player"
    case EntityZombie:
        return "Zombie"
    case EntitySkeleton:
        return "Skeleton"
    case EntityCreeper:
        return "Creeper"
    case EntityItem:
        return "Item"
    default:
        return "Entity"
    }
}

func (cm *CombatManager) GetRecentEvents(limit int) []*CombatEvent {
    cm.eventMutex.RLock()
    defer cm.eventMutex.RUnlock()
    
    if limit <= 0 || limit >= len(cm.events) {
        return cm.events
    }
    
    return cm.events[len(cm.events)-limit:]
}

func (cm *CombatManager) ClearOldEvents(maxAge time.Duration) {
    cm.eventMutex.Lock()
    defer cm.eventMutex.Unlock()
    
    now := time.Now()
    validEvents := make([]*CombatEvent, 0)
    
    for _, event := range cm.events {
        if now.Sub(event.Timestamp) <= maxAge {
            validEvents = append(validEvents, event)
        }
    }
    
    cm.events = validEvents
}