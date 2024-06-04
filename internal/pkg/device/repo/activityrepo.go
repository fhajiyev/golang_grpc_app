package repo

import (
	"strconv"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/guregu/dynamo"
)

const (
	keyDeviceID     = "did"
	keyCreatedAt    = "ca"
	keyActivityType = "at"
	keyCampaignID   = "cid"
	keyTTL          = "t"
)

// Activity struct definition
type Activity struct {
	DeviceID   int64   `dynamo:"did,hash"`
	CreatedAt  float64 `dynamo:"ca,range"`
	ActionType string  `dynamo:"at"`
	CampaignID int64   `dynamo:"cid"`
	TTL        int64   `dynamo:"t"`
}

// ActivityRepo type definition
type ActivityRepo struct {
	dynamoTable *dynamo.Table
}

// GetByID func definition
func (r *ActivityRepo) GetByID(deviceID int64) (*device.Activity, error) {
	now := time.Now()

	var activities []*Activity
	err := r.dynamoTable.
		Get(keyDeviceID, deviceID).
		Range(keyCreatedAt, dynamo.GreaterOrEqual, float64(now.AddDate(0, 0, -2).Unix())).
		Filter("$ >= ?", keyTTL, now.Unix()).
		All(&activities)

	if err != nil || len(activities) == 0 {
		return nil, err
	}

	// Convert list of IDs to map
	yesterday := float64(now.AddDate(0, 0, -1).Unix())
	hourAgo := float64(now.Add(-time.Hour).Unix())

	seenCampaignCountForDay := make(map[string]int)
	seenCampaignCountForHour := make(map[string]int)
	seenCampaignIDs := make(map[string]bool)
	for _, activity := range activities {
		cID := strconv.Itoa(int(activity.CampaignID))
		seenCampaignIDs[cID] = true

		if activity.ActionType == string(device.ActivityImpression) {
			if yesterday <= activity.CreatedAt {
				seenCampaignCountForDay[cID]++
			}
			if hourAgo <= activity.CreatedAt {
				seenCampaignCountForHour[cID]++
			}
		}
	}

	return &device.Activity{
		SeenCampaignCountForDay:  seenCampaignCountForDay,
		SeenCampaignCountForHour: seenCampaignCountForHour,
		SeenCampaignIDs:          seenCampaignIDs,
	}, nil
}

// Save func definition
func (r *ActivityRepo) Save(deviceID int64, campaignID int64, activityType device.ActivityType) error {
	now := time.Now()
	ms := (now.UnixNano() / 1000) % 1000000
	activity := &Activity{
		DeviceID:   deviceID,
		CreatedAt:  float64(now.Unix()) + (float64(ms) / 1000000),
		ActionType: string(activityType),
		CampaignID: campaignID,
		TTL:        time.Now().AddDate(0, 0, 2).Unix(),
	}
	return r.dynamoTable.Put(activity).Run()
}

// NewActivityRepo returns new ActivityRepository implementation
func NewActivityRepo(dynamoTable *dynamo.Table) *ActivityRepo {
	return &ActivityRepo{dynamoTable}
}
