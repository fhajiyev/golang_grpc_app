package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbapp"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	uuid "github.com/satori/go.uuid"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/stretchr/testify/require"
)

const (
	welcomeRewardConfigCacheKeyFormat = "CACHE_GO_WRCS-%v"
)

var userOrder int64 = 0

func generateDevice(appID int64) dbdevice.Device {
	userOrder += 1
	device := dbdevice.Device{
		AppID:           appID,
		UnitDeviceToken: fmt.Sprintf("WRTestUser-%d", userOrder),
		IFA:             uuid.NewV4().String(),
	}
	buzzscreen.Service.DB.Save(&device)
	deviceRegisterSeconds := time.Now().Unix()
	tests.SetDeviceRegisteredSeconds(device.ID, deviceRegisterSeconds)
	return device
}

// DeleteWelcomeRewardCache check if cache has been set up to 1 second with 50 ms interval,
// and if the cache has been set, delete the cache.
// Return true if the cache has been deleted, false otherwise
func DeleteWelcomeRewardCache(unitID int64) bool {
	cacheKey := fmt.Sprintf(welcomeRewardConfigCacheKeyFormat, unitID)
	cacheSource := rediscache.NewSource(buzzscreen.Service.Redis)
	var configCache repo.WelcomeRewardConfigCache
	for i := 0; i < 200; i++ {
		time.Sleep(time.Millisecond * 50)
		if err := cacheSource.GetCache(cacheKey, &configCache); err == nil {
			rediscache.NewSource(buzzscreen.Service.Redis).DeleteCache(cacheKey)
			return true
		}
	}
	return false
}

func getDefaultWelcomeRewardConfig(unitID int64, country *string) *dbapp.WelcomeRewardConfig {
	startTime := time.Now().Add(-time.Hour)
	endTime := startTime.Add(time.Hour * 23)
	return &(dbapp.WelcomeRewardConfig{
		UnitID:    		unitID,
		Amount:         100,
		StartTime:      &startTime,
		EndTime:   	&endTime,
		Country:   	country,
	})
}

func checkWelcomeRewardGiven(t *testing.T, deviceID int64, unitID int64, shouldGiven bool) *model.WelcomeReward {
	reward := model.WelcomeReward{
		DeviceID: int64(deviceID),
	}
	err := buzzscreen.Service.DB.Where("device_id = ? AND unit_id = ?", deviceID, unitID).First(&reward).Error
	if err == nil && shouldGiven == false {
		t.Fatalf("checkWelcomeRewardGiven() - should not be recevied for device - %v", reward.DeviceID)
	} else if err != nil && shouldGiven {
		t.Fatalf("checkWelcomeRewardGiven() - should be given with id: %v, err: %s", reward.DeviceID, err)
	}
	return &reward
}

func TestWelcomeUnitRegisteredSecondOnAppUnitSameID(t *testing.T) {
	var unitID int64 = 33337
	var deviceRegisterSeconds int64 = 377
	device := generateDevice(11111)
	defer buzzscreen.Service.DB.Delete(&device)
	deviceID := device.ID

	tests.SetDeviceRegisteredSeconds(deviceID, deviceRegisterSeconds)
	profile := tests.GetProfileByID(deviceID)
	_, ok := (*profile.UnitRegisteredSeconds)[unitID]
	require.Equal(t, ok, false)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, unitID, "welcomeRewardUnitTest")
	require.Equal(t, (*profile.UnitRegisteredSeconds)[unitID], deviceRegisterSeconds, "TestUnitRegisteredSecondOnAppUnitSameID - Register Seconds")

}

func TestUnitRegisteredSecondOnAppUnitDifferentID(t *testing.T) {
	var unitID int64 = 33337
	var appID int64 = 33338
	var deviceRegisterSeconds int64 = 377
	device := generateDevice(11111)
	defer buzzscreen.Service.DB.Delete(&device)
	deviceID := device.ID

	tests.SetDeviceRegisteredSeconds(deviceID, deviceRegisterSeconds)
	profile := tests.GetProfileByID(deviceID)
	_, ok := (*profile.UnitRegisteredSeconds)[unitID]
	require.Equal(t, ok, false)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	require.NotEqual(t, (*profile.UnitRegisteredSeconds)[unitID], deviceRegisterSeconds, "TestUnitRegisteredSecondOnAppUnitSameID - Register Seconds")

}

