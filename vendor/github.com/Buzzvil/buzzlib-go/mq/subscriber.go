package mq

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// Subscriber interface
type Subscriber interface {
	Subscribe(topicHandleMap map[string]func(amqp.Delivery)) error
	QueueSize() (int, error)
	Close() error
}

type subscriber struct {
	session

	queueName  string
	topics     []string
	shutdown   chan bool
	shutdownOK chan bool

	numSubscribers int
	mutex          sync.Mutex
}

// TopicHandler roles handle if routing key is matched regex
type topicHandler struct {
	regex  *regexp.Regexp
	topic  string
	handle func(amqp.Delivery)
}

const (
	deadLetterExchangePrefix string = "buzzvil.deadletter."
	deadLetterQueuePrefix    string = "buzzvil.dead."
)

// CreateSubscriber declares pre-defined routing, and automatically
// attempts to connect to the server.
func CreateSubscriber(addr, queueName string, topics []string, l logger) (Subscriber, error) {
	s := subscriber{
		session: session{
			logger: newMQLoggerWrapper(l),
			done:   make(chan bool),
		},
		queueName:  queueName,
		topics:     topics,
		shutdown:   make(chan bool),
		shutdownOK: make(chan bool),
	}

	s.declareFunc = s.declareQueue

	go s.handleReconnect(addr)
	return &s, nil

}
func (s *subscriber) topicToRegex(topic string) (*regexp.Regexp, error) {
	// topic to regex pattern. reference: https://stackoverflow.com/questions/50679145/
	pattern := topic
	pattern = strings.Replace(pattern, "*", "([^.]+)", -1)
	pattern = strings.Replace(pattern, "#", "([^.]+.?)+", -1)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("Failed to regexp.Compile. topic: %s, err: %v", topic, err)
	}

	return regex, nil
}

func (s *subscriber) declareQueue(ch *amqp.Channel) error {
	err := s.declareBaseExchanges(ch)
	if err != nil {
		return err
	}

	err = s.declareDeadLetterExchange(ch, s.queueName)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(s.queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": deadLetterExchangePrefix + s.queueName,
	})
	if err != nil {
		return err
	}

	for _, topic := range s.topics {
		err = ch.QueueBind(s.queueName, topic, rootExchange, false, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *subscriber) declareDeadLetterExchange(ch *amqp.Channel, queueName string) error {
	exchange := deadLetterExchangePrefix + queueName
	queue := deadLetterQueuePrefix + queueName

	err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare \"%s\" exchange. err: %v", exchange, err)
	}

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare \"%s\" queue. err: %v", queue, err)
	}

	err = ch.QueueBind(queue, "#", exchange, false, nil)
	if err != nil {
		return fmt.Errorf("bind \"%s\" queue to \"%s\" exchange. err: %v", queue, exchange, err)
	}

	return nil
}

// Stream will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
func (s *subscriber) Subscribe(topicHandleMap map[string]func(amqp.Delivery)) error {
	topicHandlers := make([]topicHandler, 0)
	for topic, handle := range topicHandleMap {
		regex, err := s.topicToRegex(topic)
		if err != nil {
			return err
		}

		th := topicHandler{
			topic:  topic,
			handle: handle,
			regex:  regex,
		}
		topicHandlers = append(topicHandlers, th)
	}

	s.increaseSubcriber()
	defer s.decreaseSubscriber()

	return s.subscribe(topicHandlers)
}

func (s *subscriber) subscribe(topicHandlers []topicHandler) error {
	ch, err := s.newChannel()
	if err != nil {
		return err
	}
	notifyChanClose := ch.NotifyClose(make(chan *amqp.Error))

	deliveries, err := ch.Consume(
		s.queueName,
		"",    // Consumer
		false, // Auto-Ack
		false, // Exclusive
		false, // No-local
		false, // No-Wait
		nil,   // Args
	)
	if err != nil {
		return err
	}

	s.logger.Infof("start subscriber")
	defer s.logger.Infof("subscriber closed")

	for {
		select {
		case <-s.shutdown:
			s.shutdownOK <- true
			return ClosedSessionError{}
		case <-s.notifyConnClose:
			return ClosedConnectionError{}
		case <-notifyChanClose:
			return ClosedChannelError{}
		case d := <-deliveries:
			s.handle(topicHandlers, d)
		}
	}
}

func (s *subscriber) handle(topicHandlers []topicHandler, d amqp.Delivery) {
	for _, th := range topicHandlers {
		if th.regex.MatchString(d.RoutingKey) {
			th.handle(d)
			return
		}
	}

	s.logger.Warnf("unroutable Message. routingKey: %s", d.RoutingKey)
	d.Nack(false, false)
}

func (s *subscriber) QueueSize() (int, error) {
	ch, err := s.newChannel()
	if err != nil {
		return 0, err
	}
	defer ch.Close()

	q, err := ch.QueueInspect(s.queueName)
	if err != nil {
		return 0, err
	}

	return q.Messages, nil
}

func (s *subscriber) increaseSubcriber() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.numSubscribers++
}

func (s *subscriber) decreaseSubscriber() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.numSubscribers--
}

func (s *subscriber) shutdownSubscriber() bool {
	s.shutdown <- true
	select {
	case <-s.shutdownOK:
	case <-time.After(time.Second * 2):
		return false
	}

	return true
}

// Close will cleanly shutdown the channel and connection.
func (s *subscriber) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := 0; i < s.numSubscribers; i++ {
		if !s.shutdownSubscriber() {
			break
		}
	}

	return s.close()
}
