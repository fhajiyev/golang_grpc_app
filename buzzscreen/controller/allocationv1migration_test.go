package controller_test

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbapp"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	uuid "github.com/satori/go.uuid"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/Buzzvil/go-test/test"
	"github.com/stretchr/testify/require"
)

const (
	welcomeRewardConfigCacheKeyFormat = "CACHE_GO_WRCS-%v"
)

var userOrder int64 = 0

func generateUniqueUDT() string {
	userOrder += 1
	return fmt.Sprintf("TestUser-%d", userOrder)
}

func getDefaultWelcomeRewardConfig(unitID int64) *dbapp.WelcomeRewardConfig {
	startTime := time.Now().Add(-time.Hour)
	endTime := startTime.Add(time.Hour * 23)
	return &(dbapp.WelcomeRewardConfig{
		UnitID:    unitID,
		Amount:    100,
		StartTime: &startTime,
		EndTime:   &endTime,
		Country:   nil,
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

// DeleteWelcomeRewardCache check if cache has been set up to 1 second with 50 ms interval,
// and if the cache has been set, delete the cache.
// Return true if the cache has been deleted, false otherwise
func DeleteWelcomeRewardCache(unitID int64) bool {
	cacheKey := fmt.Sprintf(welcomeRewardConfigCacheKeyFormat, unitID)
	cacheSource := rediscache.NewSource(buzzscreen.Service.Redis)
	var configCache repo.WelcomeRewardConfigCache
	for i := 0; i < 20; i++ {
		time.Sleep(time.Millisecond * 50)
		if err := cacheSource.GetCache(cacheKey, &configCache); err == nil {
			rediscache.NewSource(buzzscreen.Service.Redis).DeleteCache(cacheKey)
			return true
		}
	}
	return false
}

/**
엉뚱한 UnitID
*/
func TestWelcomeRewardNotReceivable(t *testing.T) {
	udt := generateUniqueUDT()
	ifa := uuid.NewV4().String()
	migrationAppID := tests.HsKrAppID
	unitID := getAppV1DefaultUnit(migrationAppID)

	params := buildV1BaseTestRequest()
	deviceResponce, _ := execTestDevice(t, migrationAppID, udt, ifa)
	deviceID := int64(deviceResponce.Result["device_id"].(float64))

	params.Set("device_id", strconv.FormatInt(deviceID, 10))
	params.Set("app_id", strconv.FormatInt(migrationAppID, 10))

	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID), "Cache deletion failed!")

	reward := model.WelcomeReward{
		DeviceID: int64(deviceID),
	}
	err := buzzscreen.Service.DB.Where("device_id = ?", reward.DeviceID).First(&reward).Error
	if err == nil {
		t.Fatalf("TestWelcomeRewardNotReceivable() - should not be recevied for device - %v", deviceID)
	}
}

/**
"com.buzzvil.old" -> "com.buzzvil.new"
*/
func TestWelcomeRewardConstant(t *testing.T) {
	udt := generateUniqueUDT()
	ifa := uuid.NewV4().String()

	params := buildV1BaseTestRequest()
	// 미래 시간으로 WRC Insert
	rewardUnitID := getAppV1DefaultUnit(tests.TestAppID2)
	wrc := getDefaultWelcomeRewardConfig(rewardUnitID)
	startTime := time.Now().Add(time.Hour)
	endTime := startTime.Add(time.Hour * 24)
	wrc.StartTime, wrc.EndTime = &startTime, &endTime
	rewardUnitID, rewardAmount := wrc.UnitID, wrc.Amount

	buzzscreen.Service.DB.Save(&wrc)
	defer buzzscreen.Service.DB.Delete(&wrc)

	deviceResponse, _ := execTestDevice(t, tests.TestAppID2, udt, ifa)
	deviceID := int64(deviceResponse.Result["device_id"].(float64))
	tests.SetDeviceUnitRegisteredSeconds(deviceID, rewardUnitID, time.Now().Unix())
	device := dbdevice.Device{
		ID: deviceID,
	}
	err := buzzscreen.Service.DB.Where(&device).First(&device).Error
	if err != nil {
		t.Fatalf("TestWelcomeRewardConstant() - with err - %s", err)
	}

	params.Set("device_id", strconv.FormatInt(deviceID, 10))
	params.Set("app_id", strconv.FormatInt(tests.TestAppID2, 10))

	time.Sleep(time.Second)
	checkWelcomeRewardGiven(t, deviceID, rewardUnitID, false) // WRC 가 미래 시간 이라서 실패

	startTime = time.Now().Add(-time.Hour * 24)
	endTime = time.Now().Add(-time.Hour)
	wrc.StartTime = &startTime
	wrc.EndTime = &endTime
	buzzscreen.Service.DB.Save(&wrc)

	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(rewardUnitID), "Cache deletion failed!")
	checkWelcomeRewardGiven(t, deviceID, rewardUnitID, false) // WRC 가 과거 시간 이라서 실패

	startTime = time.Now().Add(-time.Hour)
	endTime = time.Now().Add(time.Hour)
	wrc.StartTime = &startTime
	wrc.EndTime = &endTime
	buzzscreen.Service.DB.Save(&wrc)

	threeDaysBefore := time.Now().Add(-time.Hour*24*3 - time.Hour)
	err = tests.SetDeviceUnitRegisteredSeconds(deviceID, rewardUnitID, threeDaysBefore.Unix())
	if err != nil {
		t.Fatalf("TestWelcomeRewardConstant() - while setting device URS with err - %s", err)
	}
	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(rewardUnitID), "Cache deletion failed!")
	checkWelcomeRewardGiven(t, deviceID, rewardUnitID, false) // Registered seconds < WRC.StartTime

	err = tests.SetDeviceUnitRegisteredSeconds(deviceID, rewardUnitID, time.Now().Unix())
	if err != nil {
		t.Fatalf("TestWelcomeRewardConstant() - while setting device URS with err - %s", err)
	}
	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(rewardUnitID), "Cache deletion failed!")
	reward := checkWelcomeRewardGiven(t, deviceID, rewardUnitID, true)

	test.AssertEqual(t, rewardUnitID, reward.UnitID, "TestWelcomeRewardConstant - UnitID")
	test.AssertEqual(t, deviceID, reward.DeviceID, "TestWelcomeRewardConstant - DeviceID")
	test.AssertEqual(t, model.StatusPending, reward.Status, "TestWelcomeRewardConstant - Status")
	test.AssertEqual(t, rewardAmount, reward.Amount, "TestWelcomeRewardConstant - Amount")
}

