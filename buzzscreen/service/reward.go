package service

import (
	"context"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func welcomeRewardExists(deviceID int64, unitID int64) (bool, error) {
	reward := model.WelcomeReward{
		DeviceID: deviceID,
		UnitID:   unitID,
	}

	err := buzzscreen.Service.DB.Where(&reward).First(&reward).Error
	if err == nil {
		return true, nil
	}

	if err == gorm.ErrRecordNotFound {
		err = nil
	}

	return false, err
}

func saveWelcomeReward(deviceID int64, unitDeviceToken string, wrc *app.WelcomeRewardConfig) {
	reward := model.WelcomeReward{
		DeviceID:        deviceID,
		UnitID:          wrc.UnitID,
		UnitDeviceToken: unitDeviceToken,
		Status:          model.StatusPending,
		Amount:          wrc.Amount,
		ConfigID:        wrc.ID,
	}

	err := buzzscreen.Service.DB.Save(&reward).Error
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 { //Duplicate entry
			core.Logger.Warnf("saveWelcomeReward() - Reward is already given to (%v-%v) on unit %v, config %v", deviceID, unitDeviceToken, wrc.UnitID, wrc.ID)
		} else {
			core.Logger.Errorf("saveWelcomeReward() - error with device (%v-%v) on unit %v, config %v, reward amount %v: %v", deviceID, unitDeviceToken, wrc.UnitID, wrc.ID, wrc.Amount, err)
		}
	}
}

// GiveWelcomeReward give welcome reward to a devie if it has a receivable welcome reward
func GiveWelcomeReward(ctx context.Context, profile *device.Profile, unitDeviceToken string, unitID int64, country string) {
	deviceID := profile.ID
	unitRegisterSeconds := (*profile.UnitRegisteredSeconds)[unitID]
	appUseCase := buzzscreen.Service.AppUseCase
	// wrc will have reward amount > 0 with given unitRegisterSecond, or wrc is null.
	wrc, err := appUseCase.GetRewardableWelcomeRewardConfig(ctx, unitID, country, unitRegisterSeconds)
	if wrc == nil || err != nil {
		return
	}
	rewardExists, err := welcomeRewardExists(deviceID, unitID)

	if rewardExists {
		return
	}
	if err != nil {
		core.Logger.Errorf("GiveWelcomeReward() - %v(%v-%v), Checking if reward is already given failed, try save reward anyway", unitID, deviceID, unitDeviceToken)
	}

	saveWelcomeReward(deviceID, unitDeviceToken, wrc)
}
