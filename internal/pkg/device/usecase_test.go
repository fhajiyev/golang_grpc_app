package device_test

import (
	"math/rand"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.profileRepo = new(mockProfileRepo)
	ts.activityRepo = new(mockActivityRepo)
	ts.useCase = device.NewUseCase(ts.repo, ts.profileRepo, ts.activityRepo)
}

func (ts *UseCaseTestSuite) TearDownTest() {
	ts.repo.AssertExpectations(ts.T())
	ts.profileRepo.AssertExpectations(ts.T())
	ts.activityRepo.AssertExpectations(ts.T())
}

type UseCaseTestSuite struct {
	suite.Suite
	useCase      device.UseCase
	repo         *mockRepo
	profileRepo  *mockProfileRepo
	activityRepo *mockActivityRepo
}

func (ts *UseCaseTestSuite) Test_GetByID_InvalidArgs() {
	ts.Run("success", func() {
		var device *device.Device
		faker.FakeData(&device)
		device.ID = int64(rand.Int63() + 1)

		ts.repo.On("GetByID", device.ID).Return(device, nil).Once()

		res, err := ts.useCase.GetByID(device.ID)

		ts.NoError(err)
		ts.NotNil(res)
		ts.Equal(device.ID, res.ID)
	})
	ts.Run("invalid args", func() {
		deviceID := int64(0)
		res, err := ts.useCase.GetByID(deviceID)
		_, ok := err.(device.InvalidArgumentError)

		ts.Error(err)
		ts.True(ok)
		ts.Nil(res)
	})
}

func (ts *UseCaseTestSuite) TestValidateUnitDeviceToken() {
	ts.Run("success", func() {
		udt := "TEST_UNIT_DEVICE_TOKEN"
		ok, err := ts.useCase.ValidateUnitDeviceToken(udt)
		ts.NoError(err)
		ts.True(ok)
	})
	ts.Run("error starts with nil character", func() {
		udt := "\u0000$r"
		ok, err := ts.useCase.ValidateUnitDeviceToken(udt)
		ts.Error(err)
		ts.False(ok)
	})
	ts.Run("empty string", func() {
		udt := ""
		ok, err := ts.useCase.ValidateUnitDeviceToken(udt)
		ts.Error(err)
		ts.False(ok)
	})
}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) GetByID(deviceID int64) (*device.Device, error) {
	ret := r.Called(deviceID)
	return ret.Get(0).(*device.Device), ret.Error(1)
}

func (r *mockRepo) GetByAppIDAndIFA(appID int64, ifa string) (*device.Device, error) {
	ret := r.Called(appID, ifa)
	return ret.Get(0).(*device.Device), ret.Error(1)
}

func (r *mockRepo) GetByAppIDAndPubUserID(appID int64, pubUserID string) (*device.Device, error) {
	ret := r.Called(appID, pubUserID)
	return ret.Get(0).(*device.Device), ret.Error(1)
}

func (r *mockRepo) UpsertDevice(d device.Device) (*device.Device, error) {
	ret := r.Called(d)
	return ret.Get(0).(*device.Device), ret.Error(1)
}

type mockProfileRepo struct {
	mock.Mock
}

func (r *mockProfileRepo) GetByID(deviceID int64) (*device.Profile, error) {
	ret := r.Called(deviceID)
	return ret.Get(0).(*device.Profile), ret.Error(1)
}

func (r *mockProfileRepo) Save(dp device.Profile) error {
	ret := r.Called(dp)
	return ret.Error(0)
}

func (r *mockProfileRepo) SavePackage(dp device.Profile) error {
	ret := r.Called(dp)
	return ret.Error(0)
}

func (r *mockProfileRepo) SaveUnitRegisteredSeconds(dp device.Profile) error {
	ret := r.Called(dp)
	return ret.Error(0)
}

func (r *mockProfileRepo) Delete(dp device.Profile) error {
	ret := r.Called(dp)
	return ret.Error(0)
}

type mockActivityRepo struct {
	mock.Mock
}

func (r *mockActivityRepo) GetByID(deviceID int64) (*device.Activity, error) {
	ret := r.Called(deviceID)
	return ret.Get(0).(*device.Activity), ret.Error(1)
}

func (r *mockActivityRepo) Save(deviceID int64, campaignID int64, activityType device.ActivityType) error {
	ret := r.Called(deviceID)
	return ret.Error(0)
}
