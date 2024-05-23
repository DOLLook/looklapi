package appcontext

import (
	"fmt"
	"log"
)

// an AppObserver is one of subscriber that subscribe the application runtime event
type AppObserver interface {

	// received app event and process.
	// for event publish well, the developers must deal with the panic by their self
	OnApplicationEvent(event interface{})

	// register to the application event publisher
	Subscribe()
}

func onEvent(observer AppObserver, event interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if throw, ok := err.(error); ok {
				log.Println(fmt.Sprintf("%v %s", "ERROR", throw.Error()))
			} else if msg, ok := err.(string); ok {
				log.Println(fmt.Sprintf("%v %s", "ERROR", msg))
			}
		}
	}()

	observer.OnApplicationEvent(event)
}
