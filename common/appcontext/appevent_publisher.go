package appcontext

import (
	"reflect"
	"sync"
)

// the application event publisher.
// it is one instance on the application runtime
type appEventPublisher struct {
	observers map[reflect.Type][]AppObserver
	mu        *sync.Mutex
}

var eventPublisher *appEventPublisher
var once = &sync.Once{}

// get the singleton application event publisher
func GetAppEventPublisher() *appEventPublisher {
	if eventPublisher == nil {
		once.Do(func() {
			eventPublisher = &appEventPublisher{mu: &sync.Mutex{}, observers: make(map[reflect.Type][]AppObserver)}
		})
	}

	return eventPublisher
}

// register observer to the application event publisher
// @eventType the event type which the observer intrested in
func (publisher *appEventPublisher) Subscribe(observer AppObserver, eventType reflect.Type) {
	if observer == nil || eventType == nil {
		return
	}
	publisher.mu.Lock()
	defer publisher.mu.Unlock()

	needAdd := true
	if obs, ok := publisher.observers[eventType]; ok {
		for _, ob := range obs {
			if ob == observer {
				needAdd = false
				break
			}
		}
	}
	if needAdd {
		publisher.observers[eventType] = append(publisher.observers[eventType], observer)
	}
}

// publish any event to the observers that
// have registered to the application event publisher
func (publisher *appEventPublisher) PublishEvent(event interface{}) {
	if event == nil {
		return
	}
	etype := reflect.TypeOf(event)
	if observers, ok := publisher.observers[etype]; ok {
		for _, observer := range observers {
			observer.OnApplicationEvent(event)
		}
	}
}
