package repo

import (
	"context"
	"fmt"
	"time"

	rewardsvc "github.com/Buzzvil/buzzapis/go/reward"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/go-redis/cache"
)

// Repository struct definition
type Repository struct {
	client        RewardServiceClient
	mapper        *mapper
	protoBuilder  *protoBuilder
	urlBuilder    *urlBuilder
	entityBuilder *entityBuilder
	redisCache    rediscache.RedisSource
}

const trackingURLCacheExpiration = time.Minute * 5

// GetRewardStatus calls GetRewardStatus to rewardsvc
func (r *Repository) GetRewardStatus(token event.Token, auth header.Auth) (string, error) {
	ctx := context.Background()
	ctx = header.AppendAuthToOutgoingContext(ctx, auth)

	req, err := r.protoBuilder.buildCheckRewardStatusRequest(token)
	if err != nil {
		return "", err
	}

	res, err := r.client.CheckRewardStatus(ctx, req)
	if err != nil {
		return "", err
	}

	return r.mapper.mapToRewardStatus(res.Status)
}

// GetEventsMap calls IssueRewards to rewardsvc and build event data
// TODO support article
func (r *Repository) GetEventsMap(resources []event.Resource, unitID int64, a header.Auth, tokenEncrypter event.TokenEncrypter) (map[int64]event.Events, error) {
	resRewardsMap, err := r.issueRewards(resources, a)
	if err != nil {
		return nil, err
	}

	eventsMap := make(map[int64]event.Events)
	for _, resource := range resources {
		protoRewards, ok := resRewardsMap[resource]
		if !ok {
			continue
		}

		events, err := r.buildEventsForResource(*protoRewards, resource, unitID, tokenEncrypter)
		if err != nil {
			return nil, err
		} else if events == nil || len(events) == 0 {
			continue
		}

		eventsMap[resource.ID] = events
	}

	return eventsMap, nil
}

func (r *Repository) buildEventsForResource(protoRewards rewardsvc.Rewards, resource event.Resource, unitID int64, tokenEncrypter event.TokenEncrypter) (event.Events, error) {
	events := make(event.Events, 0)
	for _, protoReward := range protoRewards.Rewards {
		token, err := r.entityBuilder.buildToken(protoReward, resource, unitID)
		if err != nil {
			return nil, err
		}

		tokenStr, err := tokenEncrypter.Build(*token)
		if err != nil {
			return nil, err
		}

		trackingURL := r.urlBuilder.buildTrackEventURL(tokenStr)
		statusCheckURL := r.urlBuilder.buildStatusCheckURL(tokenStr)

		e, err := r.entityBuilder.buildEvent(protoReward, trackingURL, statusCheckURL)
		if err != nil {
			return nil, err
		}

		events = append(events, *e)
	}

	return events, nil
}

func (r *Repository) issueRewards(resources []event.Resource, a header.Auth) (map[event.Resource]*rewardsvc.Rewards, error) {
	req, err := r.protoBuilder.buildIssueRewardsRequest(resources)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ctx = header.AppendAuthToOutgoingContext(ctx, a)
	res, err := r.client.IssueRewards(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Failed to call IssueRewards. Error may be occured by rewardsvc. err: %s", err)
	}

	return r.protoBuilder.buildRewardsForResourceMap(resources, res.RewardsForResourceMap)
}

// SaveTrackingURL saves trackingURL to cache
func (r *Repository) SaveTrackingURL(deviceID int64, resource event.Resource, trackingURL string) {
	cacheKey := r.getCacheKeyTrackingURL(deviceID, resource)
	r.redisCache.SetCacheAsync(cacheKey, trackingURL, trackingURLCacheExpiration)
}

// GetTrackingURL returns trackingURL from cache
func (r *Repository) GetTrackingURL(deviceID int64, resource event.Resource) (string, error) {
	cacheKey := r.getCacheKeyTrackingURL(deviceID, resource)

	var trackURL string
	err := r.redisCache.GetCache(cacheKey, &trackURL)
	if err == cache.ErrCacheMiss {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return trackURL, nil
}

// DeleteTrackingURL deltes trackingURL in cache
func (r *Repository) DeleteTrackingURL(deviceID int64, resource event.Resource) error {
	cacheKey := r.getCacheKeyTrackingURL(deviceID, resource)

	err := r.redisCache.DeleteCache(cacheKey)
	if err != nil && err != cache.ErrCacheMiss { // return nil for cache miss case
		return err
	}

	return nil
}

func (r *Repository) getCacheKeyTrackingURL(deviceID int64, resource event.Resource) string {
	return fmt.Sprintf("CACHE_GO_TRACKINGURL-%v-%v-%v", deviceID, resource.ID, resource.Type)
}

// New returns Repository struct
func New(client RewardServiceClient, buzzScreenAPIURL string, redisCache rediscache.RedisSource) *Repository {
	m := &mapper{}
	return &Repository{
		client:        client,
		mapper:        m,
		protoBuilder:  &protoBuilder{m},
		urlBuilder:    &urlBuilder{buzzScreenAPIURL},
		entityBuilder: &entityBuilder{m},
		redisCache:    redisCache,
	}
}