func TestWelcomeRewardCampaignCacheing(t *testing.T) {
	var unitID int64 = 33337
	var appID int64 = 33338
	var rewardAmount = 37
	country := "DS"
	device1 := generateDevice(11111)
	device2 := generateDevice(11112)
	defer buzzscreen.Service.DB.Delete(&device1)
	defer buzzscreen.Service.DB.Delete(&device2)

	wrc := getDefaultWelcomeRewardConfig(unitID, &country)
	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now().Add(time.Hour)
	wrc.StartTime = &startTime
	wrc.EndTime = &endTime
	wrc.Amount = rewardAmount
	buzzscreen.Service.DB.Save(&wrc)
	defer buzzscreen.Service.DB.Delete(&wrc)

	profile1 := tests.GetProfileByID(device1.ID)
	service.UpdateProfileUnitRegisterSeconds(profile1, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile1, device1.UnitDeviceToken, unitID, country)
	t.Log("first reward check")
	reward := checkWelcomeRewardGiven(t, device1.ID, unitID, true)

	require.Equal(t, unitID, reward.UnitID, "TestSingleWelcomeRewardConstant - UnitID")
	require.Equal(t, device1.ID, reward.DeviceID, "TestSingleWelcomeRewardConstant - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestSingleWelcomeRewardConstant - Status")
	require.Equal(t, rewardAmount, reward.Amount, "TestSingleWelcomeRewardConstant - Amount")
	require.Equal(t, wrc.ID, reward.ConfigID, "TestSingleWelcomeRewardConstant - ConfigID")

	profile2 := tests.GetProfileByID(device2.ID)
	service.UpdateProfileUnitRegisterSeconds(profile2, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile2, device2.UnitDeviceToken, unitID, country)
	t.Log("second reward check")
	reward = checkWelcomeRewardGiven(t, device2.ID, unitID, true)

	require.Equal(t, unitID, reward.UnitID, "TestSingleWelcomeRewardConstant - UnitID")
	require.Equal(t, device2.ID, reward.DeviceID, "TestSingleWelcomeRewardConstant - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestSingleWelcomeRewardConstant - Status")
	require.Equal(t, rewardAmount, reward.Amount, "TestSingleWelcomeRewardConstant - Amount")
	require.Equal(t, wrc.ID, reward.ConfigID, "TestSingleWelcomeRewardConstant - ConfigID")
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")

}

