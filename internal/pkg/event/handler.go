package event

// MessageHandler defines an interface providing publish event
type MessageHandler interface {
	Publish(ingredients Token) error
}
