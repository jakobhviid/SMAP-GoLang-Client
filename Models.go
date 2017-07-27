package SMAPClient

//SubscribtionMessage is a container for the JSON data comming from the archiver
type SubscribtionMessage struct {
	Path     string
	UUID     string
	Readings []SubscribtionReadingContainer
}

//SubscribtionReadingContainer contains readings
type SubscribtionReadingContainer struct {
	UnixTime int64
	Value    string
}