func TestSingleWelcomeRewardConstant(t *testing.T) {
	var unitID int64 = 33337
	var appID int64 = 33338
	var rewardAmount = 37
	country := "DS"
	device := generateDevice(11111)
	defer buzzscreen.Service.DB.Delete(&device)
	deviceID := device.ID
	wrc := getDefaultWelcomeRewardConfig(unitID, &country)
	startTime := time.Now().Add(time.Hour)
	endTime := startTime.Add(time.Hour * 24)
	wrc.StartTime, wrc.EndTime = &startTime, &endTime
	wrc.Amount = rewardAmount
	buzzscreen.Service.DB.Save(&wrc)
	defer buzzscreen.Service.DB.Delete(&wrc)
	profile := tests.GetProfileByID(deviceID)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	checkWelcomeRewardGiven(t, deviceID, unitID, false) // WRC 가 미래 시간 이라서 실패

	startTime = time.Now().Add(-time.Hour * 24)
	endTime = time.Now().Add(-time.Hour)
	wrc.StartTime = &startTime
	wrc.EndTime = &endTime
	buzzscreen.Service.DB.Save(&wrc)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	checkWelcomeRewardGiven(t, deviceID, unitID, false) // WRC 가 과거 시간 이라서 실패

	startTime = time.Now().Add(-time.Hour)
	endTime = time.Now().Add(time.Hour)
	wrc.StartTime = &startTime
	wrc.EndTime = &endTime
	buzzscreen.Service.DB.Save(&wrc)
	threeDaysBefore := time.Now().Add(-time.Hour*24*3 - time.Hour)
	err := tests.SetDeviceUnitRegisteredSeconds(deviceID, unitID, threeDaysBefore.Unix())
	if err != nil {
		t.Fatalf("TestSingleWelcomeRewardConstant() - while setting device URS with err - %s", err)
	}
	profile = tests.GetProfileByID(deviceID) // get modified profile
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	checkWelcomeRewardGiven(t, deviceID, unitID, false) // Registered seconds < WRC.StartTime

	err = tests.SetDeviceUnitRegisteredSeconds(deviceID, unitID, time.Now().Unix())
	if err != nil {
		t.Fatalf("TestSingleWelcomeRewardConstant() - while setting device URS with err - %s", err)
	}
	fakeUnitID := unitID + 1
	profile = tests.GetProfileByID(deviceID) // get modified profile
	service.UpdateProfileUnitRegisterSeconds(profile, fakeUnitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, fakeUnitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(fakeUnitID), "Cache deletion failed!")
	checkWelcomeRewardGiven(t, deviceID, fakeUnitID, false) // Failed due to unit without WRC

	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, "JP")
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	checkWelcomeRewardGiven(t, deviceID, unitID, false) // Failed due to country without WRC

	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")

	reward := checkWelcomeRewardGiven(t, deviceID, unitID, true)

	require.Equal(t, unitID, reward.UnitID, "TestSingleWelcomeRewardConstant - UnitID")
	require.Equal(t, deviceID, reward.DeviceID, "TestSingleWelcomeRewardConstant - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestSingleWelcomeRewardConstant - Status")
	require.Equal(t, rewardAmount, reward.Amount, "TestSingleWelcomeRewardConstant - Amount")
	require.Equal(t, wrc.ID, reward.ConfigID, "TestSingleWelcomeRewardConstant - Amount")

}

func TestInfiniteEnddateWelcomeRewardConstant(t *testing.T) {
	var unitID int64 = 45000
	var appID int64 = 45000
	var rewardAmount = 45
	country := "DS"
	device := generateDevice(12345)
	defer buzzscreen.Service.DB.Delete(&device)
	deviceID := device.ID
	wrc := getDefaultWelcomeRewardConfig(unitID, &country)
	startTime := time.Now().Add(-time.Hour*24*3)
	maxNumRewards := 100
	wrc.StartTime = &startTime
	wrc.EndTime = nil
	wrc.Amount = rewardAmount
	wrc.MaxNumRewards = &maxNumRewards

	buzzscreen.Service.DB.Save(&wrc)
	defer buzzscreen.Service.DB.Delete(&wrc)

	profile := tests.GetProfileByID(deviceID)

	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	reward := checkWelcomeRewardGiven(t, deviceID, unitID, true)

	require.Equal(t, unitID, reward.UnitID, "TestSingleWelcomeRewardConstant - UnitID")
	require.Equal(t, deviceID, reward.DeviceID, "TestSingleWelcomeRewardConstant - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestSingleWelcomeRewardConstant - Status")
	require.Equal(t, rewardAmount, reward.Amount, "TestSingleWelcomeRewardConstant - Amount")
	require.Equal(t, wrc.ID, reward.ConfigID, "TestSingleWelcomeRewardConstant - Amount")
}

