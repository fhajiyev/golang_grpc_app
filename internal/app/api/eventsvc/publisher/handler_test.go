package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"testing"

	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/bxcodec/faker"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

func (ts *RepoTestSuite) TestPublishEvent() {
	a := ts.createAuth()
	ingredients := ts.createToken()
	resourceData := ts.createResourceData(ingredients.Resource)
	message := ts.buildEventMessage(ingredients, resourceData)
	routingKey := ts.buildRoutingKey(message)

	body, err := json.Marshal(message)
	ts.NoError(err)

	equalHeader := func(header amqp.Table) bool {
		if header["Buzz-App-Id"] != a.AppID {
			return false
		} else if header["Buzz-Account-Id"] != a.AccountID {
			return false
		} else if header["Buzz-Publisher-User-Id"] != a.PublisherUserID {
			return false
		} else if header["Buzz-Ifa"] != a.IFA {
			return false
		}
		return true
	}
	equalBody := func(expectedBody []byte) bool {
		return string(body) == string(expectedBody)
	}
	equalPublishing := func(pub amqp.Publishing) bool {
		if !equalHeader(pub.Headers) {
			log.Printf("not equal header")
			return false
		} else if !equalBody(pub.Body) {
			log.Printf("not equal body %s\n%s", string(body), string(pub.Body))
			return false
		}
		return true
	}

	ts.publisher.On("Push", mock.MatchedBy(equalPublishing), routingKey).Return(nil).Once()

	handler := ts.messageHandler.Handler(resourceData, a)
	err = handler.Publish(ingredients)
	ts.NoError(err)
	ts.publisher.AssertExpectations(ts.T())
}

func (ts *RepoTestSuite) buildRoutingKey(message eventMessage) string {
	return fmt.Sprintf("%s.%s", message.ResourceType, message.Event)
}

func (ts *RepoTestSuite) createAuth() header.Auth {
	a := header.Auth{}
	ts.NoError(faker.FakeData(&a))
	return a
}

func (ts *RepoTestSuite) createToken() event.Token {
	return event.Token{
		TransactionID: uuid.NewV4().String(),
		Resource:      ts.createResource(),
		EventType:     "landed",
		UnitID:        rand.Int63n(100000) + 1,
	}
}

func (ts *RepoTestSuite) createResourceData(resource event.Resource) ResourceData {
	return ResourceData{
		ID:             resource.ID,
		Name:           "TEST_AD_NAME",
		OrganizationID: rand.Int63n(1000000) + 1,
		RevenueType:    "cpc",
		Extra: map[string]interface{}{
			"unit": map[string]interface{}{
				"id": rand.Int63n(100000) + 1,
			},
		},
	}
}

func (ts *RepoTestSuite) createResource() event.Resource {
	return event.Resource{
		ID:   rand.Int63n(100000) + 1,
		Type: "ad",
	}
}

func (ts *RepoTestSuite) buildEventMessage(ingredients event.Token, resoureData ResourceData) eventMessage {
	message := eventMessage{
		baseMessage: baseMessage{
			ResourceID:   ingredients.Resource.ID,
			ResourceType: ingredients.Resource.Type,
			Event:        ingredients.EventType,
		},
		Extra: eventExtra{
			Reward: reward{
				TransactionID: ingredients.TransactionID,
			},
			ResourceData: resoureData,
			UnitData: UnitData{
				ID: ingredients.UnitID,
			},
		},
	}
	return message
}

type RepoTestSuite struct {
	suite.Suite
	publisher      *mockMQPublisher
	messageHandler *Wrapper
}

func (ts *RepoTestSuite) SetupTest() {
	ts.publisher = new(mockMQPublisher)
	ts.messageHandler = &Wrapper{ts.publisher}
}

type mockMQPublisher struct {
	mock.Mock
}

func (p *mockMQPublisher) Push(publishing amqp.Publishing, routingKey string) error {
	ret := p.Called(publishing, routingKey)
	return ret.Error(0)
}

func (p *mockMQPublisher) Close() error {
	ret := p.Called()
	return ret.Error(0)
}
