package event

import (
	"reflect"
	"sync"
	"time"

	"github.com/davidfantasy/embedded-mqtt-broker/logger"
)

var DefaultEventBus = AsyncEventBus{
	handlerMap: make(map[EventType][]eventHandlerWapper),
}

type Event struct {
	EventType EventType
	Ts        int64
	Data      interface{}
}

type EventType int
type EventHandler func(event Event)

type EventBus interface {
	Publish(event Event)
	Subscribe(eventType EventType, handler EventHandler)
}

func NewEvent(eventType EventType, data any) Event {
	return Event{EventType: eventType, Data: data, Ts: time.Now().UnixMilli()}
}

type eventHandlerWapper struct {
	handler EventHandler
	ch      chan Event
}

func wrapHandler(handler EventHandler) eventHandlerWapper {
	wrapped := eventHandlerWapper{handler: handler, ch: make(chan Event, 100)}
	go func() {
		for event := range wrapped.ch {
			callHandler(event, wrapped.handler)
		}
	}()
	return wrapped
}

func callHandler(event Event, handler EventHandler) {
	defer func() {
		if r := recover(); r != nil {
			logger.ERROR.Printf("事件处理时发生异常：%v,%v\n", event, r)
		}
	}()
	handler(event)
}

type AsyncEventBus struct {
	handlerMap map[EventType][]eventHandlerWapper
}

var handlerMu sync.Mutex

func (bus *AsyncEventBus) Publish(event Event) {
	handlerMu.Lock()
	handlers := bus.handlerMap[event.EventType]
	handlerMu.Unlock()
	for _, handler := range handlers {
		select {
		case handler.ch <- event:
		default:
			logger.WARN.Printf("事件处理器被阻塞了：%v\n", event.EventType)
		}
	}
}

func (bus *AsyncEventBus) Subscribe(eventType EventType, handler EventHandler) {
	if handler == nil {
		logger.WARN.Println("事件处理函数为nil")
		return
	}
	handlerMu.Lock()
	defer handlerMu.Unlock()
	wrappers := bus.handlerMap[eventType]
	if wrappers == nil {
		wrappers = make([]eventHandlerWapper, 0)
	}
	hp := reflect.ValueOf(handler).Pointer()
	existed := false
	for _, w := range wrappers {
		if reflect.ValueOf(w.handler).Pointer() == hp {
			existed = true
			break
		}
	}
	if !existed {
		wrapped := wrapHandler(handler)
		wrappers = append(wrappers, wrapped)
		bus.handlerMap[eventType] = wrappers
	}
}
