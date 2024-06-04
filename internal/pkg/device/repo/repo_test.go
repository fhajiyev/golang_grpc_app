package repo_test

import (
	"math/rand"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device/repo"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func (ts *DeviceRepoTestSuite) Test_GetByID() {
	var dbd dbdevice.Device
	ts.NoError(faker.FakeData(&dbd))
	dbd.ID = rand.Int63n(999999) + 1
	ts.ds.On("GetByID", dbd.ID).Return(&dbd, nil).Once()

	d, err := ts.repo.GetByID(dbd.ID)
	ts.NoError(err)
	ts.equalDevice(*d, dbd)
}

func (ts *DeviceRepoTestSuite) TestGetByAppIDAndIFA() {
	appID := rand.Int63n(999999) + 1
	var ifa string
	var dbd dbdevice.Device
	ts.NoError(faker.FakeData(&ifa))
	ts.NoError(faker.FakeData(&dbd))
	ts.ds.On("GetByAppIDAndIFA", appID, ifa).Return(&dbd, nil).Once()

	d, err := ts.repo.GetByAppIDAndIFA(appID, ifa)
	ts.NoError(err)
	ts.equalDevice(*d, dbd)
}

func (ts *DeviceRepoTestSuite) TestUpsertDeviceByDS() {
	var dbd dbdevice.Device
	var d device.Device
	ts.NoError(faker.FakeData(&d))
	ts.NoError(faker.FakeData(&dbd))
	ts.ds.On("UpsertDevice", mock.Anything).Return(&dbd, nil).Once()

	du, err := ts.repo.UpsertDevice(d)
	ts.NoError(err)
	ts.equalDevice(*du, dbd)
}

func (ts *DeviceRepoTestSuite) equalDevice(d device.Device, dbd dbdevice.Device) bool {
	return ts.Equal(dbd.ID, d.ID) &&
		ts.Equal(d.AppID, dbd.AppID) &&
		ts.Equal(d.UnitDeviceToken, dbd.UnitDeviceToken) &&
		ts.Equal(d.IFA, dbd.IFA) &&
		ts.Equal(d.Address, dbd.Address) &&
		ts.Equal(d.Birthday, dbd.Birthday) &&
		ts.Equal(d.Carrier, dbd.Carrier) &&
		ts.Equal(d.DeviceName, dbd.DeviceName) &&
		ts.Equal(d.Resolution, dbd.Resolution) &&
		ts.Equal(d.YearOfBirth, dbd.YearOfBirth) &&
		ts.Equal(d.SDKVersion, dbd.SDKVersion) &&
		ts.Equal(d.Sex, dbd.Sex) &&
		ts.Equal(d.Packages, dbd.Packages) &&
		ts.Equal(d.PackageName, dbd.PackageName) &&
		ts.Equal(d.SignupIP, dbd.SignupIP) &&
		ts.Equal(d.SerialNumber, dbd.SerialNumber) &&
		ts.Equal(d.CreatedAt, dbd.CreatedAt) &&
		ts.Equal(d.UpdatedAt, dbd.UpdatedAt)
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(DeviceRepoTestSuite))
}

type DeviceRepoTestSuite struct {
	suite.Suite
	repo device.Repository
	ds   *mockDeviceDS
}

func (ts *DeviceRepoTestSuite) SetupSuite() {
	ts.ds = &mockDeviceDS{}
	ts.repo = repo.New(ts.ds)
}

type mockDeviceDS struct {
	mock.Mock
}

func (r *mockDeviceDS) GetByID(deviceID int64) (*dbdevice.Device, error) {
	ret := r.Called(deviceID)
	return ret.Get(0).(*dbdevice.Device), ret.Error(1)
}

func (r *mockDeviceDS) GetByAppIDAndIFA(appID int64, ifa string) (*dbdevice.Device, error) {
	ret := r.Called(appID, ifa)
	return ret.Get(0).(*dbdevice.Device), ret.Error(1)
}

func (r *mockDeviceDS) GetByAppIDAndPubUserID(appID int64, pubUserID string) (*dbdevice.Device, error) {
	ret := r.Called(appID, pubUserID)
	return ret.Get(0).(*dbdevice.Device), ret.Error(1)
}

func (r *mockDeviceDS) UpsertDevice(deviceReq dbdevice.Device) (*dbdevice.Device, error) {
	ret := r.Called(deviceReq)
	return ret.Get(0).(*dbdevice.Device), ret.Error(1)
}
