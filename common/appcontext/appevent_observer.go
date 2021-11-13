package appcontext

// an AppObserver is one of subscriber that subscribe the application runtime event
type AppObserver interface {

	// recieved app event and process.
	// for event publish well, the developers must deal with the panic by their self
	OnApplicationEvent(event interface{})

	// regiser to the application event publisher
	Subscribe()
}
