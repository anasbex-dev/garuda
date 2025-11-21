package world

import (
    "math"
    "time"
)

type DamageSource int

const (
    DamageSourceEntity DamageSource = iota
    DamageSourceProjectile
    DamageSourceFall
    DamageSourceFire
    DamageSourceLava
    DamageSourceDrowning
    DamageSourceBlock
    DamageSourceMagic
    DamageSourceStarvation
)

type CombatManager struct {
    world *World
}

type DamageInfo struct {
    Amount       float32
    Source       DamageSource
    Attacker     *Entity
    Victim       *Entity
    Timestamp    time.Time
    Critical     bool
    Knockback    float32
}

func NewCombatManager(world *World) *CombatManager {
    return &CombatManager{
        world: world,
    }
}

func (cm *CombatManager) Attack(attacker, victim *Entity, item *ItemStack) *DamageInfo {
    if attacker.Dead || victim.Dead {
        return nil
    }

    // Calculate base damage
    baseDamage := cm.calculateBaseDamage(attacker, victim, item)
    
    // Check for critical hit
    critical := cm.isCriticalHit(attacker)
    if critical {
        baseDamage *= 1.5
    }

    // Apply armor reduction
    finalDamage := cm.applyArmorReduction(victim, baseDamage)
    
    // Apply enchantments
    finalDamage = cm.applyEnchantments(attacker, victim, finalDamage, item)

    damageInfo := &DamageInfo{
        Amount:    finalDamage,
        Source:    DamageSourceEntity,
        Attacker:  attacker,
        Victim:    victim,
        Timestamp: time.Now(),
        Critical:  critical,
        Knockback: 0.4,
    }

    // Apply the damage
    cm.applyDamage(damageInfo)

    // Apply knockback
    cm.applyKnockback(damageInfo)

    // Play effects
    cm.playAttackEffects(damageInfo)

    return damageInfo
}

func (cm *CombatManager) calculateBaseDamage(attacker, victim *Entity, item *ItemStack) float32 {
    baseDamage := float32(1.0) // Base hand damage

    if item != nil && item.ID != 0 {
        // Weapon damage
        switch item.ID {
        case 268: // Wooden sword
            baseDamage = 4.0
        case 267: // Iron sword
            baseDamage = 6.0
        case 272: // Stone sword
            baseDamage = 5.0
        case 276: // Diamond sword
            baseDamage = 7.0
        case 283: // Gold sword
            baseDamage = 4.0
        // Add more weapons...
        default:
            // Tool damage
            baseDamage = 2.0
        }
    }

    // Mob-specific base damage
    if mobData, ok := attacker.Data.(*MobData); ok {
        baseDamage = mobData.AttackDamage
    }

    return baseDamage
}

func (cm *CombatManager) isCriticalHit(attacker *Entity) bool {
    // Critical hit if attacker is falling and not on ground
    if !attacker.OnGround && attacker.Motion.Y < 0 {
        return true
    }
    return false
}

func (cm *CombatManager) applyArmorReduction(victim *Entity, damage float32) float32 {
    // Simple armor reduction (will be improved with actual armor system)
    armorPoints := cm.getArmorPoints(victim)
    reduction := float32(armorPoints) * 0.04
    if reduction > 0.8 {
        reduction = 0.8
    }
    
    return damage * (1 - reduction)
}

func (cm *CombatManager) getArmorPoints(entity *Entity) int {
    if entity.Type == EntityPlayer {
        // TODO: Implement actual armor calculation
        return 0
    }
    
    // Mob armor points
    switch entity.Type {
    case EntityZombie:
        return 2
    case EntitySkeleton:
        return 0
    default:
        return 0
    }
}

func (cm *CombatManager) applyEnchantments(attacker, victim *Entity, damage float32, item *ItemStack) float32 {
    if item == nil {
        return damage
    }

    // TODO: Implement enchantment system
    // For now, just return base damage
    return damage
}

func (cm *CombatManager) applyDamage(damageInfo *DamageInfo) {
    victim := damageInfo.Victim
    
    // Apply damage to entity
    cm.world.entityManager.DamageEntity(victim, damageInfo.Amount, damageInfo.Attacker)
    
    // Play hurt animation and sound
    cm.playHurtEffects(damageInfo)
    
    // Handle player-specific damage
    if victim.Type == EntityPlayer {
        cm.handlePlayerDamage(damageInfo)
    }
}

func (cm *CombatManager) applyKnockback(damageInfo *DamageInfo) {
    victim := damageInfo.Victim
    attacker := damageInfo.Attacker
    
    if attacker == nil {
        return
    }

    // Calculate direction from attacker to victim
    dx := victim.Position.X - attacker.Position.X
    dz := victim.Position.Z - attacker.Position.Z
    
    // Normalize direction
    distance := float32(math.Sqrt(float64(dx*dx + dz*dz)))
    if distance > 0 {
        dx /= distance
        dz /= distance
    }
    
    // Apply knockback
    knockbackStrength := damageInfo.Knockback
    if damageInfo.Critical {
        knockbackStrength *= 1.5
    }
    
    victim.Motion.X = dx * knockbackStrength
    victim.Motion.Z = dz * knockbackStrength
    victim.Motion.Y = 0.4 // Small upward motion
}

