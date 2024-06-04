package header

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/metadata"
)

const (
	metaDataKeyAppID           = "buzz-app-id"
	metaDataKeyPublisherUserID = "buzz-publisher-user-id"
	metaDataKeyAccountID       = "buzz-account-id"
	metaDataKeyIFA             = "buzz-ifa"
	metaDataKeyUserAgent       = "user-agent"
)

// GRPCParser struct definition
type GRPCParser struct {
	md     metadata.MD
	parser Parser
}

// NewGRPCParser returns GRPCParser
func NewGRPCParser(md metadata.MD) *GRPCParser {
	return &GRPCParser{md: md, parser: Parser{}}
}

// Auth returns auth from header
func (p *GRPCParser) Auth() (*Auth, error) {
	appID, err := p.getIntValueFromMetadata(headerKeyAppID)
	if err != nil {
		return nil, err
	}

	accountID, err := p.getIntValueFromMetadata(headerKeyAccountID)
	if err != nil {
		return nil, err
	}

	publisherUserID, err := p.getStringValueFromMetadata(headerKeyPublisherUserID)
	if err != nil {
		return nil, err
	}

	ifa, err := p.getStringValueFromMetadata(headerKeyIFA)
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

// UserAgent returns BuzzUserAgent from header
func (p *GRPCParser) UserAgent() (*BuzzUserAgent, error) {
	userAgentStr, err := p.getStringValueFromMetadata(headerKeyUserAgent)
	if err != nil {
		return nil, err
	}

	return p.parser.ParseUserAgent(userAgentStr)
}

func (p *GRPCParser) getIntValueFromMetadata(key string) (int64, error) {
	strValue, err := p.getStringValueFromMetadata(key)
	if err != nil {
		return 0, err
	}

	intValue, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse int value. err: %v. key: %s, value: %s", err, key, strValue)
	}

	return intValue, nil
}

func (p *GRPCParser) getStringValueFromMetadata(key string) (string, error) {
	values := p.md.Get(key)
	if values == nil || len(values) < 1 {
		return "", fmt.Errorf("metadata does not contain value. key: %s", key)
	}

	return values[0], nil
}

// AppendAuthToOutgoingContext returns context containing auth
func AppendAuthToOutgoingContext(ctx context.Context, a Auth) context.Context {
	kv := []string{
		headerKeyAccountID, strconv.FormatInt(a.AccountID, 10),
		headerKeyAppID, strconv.FormatInt(a.AppID, 10),
		headerKeyPublisherUserID, a.PublisherUserID,
		headerKeyIFA, a.IFA,
	}

	return metadata.AppendToOutgoingContext(ctx, kv...)
}

// AppendUserAgentToOutgoingContext returns context containing user-agent
func AppendUserAgentToOutgoingContext(ctx context.Context, uaStr string) context.Context {
	kv := []string{
		headerKeyUserAgent, uaStr,
	}

	return metadata.AppendToOutgoingContext(ctx, kv...)
}
