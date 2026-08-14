package publisher

// Publisher publishes raw event payloads.
type Publisher interface {
	Publish(payload []byte) error
}