func (cm *CombatManager) handlePlayerDamage(damageInfo *DamageInfo) {
    player := cm.world.GetPlayer(damageInfo.Victim.EntityID)
    if player == nil {
        return
    }
    
    // Update player health
    player.Health -= damageInfo.Amount
    
    // Check for death
    if player.Health <= 0 {
        cm.handlePlayerDeath(player, damageInfo.Attacker)
    }
    
    // Send health update to client
    cm.sendHealthUpdate(player)
}

func (cm *CombatManager) handlePlayerDeath(player *Player, killer *Entity) {
    log.Printf("Player %s died", player.Username)
    
    // Drop player inventory
    cm.dropPlayerInventory(player)
    
    // Send death message
    deathMessage := cm.getDeathMessage(player, killer)
    cm.broadcastMessage(deathMessage)
    
    // Respawn player
    cm.respawnPlayer(player)
}

func (cm *CombatManager) getDeathMessage(player *Player, killer *Entity) string {
    if killer == nil {
        return player.Username + " died"
    }
    
    switch killer.Type {
    case EntityPlayer:
        killerPlayer := cm.world.GetPlayer(killer.EntityID)
        if killerPlayer != nil {
            return player.Username + " was slain by " + killerPlayer.Username
        }
    case EntityZombie:
        return player.Username + " was slain by Zombie"
    case EntitySkeleton:
        return player.Username + " was shot by Skeleton"
    case EntityCreeper:
        return player.Username + " was blown up by Creeper"
    case EntitySpider:
        return player.Username + " was slain by Spider"
    default:
        return player.Username + " died"
    }
    
    return player.Username + " died"
}

func (cm *CombatManager) dropPlayerInventory(player *Player) {
    // Drop all items in inventory
    for i := 0; i < player.Inventory.size; i++ {
        item := player.Inventory.GetItem(i)
        if item != nil && item.ID != 0 {
            // Create item entity
            cm.world.entityManager.CreateItemEntity(player.Position, item)
            // Clear the slot
            player.Inventory.SetItem(i, &ItemStack{ID: 0, Count: 0})
        }
    }
}

func (cm *CombatManager) respawnPlayer(player *Player) {
    // Reset health
    player.Health = player.MaxHealth
    
    // Teleport to spawn
    player.Position = minecraft.Vector3{X: 0, Y: 70, Z: 0}
    player.Rotation = minecraft.Vector2{X: 0, Y: 0}
    
    // Clear motion
    // Send respawn packet to client
    cm.sendRespawnPacket(player)
}

func (cm *CombatManager) sendHealthUpdate(player *Player) {
    // TODO: Implement health update packet
}

func (cm *CombatManager) sendRespawnPacket(player *Player) {
    // TODO: Implement respawn packet
}

func (cm *CombatManager) broadcastMessage(message string) {
    // TODO: Implement broadcast to all players
}

func (cm *CombatManager) playAttackEffects(damageInfo *DamageInfo) {
    // Play attack sound based on weapon and critical
    // TODO: Implement sound system
}

func (cm *CombatManager) playHurtEffects(damageInfo *DamageInfo) {
    // Play hurt sound and animation
    // TODO: Implement effects system
}

// Projectile combat
type Projectile struct {
    Shooter     *Entity
    Position    minecraft.Vector3
    Motion      minecraft.Vector3
    Damage      float32
    Type        ProjectileType
    Lifetime    int
}

type ProjectileType int

const (
    ProjectileArrow ProjectileType = iota
    ProjectileSnowball
    ProjectileEgg
    ProjectileEnderPearl
)

func (cm *CombatManager) ShootProjectile(shooter *Entity, projectileType ProjectileType, power float32) *Projectile {
    // Calculate initial position (in front of shooter)
    yaw := shooter.Rotation.Y * math.Pi / 180
    pitch := shooter.Rotation.X * math.Pi / 180
    
    pos := minecraft.Vector3{
        X: shooter.Position.X - float32(math.Sin(float64(yaw)))*0.5,
        Y: shooter.Position.Y + 1.5, // Eye level
        Z: shooter.Position.Z + float32(math.Cos(float64(yaw)))*0.5,
    }
    
    // Calculate motion based on look direction
    motion := minecraft.Vector3{
        X: -float32(math.Sin(float64(yaw)) * math.Cos(float64(pitch))) * power,
        Y: -float32(math.Sin(float64(pitch))) * power,
        Z: float32(math.Cos(float64(yaw)) * math.Cos(float64(pitch))) * power,
    }
    
    projectile := &Projectile{
        Shooter:  shooter,
        Position: pos,
        Motion:   motion,
        Damage:   2.0, // Base arrow damage
        Type:     projectileType,
        Lifetime: 1200, // 60 seconds at 20 ticks/sec
    }
    
    // TODO: Add projectile to world and send to clients
    return projectile
}

