package repo

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/jwt"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/guregu/dynamo"
)

// Repository struct definition
type Repository struct {
	pointTable *dynamo.Table
	mapper     *mapper
}

// GetImpressionPoints returns points array for recent maxPeriod seconds
func (r *Repository) GetImpressionPoints(deviceID int64, maxPeriod reward.Period) []reward.Point {
	now := time.Now().Unix()
	limit := maxRecordsPerHour * maxPeriod.Hours()

	dbPoints := r.getImpressionPoints(deviceID, limit)

	points := make([]reward.Point, 0)
	for _, p := range dbPoints {
		if p.CreatedAt < now-int64(maxPeriod) {
			break
		}

		point, err := r.mapper.ToEntityPoint(p)
		if err != nil {
			continue
		}

		points = append(points, *point)
	}

	return points
}

// Save requests to buzzscreen for saving reward
func (r *Repository) Save(ingr reward.RequestIngredients) (int, error) {
	req := &network.Request{
		URL:    fmt.Sprintf("%s/reward/impression-rewards", env.Config.BuzzconInternalURL),
		Method: "POST",
		Params: &url.Values{
			"unit_device_token": {ingr.UnitDeviceToken},
			"app_id":            {strconv.FormatInt(ingr.AppID, 10)},
			"unit_id":           {strconv.FormatInt(ingr.UnitID, 10)},
			"device_id":         {strconv.FormatInt(ingr.DeviceID, 10)},
			"reward":            {strconv.Itoa(ingr.Reward)},
			"base_reward":       {strconv.Itoa(ingr.BaseReward)},
			"campaign_id":       {strconv.FormatInt(ingr.CampaignID, 10)},
			"campaign_name":     {ingr.CampaignName},
			"slot":              {strconv.Itoa(ingr.Slot)},
			"click_type":        {string(ingr.ClickType)},
		},
		Header: &http.Header{
			"Authorization": {"Bearer " + jwt.GetServiceToken()},
		},
	}
	res, err := req.MakeRequest()
	if err != nil {
		return 0, reward.RemoteError{Err: err}
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		return ingr.Reward, nil // 실제로 받은 리워드
	case http.StatusConflict:
		return 0, reward.DuplicatedError{}
	case http.StatusUnprocessableEntity:
		return 0, reward.UnprocessableError{}
	default:
		return 0, reward.RemoteError{Err: fmt.Errorf("reward.repo.Save() - unknown status code: %d", res.StatusCode)}
	}
}

// New returns new reward repository
func New(pointTable *dynamo.Table) *Repository {
	return &Repository{pointTable, &mapper{}}
}