func TestWelcomeRewardCountry(t *testing.T) {
	var unitID int64 = 33337
	var appID int64 = 33338
	device := generateDevice(11111)
	defer buzzscreen.Service.DB.Delete(&device)
	deviceID := device.ID

	secondDevice := generateDevice(11112)
	secondDeviceID := secondDevice.ID

	country1 := "DX"
	country2 := "ZX"
	var rewardAmount1 = 37
	var rewardAmount2 = 13
	startTime := time.Now().Add(-time.Hour * 24)
	endTime := time.Now().Add(time.Hour * 24)

	wrc1 := getDefaultWelcomeRewardConfig(unitID, &country1)
	wrc1.StartTime, wrc1.EndTime = &startTime, &endTime
	wrc1.Amount = rewardAmount1
	buzzscreen.Service.DB.Save(&wrc1)
	defer buzzscreen.Service.DB.Delete(&wrc1)

	secondProfile := tests.GetProfileByID(secondDeviceID)
	service.UpdateProfileUnitRegisterSeconds(secondProfile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), secondProfile, device.UnitDeviceToken, unitID, "")
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	// reward not given even though input country is empty and there is no global WRC on the unit,
	// because there is a uniuqe WRC on the unit
	reward := checkWelcomeRewardGiven(t, secondDeviceID, unitID, true)

	require.Equal(t, unitID, reward.UnitID, "TestWelcomeRewardCountry - UnitID")
	require.Equal(t, secondDeviceID, reward.DeviceID, "TestWelcomeRewardCountry - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestWelcomeRewardCountry - Status")
	require.Equal(t, rewardAmount1, reward.Amount, "TestWelcomeRewardCountry - Amount")
	require.Equal(t, wrc1.ID, reward.ConfigID, "TestWelcomeRewardCountry - Amount")

	wrc2 := getDefaultWelcomeRewardConfig(unitID, &country2)
	wrc2.StartTime, wrc2.EndTime = &startTime, &endTime
	wrc2.Amount = rewardAmount2
	buzzscreen.Service.DB.Save(&wrc2)
	defer buzzscreen.Service.DB.Delete(&wrc2)

	profile := tests.GetProfileByID(deviceID)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, "")
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	// reward not given because there are multiple country with WRC on the unit
	// but country is empty and there is no global WRC
	reward = checkWelcomeRewardGiven(t, deviceID, unitID, false)

	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country1)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	reward = checkWelcomeRewardGiven(t, deviceID, unitID, true)

	require.Equal(t, unitID, reward.UnitID, "TestWelcomeRewardCountry - UnitID")
	require.Equal(t, deviceID, reward.DeviceID, "TestWelcomeRewardCountry - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestWelcomeRewardCountry - Status")
	require.Equal(t, rewardAmount1, reward.Amount, "TestWelcomeRewardCountry - Amount")
	require.Equal(t, wrc1.ID, reward.ConfigID, "TestWelcomeRewardCountry - ConfigID")

	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country2)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	var anotherReward model.WelcomeReward
	err := buzzscreen.Service.DB.Where("device_id = ? AND unit_id = ? AND id != ?", deviceID, unitID, reward.ID).First(&anotherReward).Error
	if err == nil {
		t.Fatalf(
			"TestWelcomeRewardCountry() - More than  one reward given: devicd id %d, unit %d, amount %d",
			anotherReward.DeviceID,
			anotherReward.UnitID,
			anotherReward.Amount,
		)
	}
}

func TestWelcomeRewardGlobal(t *testing.T) {
	var unitID int64 = 33337
	var appID int64 = 33338
	device := generateDevice(11111)
	defer buzzscreen.Service.DB.Delete(&device)
	deviceID := device.ID

	someCountry := "DD"
	startTime := time.Now().Add(-time.Hour * 24)
	endTime := time.Now().Add(time.Hour * 24)

	wrcGlobal := getDefaultWelcomeRewardConfig(unitID, nil)
	wrcGlobal.StartTime, wrcGlobal.EndTime = &startTime, &endTime
	wrcGlobal.Amount = 3
	buzzscreen.Service.DB.Save(&wrcGlobal)
	defer buzzscreen.Service.DB.Delete(&wrcGlobal)

	wrcCountry := getDefaultWelcomeRewardConfig(unitID, &someCountry)
	wrcCountry.StartTime, wrcCountry.EndTime = &startTime, &endTime
	wrcCountry.Amount = 7
	buzzscreen.Service.DB.Save(&wrcCountry)
	defer buzzscreen.Service.DB.Delete(&wrcCountry)

	profile := tests.GetProfileByID(deviceID)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, "XX")
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	reward := checkWelcomeRewardGiven(t, deviceID, unitID, true)

	require.Equal(t, unitID, reward.UnitID, "TestWelcomeRewardGlobal - UnitID")
	require.Equal(t, deviceID, reward.DeviceID, "TestWelcomeRewardGlobal - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestWelcomeRewardGlobal - Status")
	require.Equal(t, wrcGlobal.Amount, reward.Amount, "TestWelcomeRewardGlobal - Amount")
	require.Equal(t, wrcGlobal.ID, reward.ConfigID, "TestWelcomeRewardGlobal - ConfigID")

}

