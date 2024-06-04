package controller

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
)

const (
	// NotificationTypeFeedAdsCount type definition
	NotificationTypeFeedAdsCount = 100
)

type (
	notificationScheduleFilter interface {
		IsEligible(schedule model.NotificationSchedule) bool
	}

	minVersionCodeFilter struct {
		VersionCode *int
	}

	maxVersionCodeFilter struct {
		VersionCode *int
	}

	eligibleTimeFilter struct {
		Time time.Time
	}
)

// GetNotifications returns notifications. This method is called periodically by the clients.
func GetNotifications(c core.Context) error {
	var req dto.NotificationsRequest
	if err := bindValue(c, &req); err != nil {
		return err
	}

	err := req.UnpackSession()
	if err != nil {
		return common.NewSessionError(err)
	}

	// TODO: Ensure the sign in status using middleware.
	if !req.Session.IsSignedIn() {
		return common.NewSessionError(errors.New("invalid session"))
	}

	if req.GetTypes() == nil {
		return common.NewBindError(errors.New("types is empty"))
	}

	var schedules []model.NotificationSchedule
	buzzscreen.Service.DB.Where(&model.NotificationSchedule{UnitID: req.UnitID}).Find(&schedules)
	eligibleSchedules := filterEligibleSchedules(schedules,
		minVersionCodeFilter{VersionCode: &req.SdkVersion},
		maxVersionCodeFilter{VersionCode: &req.SdkVersion},
		eligibleTimeFilter{Time: time.Now()})
	notifications, err := buildNotifications(c, req.AdsRequest, eligibleSchedules)

	if err != nil {
		switch err.(type) {
		case reward.RemoteError:
			core.Logger.WithError(err).Warnf("GetNotifications() - err: %s", err)
			return common.NewInternalServerError(err)
		default:
			core.Logger.WithError(err).Errorf("GetNotifications() - err: %s", err)
			return common.NewInternalServerError(err)
		}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"notifications": notifications,
	})
}

func buildNotifications(c core.Context, adsRequest dto.AdsRequest, schedules []model.NotificationSchedule) ([]*dto.Notification, error) {
	var notifications []*dto.Notification
	for _, schedule := range schedules {
		switch schedule.NotificationType {
		case model.NotificationTypeFeed:
			totalReward, err := fetchTotalReward(c, adsRequest)
			if err != nil {
				return nil, err
			} else if totalReward <= 0 {
				continue
			}

			payload := map[string]string{"total_reward": strconv.Itoa(totalReward)}
			notification := buildNotification(&schedule, payload)
			notifications = append(notifications, notification)
		}
	}

	return notifications, nil
}

func buildNotification(schedule *model.NotificationSchedule, payload map[string]string) *dto.Notification {
	search := regexp.MustCompile("{[^}]+}")
	description := string(search.ReplaceAllFunc([]byte(schedule.Description), func(key []byte) []byte {
		if value, ok := payload[string(key[1:len(key)-1])]; ok {
			return []byte(value)
		}
		return []byte{}
	}))

	return &dto.Notification{
		Importance:   schedule.Importance.String(),
		Title:        schedule.Title,
		Description:  description,
		InboxSummary: schedule.InboxSummary,
		IconURL:      schedule.IconURL,
		Link:         schedule.Link(),
		Payload:      payload,
	}
}

func fetchTotalReward(c core.Context, adsReq dto.AdsRequest) (int, error) {
	ctx := c.Request().Context()
	if adsReq.UserAgent == "" {
		adsReq.UserAgent = c.Request().UserAgent()
	}

	location := buzzscreen.Service.LocationUseCase.GetClientLocation(c.Request(), adsReq.GetCountry())

	deviceID := adsReq.Session.DeviceID
	unitID := adsReq.UnitID
	adsRes, err := service.GetAdsByStatus(ctx, &adsReq, location, &adsReq, deviceID, unitID, NotificationTypeFeedAdsCount, nil)
	if err != nil {
		return -1, err
	}

	totalReward := 0
	for _, ad := range adsRes.Ads {
		totalReward += ad.LandingReward + ad.ActionReward
	}

	return totalReward, nil
}

func filterEligibleSchedules(schedules []model.NotificationSchedule, filters ...notificationScheduleFilter) []model.NotificationSchedule {
	eligibleSchedules := []model.NotificationSchedule{}
	for _, schedule := range schedules {
		eligible := true
		for _, filter := range filters {
			if !filter.IsEligible(schedule) {
				eligible = false
				break
			}
		}

		if eligible {
			eligibleSchedules = append(eligibleSchedules, schedule)
		}
	}

	return eligibleSchedules
}

// IsEligible func definition
func (filter minVersionCodeFilter) IsEligible(schedule model.NotificationSchedule) bool {
	return schedule.MinVersionCode == nil || (filter.VersionCode != nil && *filter.VersionCode >= *schedule.MinVersionCode)
}

// IsEligible func definition
func (filter maxVersionCodeFilter) IsEligible(schedule model.NotificationSchedule) bool {
	return schedule.MaxVersionCode == nil || (filter.VersionCode != nil && *filter.VersionCode <= *schedule.MaxVersionCode)
}

// IsEligible func definition
func (filter eligibleTimeFilter) IsEligible(schedule model.NotificationSchedule) bool {
	return schedule.Contains(filter.Time)
}
