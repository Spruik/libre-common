package ports

type LibreDataStoreTransactionPort interface {
	ExecuteQuery(q interface{}, vars map[string]interface{}) error
	ExecuteMutation(m interface{}, vars map[string]interface{}) error
	Commit()
	Dispose()
}

type LibreDataStoreSubscriptionPort interface {
	SetSubscriptionQuery(q interface{}, vars map[string]interface{})
	GetSubscriptionNotifications(notificationChannel chan []byte)
	StopGettingSubscriptionNotifications()
}

//The LibreDataStorePort interface defines the functions to be provided by the service acting as the data store in Libre
type LibreDataStorePort interface {

	//Connect is called to establish a connection to the data store service
	Connect() error

	//Close is called to close the data store connection
	Close() error

	//BeginTransaction starts a transaction with the data store and returns a handle for use with operations
	BeginTransaction(forUpdate bool, name string) LibreDataStoreTransactionPort

	//GetSubscription returns a handle to a database subscription
	GetSubscription(q interface{}, vars map[string]interface{}) LibreDataStoreSubscriptionPort
}
