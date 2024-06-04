package dbdevice_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/Buzzvil/buzzscreen-api/tests"
)

func TestLegacySourceSuite(t *testing.T) {
	suite.Run(t, new(LegacySourceSuite))
}

type LegacySourceSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
	db   *gorm.DB
	ds   dbdevice.DBSource
}

func (ts *LegacySourceSuite) SetupTest() {
	if buzzscreen.Service == nil {
		bsSvc := buzzscreen.New()
		server := core.NewServer(bsSvc)
		configPath := os.Getenv("GOPATH") + "/src/github.com/Buzzvil/buzzscreen-api/config/"
		server.Init(configPath, &env.Config)
		tests.SetupDatabase()
	}

	ts.db = buzzscreen.Service.DB
	ts.ds = dbdevice.NewSource(ts.db)
}

func (ts *LegacySourceSuite) BeforeTest() {
	ts.db.Delete(&dbdevice.Device{})
	ts.db.Delete(&dbdevice.DeviceUpdateHistory{})
}

func getNewTestDevice(udt, ifa string) *dbdevice.Device {
	if ifa == "" {
		ifa = uuid.NewV4().String()
	}

	packageName := "com.honeyscreen"

	return &dbdevice.Device{
		AppID:           tests.HsKrAppID,
		UnitDeviceToken: udt,
		IFA:             ifa,
		Resolution:      "1080x1776",
		DeviceName:      "Nexus 6",
		PackageName:     &packageName,
	}
}

func (ts *LegacySourceSuite) TestHandleChangeDevice() {
	udtUserA := fmt.Sprintf("TestUserA-%d", time.Now().UnixNano())
	deviceUserA := getNewTestDevice(udtUserA, "")

	udtUserB := fmt.Sprintf("TestUserB-%d", time.Now().UnixNano())
	oldDeviceUserB := getNewTestDevice(udtUserB, "")

	// User A 가입
	deviceUserA, err := ts.ds.UpsertDevice(*deviceUserA)
	ts.NoError(err)

	// User B 가입
	oldDeviceUserB, err = ts.ds.UpsertDevice(*oldDeviceUserB)
	ts.NoError(err)

	// 기기B 를 User A가 획득
	req := getNewTestDevice(deviceUserA.UnitDeviceToken, oldDeviceUserB.IFA)
	newDeviceUserA, err := ts.ds.UpsertDevice(*req)
	ts.NoError(err)

	ts.Equal(deviceUserA.ID, newDeviceUserA.ID)
	ts.Equal(newDeviceUserA.UnitDeviceToken, deviceUserA.UnitDeviceToken)
	ts.Equal(newDeviceUserA.IFA, oldDeviceUserB.IFA)
	ts.Equal(newDeviceUserA.CreatedAt.IsZero(), false)
	ts.Equal(newDeviceUserA.UpdatedAt.IsZero(), false)

	var histories []dbdevice.DeviceUpdateHistory
	err = ts.db.Where(&dbdevice.DeviceUpdateHistory{DeviceID: newDeviceUserA.ID}).Find(&histories).Error
	ts.NoError(err)
	ts.Equal(1, len(histories))
	history := histories[0]
	ts.Equal("ifa", history.UpdatedField)
	ts.Equal(deviceUserA.IFA, history.FromValue)
	ts.Equal(oldDeviceUserB.IFA, history.ToValue)

	err = ts.db.Where(&dbdevice.DeviceUpdateHistory{DeviceID: oldDeviceUserB.ID}).First(&histories).Error
	ts.NoError(err)
	ts.Equal(1, len(histories))
	history = histories[0]
	ts.Equal("ifa", history.UpdatedField)
	ts.Equal(oldDeviceUserB.IFA, history.FromValue)
	ts.Equal(fmt.Sprintf("%s_d_%v_%v", oldDeviceUserB.IFA, newDeviceUserA.ID, oldDeviceUserB.ID), history.ToValue)
}

