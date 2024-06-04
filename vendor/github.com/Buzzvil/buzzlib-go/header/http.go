package header

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	headerKeyAppID           = "Buzz-App-Id"
	headerKeyPublisherUserID = "Buzz-Publisher-User-Id"
	headerKeyAccountID       = "Buzz-Account-Id"
	headerKeyIFA             = "Buzz-Ifa"
	headerKeyUserAgent       = "User-Agent"
)

// HTTPParser struct definition
type HTTPParser struct {
	header http.Header
	parser Parser
}

// NewHTTPParser returns HTTPParser
func NewHTTPParser(header http.Header) *HTTPParser {
	return &HTTPParser{header: header, parser: Parser{}}
}

// Auth returns auth from header
func (p *HTTPParser) Auth() (*Auth, error) {
	appID, err := p.getIntValueFromHeader(headerKeyAppID)
	if err != nil {
		return nil, err
	}

	accountID, err := p.getIntValueFromHeader(headerKeyAccountID)
	if err != nil {
		return nil, err
	}

	publisherUserID, err := p.getStringValueFromHeader(headerKeyPublisherUserID)
	if err != nil {
		return nil, err
	}

	ifa, err := p.getStringValueFromHeader(headerKeyIFA)
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

// UserAgent returns user agent from header
func (p *HTTPParser) UserAgent() (*BuzzUserAgent, error) {
	userAgentStr, err := p.getStringValueFromHeader(headerKeyUserAgent)
	if err != nil {
		return nil, err
	}

	return p.parser.ParseUserAgent(userAgentStr)
}

func (p *HTTPParser) getIntValueFromHeader(key string) (int64, error) {
	strValue, err := p.getStringValueFromHeader(key)
	if err != nil {
		return 0, err
	}

	intValue, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse int value. err: %v. key: %s, value: %s", err, key, strValue)
	}

	return intValue, nil
}

func (p *HTTPParser) getStringValueFromHeader(key string) (string, error) {
	value := p.header.Get(key)
	if value == "" {
		return "", fmt.Errorf("header does not contain value. key: %s", key)
	}

	return value, nil
}

// AppendAuthToHeader returns header containing auth
func AppendAuthToHeader(header http.Header, auth Auth) http.Header {
	header.Set(headerKeyAccountID, strconv.FormatInt(auth.AccountID, 10))
	header.Set(headerKeyAppID, strconv.FormatInt(auth.AppID, 10))
	header.Set(headerKeyPublisherUserID, auth.PublisherUserID)
	header.Set(headerKeyIFA, auth.IFA)

	return header
}
