package repo

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/userreferral"
	"github.com/go-resty/resty"
	"github.com/jinzhu/gorm"
)

const giveReferralRewardPath = "/reward/referral-rewards"

// Repository struct
type Repository struct {
	db                        *gorm.DB
	mapper                    *entityMapper
	giveReferralRewardBaseURL string
	giveReferralRewardClient  *resty.Client
	verifyDeviceClient        *resty.Client
}

// New creates user repository
func New(db *gorm.DB, giveReferralRewardBaseURL string, giveReferralRewardClient *resty.Client, verifyDeviceClient *resty.Client) *Repository {
	repo := Repository{
		db:                        db,
		mapper:                    &entityMapper{},
		giveReferralRewardBaseURL: giveReferralRewardBaseURL,
		giveReferralRewardClient:  giveReferralRewardClient,
		verifyDeviceClient:        verifyDeviceClient,
	}
	return &repo
}

// CreateUser creates user
func (r *Repository) CreateUser(deviceID int64, code string, isVerified bool) (*userreferral.DeviceUser, error) {
	user := &DBDeviceUser{
		DeviceID:   deviceID,
		Code:       code,
		IsVerified: isVerified,
	}
	return r.mapper.dbDeviceUserToDeviceUser(*user), r.db.Create(user).Error
}

// UpdateUserReferrerID updates user's referrerID
func (r *Repository) UpdateUserReferrerID(userID int64, referrerID int64) error {
	return r.db.Model(&DBDeviceUser{ID: userID}).Update("referrer_id", referrerID).Error
}

// GetUserByDevice returns user by deviceID
func (r *Repository) GetUserByDevice(deviceID int64) (*userreferral.DeviceUser, error) {
	user := &DBDeviceUser{}
	err := r.db.Where(&DBDeviceUser{DeviceID: deviceID}).First(user).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, userreferral.NotFoundError{Message: err.Error()}
	} else if err != nil {
		return nil, err
	}
	return r.mapper.dbDeviceUserToDeviceUser(*user), nil
}

// GetUserByCode returns user by referral code
func (r *Repository) GetUserByCode(code string) (*userreferral.DeviceUser, error) {
	user := &DBDeviceUser{}
	err := r.db.Where(&DBDeviceUser{Code: code}).First(user).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, userreferral.NotFoundError{Message: err.Error()}
	} else if err != nil {
		return nil, err
	}
	return r.mapper.dbDeviceUserToDeviceUser(*user), nil
}

// GetReferralCountByUser returns how much a referrer is referred
func (r *Repository) GetReferralCountByUser(referrerID int64) (int, error) {
	var total int
	if err := r.db.Model(&DBDeviceUser{}).Where(&DBDeviceUser{ReferrerID: referrerID}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// IsVerifiedDevice is verifying device by unit deivce token
func (r *Repository) IsVerifiedDevice(verifyURL string, udt string) error {
	resp, err := r.verifyDeviceClient.R().
		SetQueryParams(map[string]string{
			"user_id": udt,
		}).
		Get(verifyURL)

	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return userreferral.APICallError{Message: fmt.Sprintf("isDeviceVerified failed status code : %v", resp.StatusCode())}
	}
	return nil
}

// GiveReferralReward gives reward to referee and referrer
func (r *Repository) GiveReferralReward(ingr userreferral.GiveReferralRewardRequestIngredients) error {
	resp, err := r.giveReferralRewardClient.R().
		SetHeader("Authorization", "Bearer "+ingr.JWT).
		SetFormData(map[string]string{
			"referee_device_id":  strconv.FormatInt(ingr.RefereeDeviceID, 10),
			"referee_reward":     strconv.Itoa(ingr.RefereeReward),
			"referee_title":      ingr.TitleForReferral.TitleForReferee,
			"referrer_device_id": strconv.FormatInt(ingr.ReferrerDeviceID, 10),
			"referrer_reward":    strconv.Itoa(ingr.ReferrerReward),
			"referrer_title":     ingr.TitleForReferral.TitleForReferrer,
			"referrer_max_title": ingr.TitleForReferral.TitleForMaxReferrer,
		}).
		Post(r.giveReferralRewardBaseURL + giveReferralRewardPath)

	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return userreferral.APICallError{Message: fmt.Sprintf("giveReferralPoint failed status code : %v", resp.StatusCode())}
	}
	return nil
}
