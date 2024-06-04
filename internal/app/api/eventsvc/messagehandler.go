package eventsvc

import (
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/eventsvc/publisher"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
)

// Publisher interface definition
type Publisher interface {
	Handler(resourceData publisher.ResourceData, auth header.Auth) *event.MessageHandler
}
