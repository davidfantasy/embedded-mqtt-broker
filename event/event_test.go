package event

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventBusBase(t *testing.T) {
	const test_event = -1
	val := 0
	handler1 := func(event Event) {
		val += 50
	}
	handler2 := func(event Event) {
		val += 100
	}
	DefaultEventBus.Subscribe(test_event, handler1)
	DefaultEventBus.Subscribe(test_event, handler2)
	DefaultEventBus.Subscribe(test_event, handler1)
	event := NewEvent(test_event, 50)
	DefaultEventBus.Publish(event)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 150, val)
}

func TestEventBus2(t *testing.T) {
	const testEvent1 = 1
	const testEvent2 = 2
	var wg sync.WaitGroup
	wg.Add(2)
	handler1 := func(event Event) {
		wg.Done()
	}
	handler2 := func(event Event) {
		wg.Done()
	}
	DefaultEventBus.Subscribe(testEvent1, handler1)
	DefaultEventBus.Subscribe(testEvent2, handler2)
	DefaultEventBus.Publish(NewEvent(testEvent1, ""))
	DefaultEventBus.Publish(NewEvent(testEvent2, ""))
	wg.Wait()
	fmt.Println("test is passed!")
}