func (cm *CombatManager) UpdateProjectile(projectile *Projectile) bool {
    projectile.Lifetime--
    if projectile.Lifetime <= 0 {
        return true // Remove projectile
    }
    
    // Apply gravity
    projectile.Motion.Y -= 0.05
    
    // Update position
    projectile.Position.X += projectile.Motion.X
    projectile.Position.Y += projectile.Motion.Y
    projectile.Position.Z += projectile.Motion.Z
    
    // Check for collisions
    if cm.checkProjectileCollision(projectile) {
        cm.handleProjectileImpact(projectile)
        return true // Remove projectile
    }
    
    return false
}

func (cm *CombatManager) checkProjectileCollision(projectile *Projectile) bool {
    // Check block collision
    blockPos := minecraft.BlockPos{
        X: int32(math.Floor(float64(projectile.Position.X))),
        Y: int32(math.Floor(float64(projectile.Position.Y))),
        Z: int32(math.Floor(float64(projectile.Position.Z))),
    }
    
    block := cm.world.GetBlock(blockPos)
    if block.ID != 0 && !block.IsTransparent() {
        return true
    }
    
    // Check entity collision
    entities := cm.world.entityManager.GetEntitiesInRange(projectile.Position, 1.0)
    for _, entity := range entities {
        if entity.EntityID == projectile.Shooter.EntityID {
            continue // Don't hit shooter
        }
        
        entityAABB := cm.world.entityManager.getEntityAABB(entity)
        projectileAABB := AABB{
            MinX: float64(projectile.Position.X) - 0.1,
            MinY: float64(projectile.Position.Y) - 0.1,
            MinZ: float64(projectile.Position.Z) - 0.1,
            MaxX: float64(projectile.Position.X) + 0.1,
            MaxY: float64(projectile.Position.Y) + 0.1,
            MaxZ: float64(projectile.Position.Z) + 0.1,
        }
        
        if entityAABB.CollidesWith(&projectileAABB) {
            cm.handleProjectileHitEntity(projectile, entity)
            return true
        }
    }
    
    return false
}

func (cm *CombatManager) handleProjectileImpact(projectile *Projectile) {
    // Play impact effect based on projectile type
    switch projectile.Type {
    case ProjectileArrow:
        // TODO: Play arrow stick sound
    case ProjectileSnowball:
        // TODO: Play snowball break sound
    case ProjectileEgg:
        // TODO: Play egg break sound and chance to spawn chicken
    case ProjectileEnderPearl:
        // TODO: Teleport shooter to impact location
    }
}

func (cm *CombatManager) handleProjectileHitEntity(projectile *Projectile, entity *Entity) {
    damageInfo := &DamageInfo{
        Amount:    projectile.Damage,
        Source:    DamageSourceProjectile,
        Attacker:  projectile.Shooter,
        Victim:    entity,
        Timestamp: time.Now(),
        Critical:  false,
        Knockback: 0.2,
    }
    
    cm.applyDamage(damageInfo)
}

// Environmental damage
func (cm *CombatManager) CheckEnvironmentalDamage(entity *Entity) {
    pos := entity.Position
    
    // Check for fire damage
    blockPos := minecraft.BlockPos{
        X: int32(math.Floor(float64(pos.X))),
        Y: int32(math.Floor(float64(pos.Y))),
        Z: int32(math.Floor(float64(pos.Z))),
    }
    
    block := cm.world.GetBlock(blockPos)
    
    if block.ID == 10 || block.ID == 11 { // Lava
        damageInfo := &DamageInfo{
            Amount:    4.0,
            Source:    DamageSourceLava,
            Attacker:  nil,
            Victim:    entity,
            Timestamp: time.Now(),
        }
        cm.applyDamage(damageInfo)
    }
    
    // Check for fall damage
    if entity.Motion.Y < -0.5 && entity.OnGround {
        fallDamage := math.Abs(float64(entity.Motion.Y)) * 2.0
        if fallDamage > 3.0 {
            damageInfo := &DamageInfo{
                Amount:    float32(fallDamage),
                Source:    DamageSourceFall,
                Attacker:  nil,
                Victim:    entity,
                Timestamp: time.Now(),
            }
            cm.applyDamage(damageInfo)
        }
    }
    
    // Check for drowning
    if cm.isEntityInWater(entity) && entity.Age%20 == 0 { // Every second
        damageInfo := &DamageInfo{
            Amount:    2.0,
            Source:    DamageSourceDrowning,
            Attacker:  nil,
            Victim:    entity,
            Timestamp: time.Now(),
        }
        cm.applyDamage(damageInfo)
    }
}

func (cm *CombatManager) isEntityInWater(entity *Entity) bool {
    pos := entity.Position
    blockPos := minecraft.BlockPos{
        X: int32(math.Floor(float64(pos.X))),
        Y: int32(math.Floor(float64(pos.Y))),
        Z: int32(math.Floor(float64(pos.Z))),
    }
    
    block := cm.world.GetBlock(blockPos)
    return block.ID == 8 || block.ID == 9 // Water
}