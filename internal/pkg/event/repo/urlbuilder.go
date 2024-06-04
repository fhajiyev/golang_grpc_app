package repo

import (
	"fmt"
	"net/url"
)

type urlBuilder struct {
	buzzScreenAPIURL string
}

func (b *urlBuilder) buildTrackEventURL(token string) string {
	values := &url.Values{"token": []string{token}}
	return fmt.Sprintf("%s/api/track-event?%s", b.buzzScreenAPIURL, values.Encode())
}

func (b *urlBuilder) buildStatusCheckURL(token string) string {
	values := &url.Values{"token": []string{token}}
	return fmt.Sprintf("%s/api/reward-status?%s", b.buzzScreenAPIURL, values.Encode())
}
