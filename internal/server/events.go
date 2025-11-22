package server

import (
	"garuda/internal/protocol/minecraft"
	"time"
)

// Event system untuk plugins
type EventType string

const (
	PlayerJoinEvent    EventType = "player_join"
	PlayerQuitEvent    EventType = "player_quit"
	PlayerChatEvent    EventType = "player_chat"
	PlayerMoveEvent    EventType = "player_move"
	BlockBreakEvent    EventType = "block_break"
	BlockPlaceEvent    EventType = "block_place"
)

type Event struct {
	Type      EventType
	Player    *Player
	Data      interface{}
	Timestamp time.Time
	Cancelled bool
}

type EventHandler func(event *Event)

type EventManager struct {
	handlers map[EventType][]EventHandler
}

func NewEventManager() *EventManager {
	return &EventManager{
		handlers: make(map[EventType][]EventHandler),
	}
}

func (em *EventManager) RegisterHandler(eventType EventType, handler EventHandler) {
	em.handlers[eventType] = append(em.handlers[eventType], handler)
}

func (em *EventManager) CallEvent(event *Event) {
	if handlers, exists := em.handlers[event.Type]; exists {
		for _, handler := range handlers {
			handler(event)
			if event.Cancelled {
				break
			}
		}
	}
}