package mq

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// Publisher interface
type Publisher interface {
	Push(publishing amqp.Publishing, routingKey string) error
	Close() error
}

type publisher struct {
	session
}

const (
	resendDelay       time.Duration = 500 * time.Millisecond
	maximumRetryCount int           = 3
)

// CreatePublisher declares pre-defined routing, and automatically
// attempts to connect to the server.
func CreatePublisher(addr string, l logger) Publisher {
	p := publisher{
		session: session{
			done:   make(chan bool),
			logger: newMQLoggerWrapper(l),
		},
	}
	p.declareFunc = p.declarePublisher

	go p.handleReconnect(addr)
	return &p
}

func (p *publisher) declarePublisher(ch *amqp.Channel) error {
	err := p.declareBaseExchanges(ch)
	if err != nil {
		return err
	}
	return nil
}

// Push will push data onto the queue, and wait for a confirm.
// If no confirms are received until within the resendTimeout,
// it continuously re-sends messages until a confirm is received.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (p *publisher) Push(publishing amqp.Publishing, routingKey string) error {
	var err error
	for i := 0; i < maximumRetryCount; i++ {
		err = p.push(publishing, routingKey)
		if err == nil {
			return nil
		}

		p.logger.Warnf("%d'th try, Push didn't confirm. Retrying... err: %v", i+1, err)
	}

	return fmt.Errorf("Failed to push. last error: %v", err)
}

func (p *publisher) push(publishing amqp.Publishing, routingKey string) error {
	ch, err := p.newChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Publish(
		rootExchange, // Exchange
		routingKey,   // Routing key
		false,        // Mandatory
		false,        // Immediate
		publishing,
	)
}

// Close will cleanly shutdown the channel and connection.
func (p *publisher) Close() error {
	return p.close()
}