func TestWelcomeRewardRetention(t *testing.T) {
	var unitID int64 = 33337
	var appID int64 = 33338
	device := generateDevice(11111)
	defer buzzscreen.Service.DB.Delete(&device)
	deviceID := device.ID

	country := "DD"
	startTime := time.Now().Add(-time.Hour * 72)
	endTime := time.Now().Add(time.Hour * 24)
	wrc := getDefaultWelcomeRewardConfig(unitID, &country)
	wrc.StartTime, wrc.EndTime = &startTime, &endTime
	wrc.Amount = 3
	wrc.RetentionDays = 2
	buzzscreen.Service.DB.Save(&wrc)
	defer buzzscreen.Service.DB.Delete(&wrc)

	profile := tests.GetProfileByID(deviceID)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	// Reward not given due to lacking retention days
	checkWelcomeRewardGiven(t, deviceID, unitID, false)

	unitRegisterSeconds := time.Now().Add(-time.Hour * 47).Unix()
	tests.SetDeviceUnitRegisteredSeconds(deviceID, unitID, unitRegisterSeconds)
	profile = tests.GetProfileByID(deviceID)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	// Reward still not given due to lacking retention days
	checkWelcomeRewardGiven(t, deviceID, unitID, false)

	unitRegisterSeconds = time.Now().Add(-time.Hour * 48).Unix()
	tests.SetDeviceUnitRegisteredSeconds(deviceID, unitID, unitRegisterSeconds)
	profile = tests.GetProfileByID(deviceID)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	reward := checkWelcomeRewardGiven(t, deviceID, unitID, true)

	require.Equal(t, unitID, reward.UnitID, "TestWelcomeRewardRetention - UnitID")
	require.Equal(t, deviceID, reward.DeviceID, "TestWelcomeRewardRetention - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestWelcomeRewardRetention - Status")
	require.Equal(t, wrc.Amount, reward.Amount, "TestWelcomeRewardRetention - Amount")
	require.Equal(t, wrc.ID, reward.ConfigID, "TestWelcomeRewardRetention - ConfigID")
}

func TestEndingWelcomeRewardRetention(t *testing.T) {
	var unitID int64 = 33337
	var appID int64 = 33338
	device := generateDevice(11111)
	defer buzzscreen.Service.DB.Delete(&device)
	deviceID := device.ID

	country := "DD"
	startTime := time.Now().Add(-time.Hour * 120)
	endTime := time.Now().Add(-time.Hour * 49)

	wrc := getDefaultWelcomeRewardConfig(unitID, &country)
	wrc.StartTime, wrc.EndTime = &startTime, &endTime
	wrc.Amount = 3
	wrc.RetentionDays = 2
	buzzscreen.Service.DB.Save(&wrc)
	defer buzzscreen.Service.DB.Delete(&wrc)

	profile := tests.GetProfileByID(deviceID)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	// Reward not given because device is registerd after the WRC period
	checkWelcomeRewardGiven(t, deviceID, unitID, false)

	unitRegisterSeconds := time.Now().Add(-time.Hour * 50).Unix()
	tests.SetDeviceUnitRegisteredSeconds(deviceID, unitID, unitRegisterSeconds)
	profile = tests.GetProfileByID(deviceID)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	// Reward not given because WRC has already ended
	checkWelcomeRewardGiven(t, deviceID, unitID, false)
	endTime = time.Now().Add(-time.Hour * 47)
	wrc.EndTime = &endTime
	buzzscreen.Service.DB.Save(&wrc)
	unitRegisterSeconds = time.Now().Add(-time.Hour * 50).Unix()
	tests.SetDeviceUnitRegisteredSeconds(deviceID, unitID, unitRegisterSeconds)
	profile = tests.GetProfileByID(deviceID)
	service.UpdateProfileUnitRegisterSeconds(profile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), profile, device.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	reward := checkWelcomeRewardGiven(t, deviceID, unitID, true)

	require.Equal(t, unitID, reward.UnitID, "TestEndingWelcomeRewardRetention - UnitID")
	require.Equal(t, deviceID, reward.DeviceID, "TestEndingWelcomeRewardRetention - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestEndingWelcomeRewardRetention - Status")
	require.Equal(t, wrc.Amount, reward.Amount, "TestEndingWelcomeRewardRetention - Amount")
	require.Equal(t, wrc.ID, reward.ConfigID, "TestEndingWelcomeRewardRetention - ConfigID")

}

