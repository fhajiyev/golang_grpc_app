package mq

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

type session struct {
	logger          logger
	connection      *amqp.Connection
	done            chan bool
	notifyConnClose chan *amqp.Error
	isReady         bool
	declareFunc     func(ch *amqp.Channel) error
}

const (
	// When setting up the channel after a channel exception
	reInitDelay    time.Duration = time.Second
	reconnectDelay time.Duration = time.Second

	rootExchange string = "buzzvil.root"

	alternateExchange string = "buzzvil.alternate"
	unroutedQueue     string = "buzzvil.unrouted"
)

func (s *session) declareRootExchange(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(rootExchange, "topic", true, false, false, false, amqp.Table{
		"alternate-exchange": alternateExchange,
	})
	if err != nil {
		return fmt.Errorf("declare \"%s\" exchange. err: %v", rootExchange, err)
	}

	err = ch.ExchangeDeclare(alternateExchange, "fanout", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare \"%s\" exchange. err: %v", alternateExchange, err)
	}

	_, err = ch.QueueDeclare(unroutedQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare \"%s\" queue. err: %v", unroutedQueue, err)
	}

	err = ch.QueueBind(unroutedQueue, "", alternateExchange, false, nil)
	if err != nil {
		return fmt.Errorf("bind \"%s\" queue to \"%s\" exchange. err: %v", unroutedQueue, alternateExchange, err)
	}

	return nil
}

func (s *session) declareBaseExchanges(ch *amqp.Channel) error {
	return s.declareRootExchange(ch)
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (s *session) handleReconnect(addr string) {
	for {
		// non-blocking channel
		select {
		case <-s.done:
			s.logger.Infof("connection handler closed")
			return
		default:
			s.isReady = false
			s.logger.Infof("Attempt to connect")

			err := s.connect(addr)
			if err != nil {
				s.logger.Warnf("Failed to connect. Retrying... err %v", err)
				select {
				case <-s.done:
					return
				case <-time.After(reconnectDelay):
				}
				continue
			}

			s.handleReInit()
		}
	}
}

// connect will create a new AMQP connection
func (s *session) connect(addr string) error {
	var err error
	s.connection, err = amqp.Dial(addr)
	if err != nil {
		return err
	}

	s.notifyConnClose = make(chan *amqp.Error)
	s.connection.NotifyClose(s.notifyConnClose)

	s.logger.Infof("Connected!")
	return nil
}

// handleReconnect will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (s *session) handleReInit() {
	for {
		s.isReady = false

		err := s.declareRouting()
		if err != nil {
			s.logger.Warnf("Failed to initialize channel. Retrying... err: %v", err)

			select {
			case <-s.done:
				return
			case <-time.After(reInitDelay):
			}
			continue
		}

		select {
		case <-s.done:
			return
		case <-s.notifyConnClose:
			s.logger.Warnf("Connection closed. Reconnecting...")
			return
		}
	}
}

// init will initialize channel & declare queue
func (s *session) declareRouting() error {
	ch, err := s.connection.Channel()
	if err != nil {
		return err
	}

	err = s.declareFunc(ch)
	if err != nil {
		return err
	}

	err = ch.Close()
	if err != nil {
		return err
	}

	s.isReady = true
	return nil
}

func (s *session) newChannel() (*amqp.Channel, error) {
	if !s.isReady {
		select {
		case <-time.After(2 * reconnectDelay):
			if !s.isReady {
				// Fail if connection is not ready after 2 * reconnectDelay
				return nil, ClosedConnectionError{}
			}
		}
	}

	return s.connection.Channel()
}

// Close will cleanly shutdown the channel and connection.
func (s *session) close() error {
	close(s.done)
	if !s.isReady {
		return ClosedSessionError{}
	}

	err := s.connection.Close()
	if err != nil {
		return err
	}

	s.isReady = false
	return nil
}
