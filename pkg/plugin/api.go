package plugin

import (
    "garuda/internal/protocol/minecraft"
    "garuda/internal/world"
)

type PluginAPI struct {
    manager PluginManager
}

func NewPluginAPI(manager PluginManager) *PluginAPI {
    return &PluginAPI{
        manager: manager,
    }
}

func (api *PluginAPI) GetServer() Server {
    return api.manager.GetServer()
}

func (api *PluginAPI) BroadcastMessage(message string) {
    api.manager.BroadcastMessage(message)
}

func (api *PluginAPI) GetPlayer(name string) world.Player {
    return api.manager.GetPlayer(name)
}

func (api *PluginAPI) GetOnlinePlayers() []world.Player {
    return api.manager.GetOnlinePlayers()
}

func (api *PluginAPI) RegisterEvent(eventType EventType, handler EventHandler) {
    api.manager.RegisterEvent(eventType, handler)
}

func (api *PluginAPI) RegisterCommand(command string, handler CommandHandler) {
    api.manager.RegisterCommand(command, handler)
}

func (api *PluginAPI) CreateItemStack(id uint32, count byte, data uint16) minecraft.ItemStack {
    return minecraft.ItemStack{
        ID:    id,
        Count: count,
        Data:  data,
    }
}

func (api *PluginAPI) CreateBlockPosition(x, y, z int) [3]int {
    return [3]int{x, y, z}
}

func (api *PluginAPI) CreateVector3(x, y, z float32) [3]float32 {
    return [3]float32{x, y, z}
}

type PermissionManager interface {
    AddPermission(player world.Player, permission string)
    RemovePermission(player world.Player, permission string)
    HasPermission(player world.Player, permission string) bool
}

type SimplePermissionManager struct {
    permissions map[string]map[string]bool
}

func NewSimplePermissionManager() *SimplePermissionManager {
    return &SimplePermissionManager{
        permissions: make(map[string]map[string]bool),
    }
}

func (pm *SimplePermissionManager) AddPermission(player world.Player, permission string) {
    playerName := player.Username
    if pm.permissions[playerName] == nil {
        pm.permissions[playerName] = make(map[string]bool)
    }
    pm.permissions[playerName][permission] = true
}

func (pm *SimplePermissionManager) RemovePermission(player world.Player, permission string) {
    playerName := player.Username
    if pm.permissions[playerName] != nil {
        delete(pm.permissions[playerName], permission)
    }
}

func (pm *SimplePermissionManager) HasPermission(player world.Player, permission string) bool {
    playerName := player.Username
    if pm.permissions[playerName] == nil {
        return false
    }
    return pm.permissions[playerName][permission]
}

type Scheduler interface {
    RunTask(plugin Plugin, task func())
    RunTaskLater(plugin Plugin, task func(), delayTicks int)
    RunTaskTimer(plugin Plugin, task func(), delayTicks int, periodTicks int)
    CancelTasks(plugin Plugin)
}

type SimpleScheduler struct {
    tasks map[string][]*ScheduledTask
}

type ScheduledTask struct {
    TaskID      int
    Plugin      Plugin
    Task        func()
    DelayTicks  int
    PeriodTicks int
    CurrentTick int
    Cancelled   bool
}

func NewSimpleScheduler() *SimpleScheduler {
    return &SimpleScheduler{
        tasks: make(map[string][]*ScheduledTask),
    }
}

func (s *SimpleScheduler) RunTask(plugin Plugin, task func()) {
    s.runTaskInternal(plugin, task, 0, 0)
}

func (s *SimpleScheduler) RunTaskLater(plugin Plugin, task func(), delayTicks int) {
    s.runTaskInternal(plugin, task, delayTicks, 0)
}

func (s *SimpleScheduler) RunTaskTimer(plugin Plugin, task func(), delayTicks int, periodTicks int) {
    s.runTaskInternal(plugin, task, delayTicks, periodTicks)
}

func (s *SimpleScheduler) runTaskInternal(plugin Plugin, task func(), delayTicks int, periodTicks int) {
    pluginName := plugin.GetName()
    taskID := len(s.tasks[pluginName]) + 1
    
    scheduledTask := &ScheduledTask{
        TaskID:      taskID,
        Plugin:      plugin,
        Task:        task,
        DelayTicks:  delayTicks,
        PeriodTicks: periodTicks,
        CurrentTick: 0,
        Cancelled:   false,
    }
    
    s.tasks[pluginName] = append(s.tasks[pluginName], scheduledTask)
}

func (s *SimpleScheduler) CancelTasks(plugin Plugin) {
    pluginName := plugin.GetName()
    delete(s.tasks, pluginName)
}

func (s *SimpleScheduler) Tick() {
    for pluginName, tasks := range s.tasks {
        var remainingTasks []*ScheduledTask
        
        for _, task := range tasks {
            if task.Cancelled {
                continue
            }
            
            task.CurrentTick++
            
            if task.CurrentTick >= task.DelayTicks {
                if task.PeriodTicks > 0 {
                    if (task.CurrentTick-task.DelayTicks)%task.PeriodTicks == 0 {
                        task.Task()
                    }
                } else {
                    task.Task()
                    task.Cancelled = true
                }
            }
            
            if !task.Cancelled {
                remainingTasks = append(remainingTasks, task)
            }
        }
        
        s.tasks[pluginName] = remainingTasks
    }
}