func TestWelcomeRewardRetention(t *testing.T) {
	udt := generateUniqueUDT()
	ifa := uuid.NewV4().String()

	rewardUnitID := getAppV1DefaultUnit(tests.TestAppID2)
	wrc := getDefaultWelcomeRewardConfig(rewardUnitID)
	rewardAmount := wrc.Amount
	wrc.RetentionDays = 3

	buzzscreen.Service.DB.Save(&wrc)
	defer buzzscreen.Service.DB.Delete(&wrc)

	threeDaysBefore := time.Now().Add(-time.Hour*24*3 - time.Hour)

	params := buildV1BaseTestRequest()

	deviceResponse, _ := execTestDevice(t, tests.TestAppID2, udt, ifa)
	deviceID := int64(deviceResponse.Result["device_id"].(float64))

	device := dbdevice.Device{
		ID: deviceID,
	}
	err := buzzscreen.Service.DB.Where(&device).First(&device).Error
	if err != nil {
		t.Fatalf("TestWelcomeRewardRetention() - with err - %s", err)
	}

	params.Set("device_id", strconv.FormatInt(deviceID, 10))
	params.Set("app_id", strconv.FormatInt(tests.TestAppID2, 10))
	err = tests.SetDeviceUnitRegisteredSeconds(deviceID, rewardUnitID, time.Now().Unix())
	if err != nil {
		t.Fatalf("TestWelcomeRewardRetention() - while setting device URS with err - %s", err)
	}
	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(rewardUnitID), "Cache deletion failed!")
	checkWelcomeRewardGiven(t, deviceID, rewardUnitID, false) // Registered seconds 가 얼마 안되서 실패

	err = tests.SetDeviceUnitRegisteredSeconds(deviceID, rewardUnitID, threeDaysBefore.Unix())
	if err != nil {
		t.Fatalf("TestWelcomeRewardRetention() - while setting device URS with err - %s", err)
	}
	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(rewardUnitID), "Cache deletion failed!")
	checkWelcomeRewardGiven(t, deviceID, rewardUnitID, false) // Registered seconds < WRC.StartTime

	fiveDaysBefore := time.Now().Add(-time.Hour * 24 * 5)
	wrc.StartTime = &fiveDaysBefore
	buzzscreen.Service.DB.Save(&wrc)
	err = tests.SetDeviceUnitRegisteredSeconds(deviceID, rewardUnitID, time.Now().Add(-time.Hour*73).Unix())
	if err != nil {
		t.Fatalf("TestWelcomeRewardRetention() - while setting device URS with err - %s", err)
	}
	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(rewardUnitID), "Cache deletion failed!")
	reward := checkWelcomeRewardGiven(t, deviceID, rewardUnitID, true)

	test.AssertEqual(t, rewardUnitID, reward.UnitID, "TestWelcomeRewardRetention - AppID")
	test.AssertEqual(t, deviceID, reward.DeviceID, "TestWelcomeRewardRetention - DeviceID")
	test.AssertEqual(t, model.StatusPending, reward.Status, "TestWelcomeRewardRetention - Status")
	test.AssertEqual(t, rewardAmount, reward.Amount, "TestWelcomeRewardRetention - Amount")
}