func (ts *LegacySourceSuite) TestDevicePackageName() {
	// UserA 가입
	deviceWithOldPackage := getNewTestDevice(fmt.Sprintf("TestUserA-%d", time.Now().UnixNano()), "")
	deviceWithOldPackage, err := ts.ds.UpsertDevice(*deviceWithOldPackage)
	ts.NoError(err)

	deviceWithNewPackage := getNewTestDevice(deviceWithOldPackage.UnitDeviceToken, deviceWithOldPackage.IFA)

	newPackageName := "com.honeyscreen.v2"
	deviceWithNewPackage.PackageName = &newPackageName
	deviceWithNewPackage, err = ts.ds.UpsertDevice(*deviceWithNewPackage)
	ts.NoError(err)

	var duh dbdevice.DeviceUpdateHistory
	err = ts.db.Where(&dbdevice.DeviceUpdateHistory{DeviceID: deviceWithOldPackage.ID}).First(&duh).Error
	ts.NoError(err)
	ts.Equal(duh.DeviceID, deviceWithOldPackage.ID)
	ts.Equal(duh.FromValue, *deviceWithOldPackage.PackageName)
	ts.Equal(duh.ToValue, *deviceWithNewPackage.PackageName)
	ts.Equal(duh.UpdatedField, "package_name")
}

func (ts *LegacySourceSuite) TestHandleNewDeviceIFA() {
	// UserA 가입
	testDevice := getNewTestDevice(fmt.Sprintf("TestUserA-%d", time.Now().UnixNano()), uuid.NewV1().String())
	deviceUserA1, err := ts.ds.UpsertDevice(*testDevice)
	ts.NoError(err)

	// UserA가 IFA 변경
	testDevice2 := getNewTestDevice(testDevice.UnitDeviceToken, uuid.NewV1().String())
	deviceUserA2, _ := ts.ds.UpsertDevice(*testDevice2)
	ts.NoError(err)

	ts.T().Logf("UserA: %#v\nUserB: %#v", *deviceUserA1, *deviceUserA2)

	var updateHistoryDevice1 dbdevice.DeviceUpdateHistory
	err = ts.db.Where(&dbdevice.DeviceUpdateHistory{DeviceID: deviceUserA1.ID}).First(&updateHistoryDevice1).Error
	ts.NoError(err)

	ts.Equal(updateHistoryDevice1.DeviceID, deviceUserA2.ID, "TestHandleNewDevice()")
	ts.Equal(updateHistoryDevice1.FromValue, deviceUserA1.IFA, "TestHandleNewDevice()")
	ts.Equal(updateHistoryDevice1.ToValue, deviceUserA2.IFA, "TestHandleNewDevice()")
	ts.Equal(updateHistoryDevice1.UpdatedField, "ifa", "TestHandleNewDevice()")
}

func (ts *LegacySourceSuite) TestHandleNewDeviceDeviceToken() {
	db := buzzscreen.Service.DB
	ds := dbdevice.NewSource(db)
	db.Delete(&dbdevice.Device{})
	db.Delete(&dbdevice.DeviceUpdateHistory{})

	testDeviceC := getNewTestDevice(uuid.NewV1().String(), "")
	testDeviceC, err := ds.UpsertDevice(*testDeviceC)
	ts.NoError(err)
	testDeviceD := getNewTestDevice(uuid.NewV1().String(), testDeviceC.IFA)
	testDeviceD, err = ds.UpsertDevice(*testDeviceD)
	ts.NoError(err)

	var updateHistoryDeviceC dbdevice.DeviceUpdateHistory
	err = db.Where(&dbdevice.DeviceUpdateHistory{DeviceID: testDeviceC.ID}).First(&updateHistoryDeviceC).Error
	ts.NoError(err)
	ts.Equal(testDeviceC.ID, testDeviceD.ID, "TestHandleNewDevice()")
	ts.Equal(updateHistoryDeviceC.DeviceID, testDeviceD.ID, "TestHandleNewDevice()")
	ts.Equal(updateHistoryDeviceC.FromValue, testDeviceC.UnitDeviceToken, "TestHandleNewDevice()")
	ts.Equal(updateHistoryDeviceC.ToValue, testDeviceD.UnitDeviceToken, "TestHandleNewDevice()")
	ts.Equal(updateHistoryDeviceC.UpdatedField, "unit_device_token", "TestHandleNewDevice()")
}