func TestMultipleWelcomeRewardRetention(t *testing.T) {
	var unitID int64 = 33337
	var appID int64 = 33338

	country := "DD"
	startTime := time.Now().Add(-time.Hour * 120)
	endTime := time.Now().Add(-time.Hour * 48)

	endingWrc := getDefaultWelcomeRewardConfig(unitID, &country)
	endingWrc.StartTime, endingWrc.EndTime = &startTime, &endTime
	endingWrc.Amount = 3
	endingWrc.RetentionDays = 3
	buzzscreen.Service.DB.Save(&endingWrc)
	defer buzzscreen.Service.DB.Delete(&endingWrc)

	endTime2 := time.Now().Add(time.Hour * 1)

	newerWrc := getDefaultWelcomeRewardConfig(unitID, &country)
	newerWrc.StartTime = endingWrc.EndTime
	newerWrc.EndTime = &endTime2
	newerWrc.Amount = 23
	newerWrc.RetentionDays = 1
	buzzscreen.Service.DB.Save(&newerWrc)
	defer buzzscreen.Service.DB.Delete(&newerWrc)

	olderDevice := generateDevice(11111)
	defer buzzscreen.Service.DB.Delete(&olderDevice)
	unitRegisterSeconds := time.Now().Add(-time.Hour * 73).Unix()
	tests.SetDeviceUnitRegisteredSeconds(olderDevice.ID, unitID, unitRegisterSeconds)

	newerDevice := generateDevice(11111)
	defer buzzscreen.Service.DB.Delete(&newerDevice)
	unitRegisterSeconds = time.Now().Add(-time.Hour * 25).Unix()
	tests.SetDeviceUnitRegisteredSeconds(newerDevice.ID, unitID, unitRegisterSeconds)

	olderProfile := tests.GetProfileByID(olderDevice.ID)
	service.UpdateProfileUnitRegisterSeconds(olderProfile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), olderProfile, olderDevice.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	reward := checkWelcomeRewardGiven(t, olderDevice.ID, unitID, true)
	require.Equal(t, unitID, reward.UnitID, "TestMultipleWelcomeRewardRetention - UnitID")
	require.Equal(t, olderDevice.ID, reward.DeviceID, "TestMultipleWelcomeRewardRetention - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestMultipleWelcomeRewardRetention - Status")
	require.Equal(t, endingWrc.Amount, reward.Amount, "TestMultipleWelcomeRewardRetention - Amount")
	require.Equal(t, endingWrc.ID, reward.ConfigID, "TestMultipleWelcomeRewardRetention - ConfigID")

	newerProfile := tests.GetProfileByID(newerDevice.ID)
	service.UpdateProfileUnitRegisterSeconds(newerProfile, unitID, appID, "welcomeRewardUnitTest")
	service.GiveWelcomeReward(context.Background(), newerProfile, newerDevice.UnitDeviceToken, unitID, country)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")
	reward = checkWelcomeRewardGiven(t, newerDevice.ID, unitID, true)
	require.Equal(t, unitID, reward.UnitID, "TestMultipleWelcomeRewardRetention - UnitID")
	require.Equal(t, newerDevice.ID, reward.DeviceID, "TestMultipleWelcomeRewardRetention - DeviceID")
	require.Equal(t, model.StatusPending, reward.Status, "TestMultipleWelcomeRewardRetention - Status")
	require.Equal(t, newerWrc.Amount, reward.Amount, "TestMultipleWelcomeRewardRetention - Amount")
	require.Equal(t, newerWrc.ID, reward.ConfigID, "TestMultipleWelcomeRewardRetention - ConfigID")
}