func TestWelcomeRewardCountry(t *testing.T) {
	rewardUnitID := getAppV1DefaultUnit(tests.TestAppID1)
	wrcUs := getDefaultWelcomeRewardConfig(rewardUnitID)
	wrcAll := getDefaultWelcomeRewardConfig(rewardUnitID)
	couUs, couKr := "US", "KR"
	wrcUs.Country = &couUs
	wrcUs.Amount = 500

	params := buildV1BaseTestRequest()

	buzzscreen.Service.DB.Save(&wrcUs)
	defer buzzscreen.Service.DB.Delete(&wrcUs)

	buzzscreen.Service.DB.Save(&wrcAll)
	defer buzzscreen.Service.DB.Delete(&wrcAll)

	deviceResponse, _ := execTestDevice(t, tests.TestAppID1, generateUniqueUDT(), uuid.NewV4().String())
	deviceID := int64(deviceResponse.Result["device_id"].(float64))
	device := dbdevice.Device{
		ID: deviceID,
	}
	err := buzzscreen.Service.DB.Where(&device).First(&device).Error
	if err != nil {
		t.Fatalf("TestWelcomeRewardConstant() - with err - %s", err)
	}

	params.Set("device_id", strconv.FormatInt(deviceID, 10))
	params.Set("app_id", strconv.FormatInt(tests.TestAppID1, 10))
	params.Set("country", couUs)
	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(rewardUnitID), "Cache deletion failed!")

	reward := checkWelcomeRewardGiven(t, deviceID, rewardUnitID, true)

	test.AssertEqual(t, rewardUnitID, reward.UnitID, "TestWelcomeRewardCountry - AppID(US)")
	test.AssertEqual(t, deviceID, reward.DeviceID, "TestWelcomeRewardCountry - DeviceID(US)")
	test.AssertEqual(t, model.StatusPending, reward.Status, "TestWelcomeRewardCountry - Status(US)")
	test.AssertEqual(t, wrcUs.Amount, reward.Amount, "TestWelcomeRewardCountry - Amount(US)")

	deviceResponce, _ := execTestDevice(t, tests.TestAppID1, generateUniqueUDT(), uuid.NewV4().String())
	deviceID = int64(deviceResponce.Result["device_id"].(float64))
	device = dbdevice.Device{
		ID: deviceID,
	}
	err = buzzscreen.Service.DB.Where(&device).First(&device).Error
	if err != nil {
		t.Fatalf("TestWelcomeRewardConstant() - with err - %s", err)
	}

	params.Set("device_id", strconv.FormatInt(deviceID, 10))
	params.Set("app_id", strconv.FormatInt(tests.TestAppID1, 10))
	params.Set("country", couKr)
	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(rewardUnitID), "Cache deletion failed!")

	reward = checkWelcomeRewardGiven(t, deviceID, rewardUnitID, true)

	test.AssertEqual(t, rewardUnitID, reward.UnitID, "TestWelcomeRewardCountry - AppID")
	test.AssertEqual(t, deviceID, reward.DeviceID, "TestWelcomeRewardCountry - DeviceID")
	test.AssertEqual(t, model.StatusPending, reward.Status, "TestWelcomeRewardCountry - Status")
	test.AssertEqual(t, wrcAll.Amount, reward.Amount, "TestWelcomeRewardCountry - Amount")
}