func (ts *LegacySourceSuite) TestHandleNewDevice() {

	ifaC, userIDC := uuid.NewV1().String(), uuid.NewV1().String()
	deviceNameC, deviceNameD := "Nexus 5", "Galaxy S8"

	deviceUserC, err := ts.ds.UpsertDevice(dbdevice.Device{
		AppID:           tests.HsKrAppID,
		UnitDeviceToken: userIDC,
		IFA:             ifaC,
		Resolution:      "1080x1776",
		DeviceName:      deviceNameC,
	})
	ts.NoError(err)

	deviceUserD, err := ts.ds.UpsertDevice(dbdevice.Device{
		AppID:           tests.HsKrAppID,
		UnitDeviceToken: userIDC,
		IFA:             ifaC,
		Resolution:      "1080x1776",
		DeviceName:      deviceNameD,
	})
	ts.NoError(err)

	// history가 남으면 안됨
	var updateHistoryDeviceC dbdevice.DeviceUpdateHistory
	err = ts.db.Where(&dbdevice.DeviceUpdateHistory{DeviceID: deviceUserC.ID}).First(&updateHistoryDeviceC).Error
	ts.Error(err)

	var deviceUserE dbdevice.Device
	err = ts.db.Where(&dbdevice.Device{ID: deviceUserD.ID}).First(&deviceUserE).Error
	ts.NoError(err)

	ts.Equal(deviceUserC.ID, deviceUserE.ID, "TestHandleNewDeviceName()")
	ts.Equal(deviceUserE.UnitDeviceToken, userIDC, "TestHandleNewDeviceName()")
	ts.Equal(deviceUserE.IFA, ifaC, "TestHandleNewDeviceName()")
	ts.Equal(deviceUserE.DeviceName, deviceNameD, "TestHandleNewDeviceName()")
	ts.Equal(deviceUserE.CreatedAt.IsZero(), false, "TestHandleNewDeviceName()")
	ts.Equal(deviceUserE.UpdatedAt.IsZero(), false, "TestHandleNewDeviceName()")
}

func (ts *LegacySourceSuite) TestHandleDeviceFieldUpdates() {
	ifa, userID := uuid.NewV1().String(), uuid.NewV1().String()
	birthday, err := time.Parse("2006-01-02", "1986-06-07")

	if err != nil {
		ts.FailNowf("TestHandleDeviceFieldUpdates failed", "err: %s", err)
	}

	device, err := ts.ds.UpsertDevice(dbdevice.Device{
		AppID:           tests.HsKrAppID,
		UnitDeviceToken: userID,
		IFA:             ifa,
		Resolution:      "1080x1776",
		DeviceName:      "Nexus 5",
	})
	ts.NoError(err)

	carrier, sdkVersion, sex := "SK Telecom", 1440, "M"
	deviceUpdated, err := ts.ds.UpsertDevice(dbdevice.Device{
		AppID:           tests.HsKrAppID,
		UnitDeviceToken: userID,
		IFA:             ifa,
		Birthday:        &(birthday),
		Carrier:         &(carrier),
		SDKVersion:      &(sdkVersion),
		Sex:             &(sex),
	})
	ts.NoError(err)

	// history가 남으면 안됨
	var updateHistoryDeviceC dbdevice.DeviceUpdateHistory
	err = ts.db.Where(&dbdevice.DeviceUpdateHistory{DeviceID: device.ID}).First(&updateHistoryDeviceC).Error
	ts.Error(err)

	var deviceDB dbdevice.Device
	err = ts.db.Where(&dbdevice.Device{ID: device.ID}).First(&deviceDB).Error
	ts.NoError(err)

	ts.Equal(deviceDB.UnitDeviceToken, device.UnitDeviceToken, "TestHandleDeviceFieldUpdates()")
	ts.Equal(deviceDB.AppID, device.AppID, "TestHandleDeviceFieldUpdates()")
	ts.Equal(deviceDB.IFA, device.IFA, "TestHandleDeviceFieldUpdates()")
	ts.Equal(deviceDB.Resolution, device.Resolution, "TestHandleDeviceFieldUpdates()")
	ts.Equal(deviceDB.DeviceName, device.DeviceName, "TestHandleDeviceFieldUpdates()")
	ts.Equal(*(deviceDB.SDKVersion), *(deviceUpdated.SDKVersion), "TestHandleDeviceFieldUpdates()")
	ts.Equal(*(deviceDB.Sex), *(deviceUpdated.Sex), "TestHandleDeviceFieldUpdates()")
	ts.Equal(*(deviceDB.Birthday), *(deviceUpdated.Birthday), "TestHandleDeviceFieldUpdates()")
}
