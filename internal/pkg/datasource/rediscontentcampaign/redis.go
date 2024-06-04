package rediscontentcampaign

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// RedisContentCampaign maintains redis client
type RedisContentCampaign struct {
	client *redis.Client
}

const (
	dataTypeImpression = "imp"
	dataTypeClick      = "clk"
	dateHourLayout     = "2006-01-02:15"
)

// IncreaseImpression increase impression count
func (r *RedisContentCampaign) IncreaseImpression(campaignID int64, unitID int64) error {
	return r.increase(campaignID, unitID, dataTypeImpression)
}

// IncreaseClick increase click count
func (r *RedisContentCampaign) IncreaseClick(campaignID int64, unitID int64) error {
	return r.increase(campaignID, unitID, dataTypeClick)
}

func (r *RedisContentCampaign) increase(campaignID int64, unitID int64, dataType string) error {
	dateHour := time.Now()
	hashKey := getHashKey(dateHour)

	pipeline := r.client.Pipeline()
	pipeline.HIncrBy(hashKey, getKeyCampaign(campaignID, "all", dataType), 1)
	pipeline.HIncrBy(hashKey, getKeyCampaign(campaignID, unitID, dataType), 1)
	pipeline.HIncrBy(getCampaignHashKey(campaignID), getKeyTotal(dataType), 1)

	if _, err := pipeline.Exec(); err != nil {
		return err
	}

	return nil
}

func getKeyTotal(dataType string) string {
	return fmt.Sprintf("total:%v", dataType)
}

func getKeyCampaign(campaignID int64, obj interface{}, dataType string) string {
	return fmt.Sprintf("cam:%v:%v:%v", campaignID, obj, dataType)
}

func getHashKey(dateHour time.Time) string {
	return fmt.Sprintf("stat:%v", dateHourToString(dateHour))
}

func getCampaignHashKey(campaignID int64) string {
	return fmt.Sprintf("stat:cam:%v", campaignID)
}

func dateHourToString(dateHour time.Time) string {
	return dateHour.Format(dateHourLayout)
}

// NewSource returns RedisContentCmapaign struct
func NewSource(client *redis.Client) *RedisContentCampaign {
	return &RedisContentCampaign{client}
}