func TestWelcomeRewardMultipleUnit(t *testing.T) {
	appID := tests.KoreaUnit.AppID
	// bellow two units are above app's units
	unitID1 := tests.HsKrFeedUnitID
	unitID2 := tests.KoreaUnit.ID
	wrc1 := getDefaultWelcomeRewardConfig(unitID1)
	wrc2 := getDefaultWelcomeRewardConfig(unitID2)
	wrc2.Amount = 500

	params := buildV1BaseTestRequest()

	buzzscreen.Service.DB.Save(&wrc1)
	defer buzzscreen.Service.DB.Delete(&wrc1)
	buzzscreen.Service.DB.Save(&wrc2)
	defer buzzscreen.Service.DB.Delete(&wrc2)

	deviceResponse, _ := execTestDevice(t, appID, generateUniqueUDT(), uuid.NewV4().String())
	deviceID := int64(deviceResponse.Result["device_id"].(float64))
	device := dbdevice.Device{
		ID: deviceID,
	}
	err := buzzscreen.Service.DB.Where(&device).First(&device).Error
	if err != nil {
		t.Fatalf("TestWelcomeRewardConstant() - with err - %s", err)
	}

	params.Set("device_id", strconv.FormatInt(deviceID, 10))
	params.Set("unit_id", strconv.FormatInt(unitID1, 10))
	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID1), "Cache deletion failed!")

	time.Sleep(time.Second)

	reward := checkWelcomeRewardGiven(t, deviceID, unitID1, true)

	test.AssertEqual(t, unitID1, reward.UnitID, "TestWelcomeRewardCountry - AppID(US)")
	test.AssertEqual(t, deviceID, reward.DeviceID, "TestWelcomeRewardCountry - DeviceID(US)")
	test.AssertEqual(t, model.StatusPending, reward.Status, "TestWelcomeRewardCountry - Status(US)")
	test.AssertEqual(t, wrc1.Amount, reward.Amount, "TestWelcomeRewardCountry - Amount(US)")

	// get second welcome reward with the same device
	params.Set("device_id", strconv.FormatInt(deviceID, 10))
	params.Set("unit_id", strconv.FormatInt(unitID2, 10))
	sendAllocReqWithParams(t, params)
	require.Equal(t, true, DeleteWelcomeRewardCache(unitID2), "Cache deletion failed!")

	time.Sleep(time.Second)

	reward = checkWelcomeRewardGiven(t, deviceID, unitID2, true)

	test.AssertEqual(t, unitID2, reward.UnitID, "TestWelcomeRewardCountry - AppID")
	test.AssertEqual(t, deviceID, reward.DeviceID, "TestWelcomeRewardCountry - DeviceID")
	test.AssertEqual(t, model.StatusPending, reward.Status, "TestWelcomeRewardCountry - Status")
	test.AssertEqual(t, wrc2.Amount, reward.Amount, "TestWelcomeRewardCountry - Amount")
}

func sendAllocReqWithParams(t *testing.T, params *url.Values) {
	var allocRes TestAllocV1Response

	statusCode, err := (&network.Request{
		Method:  "POST",
		Params:  params,
		URL:     ts.URL + "/api/allocation/",
		Timeout: time.Minute * 10,
	}).GetResponse(&allocRes)

	if err != nil || statusCode != 200 {
		t.Fatal(statusCode, err, allocRes)
	}
}

func getAppV1DefaultUnit(appID int64) int64 {
	unit := dbapp.Unit{
		AppID: appID, UnitType: dbapp.UnitTypeLockscreen,
	}
	buzzscreen.Service.DB.Where(&unit).First(&unit)
	return unit.ID
}
