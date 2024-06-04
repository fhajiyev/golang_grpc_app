package publisher

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type builder struct {
	resourceData ResourceData
	auth         header.Auth
}

func (b *builder) buildRoutingKey(resourceType string, event string) string {
	return fmt.Sprintf("%s.%s", resourceType, event)
}

func (b *builder) buildPublishing(ingredients event.Token) (*amqp.Publishing, error) {
	h := amqp.Table{}
	h = header.AppendAuthToAMQPTable(h, b.auth)

	em := &eventMessage{
		baseMessage: baseMessage{
			ResourceID:   ingredients.Resource.ID,
			ResourceType: ingredients.Resource.Type,
			Event:        ingredients.EventType,
		},
		Extra: eventExtra{
			Reward: reward{
				TransactionID: ingredients.TransactionID,
			},
			ResourceData: b.resourceData,
			UnitData: UnitData{
				ID: ingredients.UnitID,
			},
		},
	}

	bytes, err := json.Marshal(em)
	if err != nil {
		return nil, err
	}

	pub := &amqp.Publishing{
		Headers:      h,
		ContentType:  "application/json",
		Body:         bytes,
		Timestamp:    time.Now(),
		DeliveryMode: 2,
		MessageId:    uuid.NewV4().String(),
	}

	return pub, nil
}
