package publisher

import (
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzlib-go/mq"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
)

type handler struct {
	publisher mq.Publisher
	builder   *builder
}

// Publish publishes a message
func (h *handler) Publish(ingredients event.Token) error {
	pub, err := h.builder.buildPublishing(ingredients)
	if err != nil {
		return err
	}

	routingKey := h.builder.buildRoutingKey(ingredients.Resource.Type, ingredients.EventType)
	return h.publisher.Push(*pub, routingKey)
}

// Wrapper struct definition
type Wrapper struct {
	publisher mq.Publisher
}

// Handler returns handler
func (h *Wrapper) Handler(resourceData ResourceData, auth header.Auth) event.MessageHandler {
	return &handler{
		publisher: h.publisher,
		builder: &builder{
			resourceData: resourceData,
			auth:         auth,
		},
	}
}

// New instanciates Message Handler
func New(publisher mq.Publisher) *Wrapper {
	return &Wrapper{
		publisher: publisher,
	}
}
