package header

import (
	"fmt"
	"reflect"

	"github.com/streadway/amqp"
)

const (
	tableKeyAppID           = "Buzz-App-Id"
	tableKeyPublisherUserID = "Buzz-Publisher-User-Id"
	tableKeyAccountID       = "Buzz-Account-Id"
	tableKeyIFA             = "Buzz-Ifa"
)

// AMQPParser struct definition
type AMQPParser struct {
	table amqp.Table
}

// NewAMQPParser returns AMQPParser
func NewAMQPParser(table amqp.Table) *AMQPParser {
	return &AMQPParser{table}
}

// Auth returns auth from header
func (p *AMQPParser) Auth() (*Auth, error) {
	appID, err := p.getIntValueFromHeader(tableKeyAppID)
	if err != nil {
		return nil, err
	}

	accountID, err := p.getIntValueFromHeader(tableKeyAccountID)
	if err != nil {
		return nil, err
	}

	publisherUserID, err := p.getStringValueFromHeader(tableKeyPublisherUserID)
	if err != nil {
		return nil, err
	}

	ifa, err := p.getStringValueFromHeader(tableKeyIFA)
	if err != nil {
		return nil, err
	}

	return &Auth{
		AppID:           appID,
		AccountID:       accountID,
		PublisherUserID: publisherUserID,
		IFA:             ifa,
	}, nil

}

func (p *AMQPParser) getIntValueFromHeader(key string) (int64, error) {
	value, err := p.getValueFromHeader(key)
	if err != nil {
		return 0, err
	}

	switch value.(type) {
	case int32:
		return int64(value.(int32)), nil
	case int64:
		return value.(int64), nil
	case float32:
		return int64(value.(float32)), nil
	case float64:
		return int64(value.(float64)), nil
	}

	return 0, fmt.Errorf("failed to cast value to int. key: %v, value: %v, type of value: %v", key, value, reflect.TypeOf(value))
}

func (p *AMQPParser) getStringValueFromHeader(key string) (string, error) {
	value, err := p.getValueFromHeader(key)
	if err != nil {
		return "", err
	}

	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("failed to cast value to string. key: %s, value: %v", key, value)
	}

	return strValue, nil
}

func (p *AMQPParser) getValueFromHeader(key string) (interface{}, error) {
	value, exist := p.table[key]
	if !exist {
		return "", fmt.Errorf("header does not contain value for %s", key)
	}

	return value, nil
}

// AppendAuthToAMQPTable returns context containing auth
func AppendAuthToAMQPTable(table amqp.Table, a Auth) amqp.Table {
	table[tableKeyAppID] = a.AppID
	table[tableKeyAccountID] = a.AccountID
	table[tableKeyPublisherUserID] = a.PublisherUserID
	table[tableKeyIFA] = a.IFA

	return table
}
