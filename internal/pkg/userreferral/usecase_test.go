package userreferral_test

import (
	"math/rand"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/userreferral"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

type UseCaseTestSuite struct {
	suite.Suite
	userRepo *mockUserRepo
	useCase  userreferral.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.userRepo = new(mockUserRepo)
	ts.useCase = userreferral.NewUseCase(ts.userRepo)
}

func (ts *UseCaseTestSuite) TearDownTest() {
	ts.userRepo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_GetUserByCode() {
	user := ts.getTestUser()

	ts.Run("success", func() {
		ts.userRepo.On("GetUserByCode", user.Code).Return(user, nil).Once()

		res, err := ts.useCase.GetUserByCode(user.Code)

		ts.NoError(err)
		ts.equalUser(user, res)
	})

	ts.Run("invalid argument", func() {
		res, err := ts.useCase.GetUserByCode("")
		_, ok := err.(userreferral.InvalidArgumentError)

		ts.Error(err)
		ts.True(ok)
		ts.Nil(res)
	})
}

func (ts *UseCaseTestSuite) Test_GetOrCreateUserByDevice() {
	user := ts.getTestUser()
	appID := int64(1 + rand.Intn(65536))
	udt := "abcd"
	verifyURL := "https://verify.buzzvil.com"

	ts.Run("get user success", func() {
		ts.userRepo.On("GetUserByDevice", user.DeviceID).Return(user, nil).Once()

		res, err := ts.useCase.GetOrCreateUserByDevice(user.DeviceID, appID, udt, verifyURL)

		ts.NoError(err)
		ts.equalUser(user, res)
	})

	ts.Run("invalid argument", func() {
		res, err := ts.useCase.GetOrCreateUserByDevice(int64(0), appID, udt, verifyURL)
		_, ok := err.(userreferral.InvalidArgumentError)

		ts.Error(err)
		ts.True(ok)
		ts.Nil(res)
	})

	ts.Run("create user verify", func() {
		ts.Run("success", func() {
			ts.userRepo.On("GetUserByDevice", user.DeviceID).Return(nil, userreferral.NotFoundError{}).Once()
			ts.userRepo.On("IsVerifiedDevice", verifyURL, udt).Return(nil).Once()
			ts.userRepo.On("CreateUser", user.DeviceID, mock.Anything, true).Return(user, nil).Once()

			res, err := ts.useCase.GetOrCreateUserByDevice(user.DeviceID, appID, udt, verifyURL)

			ts.NoError(err)
			ts.equalUser(user, res)
		})

		ts.Run("fail", func() {
			ts.userRepo.On("GetUserByDevice", user.DeviceID).Return(nil, userreferral.NotFoundError{}).Once()
			ts.userRepo.On("IsVerifiedDevice", verifyURL, udt).Return(userreferral.APICallError{}).Once()

			res, err := ts.useCase.GetOrCreateUserByDevice(user.DeviceID, appID, udt, verifyURL)
			_, ok := err.(userreferral.APICallError)

			ts.Nil(res)
			ts.Error(err)
			ts.True(ok)
		})
	})
}

func (ts *UseCaseTestSuite) Test_CreateReferral_Success() {
	referee := ts.getTestUser()
	referrer := ts.getTestUser()
	createIngr := ts.getTestCreateReferralIngredients(referee, referrer)
	giveIngr := ts.getTestGiveReferralRewardRequestIngredients(referee, referrer, createIngr)
	referralCnt := 3

	ts.makeUserVadationPassed(referee, referrer, createIngr)

	ts.Run("referee exists", func() {
		ts.userRepo.On("GetUserByDevice", referee.DeviceID).Return(referee, nil).Once()
		ts.userRepo.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()
		ts.userRepo.On("GetReferralCountByUser", referrer.ID).Return(referralCnt, nil).Once()
		ts.userRepo.On("GiveReferralReward", *giveIngr).Return(nil).Once()
		ts.userRepo.On("UpdateUserReferrerID", referee.ID, referrer.ID).Return(nil).Once()

		res, err := ts.useCase.CreateReferral(*createIngr)
		ts.True(res)
		ts.NoError(err)
	})

	ts.Run("referee not exists", func() {

		ts.userRepo.On("GetUserByDevice", referee.DeviceID).Return(nil, userreferral.NotFoundError{}).Once()
		ts.userRepo.On("IsVerifiedDevice", createIngr.VerifyURL, createIngr.UnitDeviceToken).Return(nil).Once()
		// referee's code is generated randomly
		ts.userRepo.On("CreateUser", referee.DeviceID, mock.Anything, true).Return(referee, nil).Once()
		ts.userRepo.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()
		ts.userRepo.On("GetReferralCountByUser", referrer.ID).Return(referralCnt, nil).Once()
		ts.userRepo.On("GiveReferralReward", *giveIngr).Return(nil).Once()
		ts.userRepo.On("UpdateUserReferrerID", referee.ID, referrer.ID).Return(nil).Once()

		res, err := ts.useCase.CreateReferral(*createIngr)
		ts.True(res)
		ts.NoError(err)
	})
}

func (ts *UseCaseTestSuite) Test_CreateReferral_InvalidArgumentError() {
	referee := ts.getTestUser()
	referrer := ts.getTestUser()
	ts.Run("invalid device ID", func() {
		createIngr := ts.getTestCreateReferralIngredients(referee, referrer)
		createIngr.DeviceID = int64(0)

		res, err := ts.useCase.CreateReferral(*createIngr)
		_, ok := err.(userreferral.InvalidArgumentError)

		ts.Error(err)
		ts.True(ok)
		ts.False(res)
	})
	ts.Run("invalid code", func() {
		createIngr := ts.getTestCreateReferralIngredients(referee, referrer)
		createIngr.Code = ""

		res, err := ts.useCase.CreateReferral(*createIngr)
		_, ok := err.(userreferral.InvalidArgumentError)

		ts.Error(err)
		ts.True(ok)
		ts.False(res)
	})
}

func (ts *UseCaseTestSuite) Test_CreateReferral_RefereeNotExistsDeviceVerifyError() {
	referee := ts.getTestUser()
	referrer := ts.getTestUser()
	createIngr := ts.getTestCreateReferralIngredients(referee, referrer)

	ts.userRepo.On("GetUserByDevice", referee.DeviceID).Return(nil, userreferral.NotFoundError{}).Once()
	ts.userRepo.On("IsVerifiedDevice", createIngr.VerifyURL, createIngr.UnitDeviceToken).Return(userreferral.APICallError{}).Once()

	res, err := ts.useCase.CreateReferral(*createIngr)
	_, ok := err.(userreferral.APICallError)

	ts.False(res)
	ts.Error(err)
	ts.True(ok)
}

func (ts *UseCaseTestSuite) Test_CreateReferral_InvalidReferrerCodeError() {
	referee := ts.getTestUser()
	referrer := ts.getTestUser()
	createIngr := ts.getTestCreateReferralIngredients(referee, referrer)

	ts.userRepo.On("GetUserByDevice", referee.DeviceID).Return(referee, nil).Once()
	ts.userRepo.On("GetUserByCode", referrer.Code).Return(nil, userreferral.NotFoundError{}).Once()

	res, err := ts.useCase.CreateReferral(*createIngr)
	_, ok := err.(userreferral.InvalidArgumentError)

	ts.False(res)
	ts.Error(err)
	ts.True(ok)
}

func (ts *UseCaseTestSuite) Test_CreateReferral_ReferrerAlreadySet() {
	referee := ts.getTestUser()
	referrer := ts.getTestUser()

	// Referrer is already set
	referee.ReferrerID = int64(1)
	referrer.Code += "1"

	createIngr := ts.getTestCreateReferralIngredients(referee, referrer)

	ts.userRepo.On("GetUserByDevice", referee.DeviceID).Return(referee, nil).Once()
	ts.userRepo.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()

	res, err := ts.useCase.CreateReferral(*createIngr)

	ts.True(res)
	ts.NoError(err)
}

func (ts *UseCaseTestSuite) Test_CreateReferral_SelfReferralError() {
	referee := ts.getTestUser()
	referrer := ts.getTestUser()

	// referee refers itself
	referee.ReferrerID = 0
	referee.ID = referrer.ID

	createIngr := ts.getTestCreateReferralIngredients(referee, referrer)

	ts.userRepo.On("GetUserByDevice", referee.DeviceID).Return(referee, nil).Once()
	ts.userRepo.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()

	res, err := ts.useCase.CreateReferral(*createIngr)
	_, ok := err.(userreferral.UserValidationError)

	ts.False(res)
	ts.Error(err)
	ts.True(ok)
}

func (ts *UseCaseTestSuite) Test_CreateReferral_RefereeNotVerified() {
	referee := ts.getTestUser()
	referrer := ts.getTestUser()
	createIngr := ts.getTestCreateReferralIngredients(referee, referrer)

	ts.makeUserVadationPassed(referee, referrer, createIngr)
	referee.IsVerified = false

	ts.userRepo.On("GetUserByDevice", referee.DeviceID).Return(referee, nil).Once()
	ts.userRepo.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()

	res, err := ts.useCase.CreateReferral(*createIngr)
	_, ok := err.(userreferral.UserValidationError)

	ts.False(res)
	ts.Error(err)
	ts.True(ok)
}

func (ts *UseCaseTestSuite) Test_CreateReferral_GiveReferralRewardError() {
	referee := ts.getTestUser()
	referrer := ts.getTestUser()
	createIngr := ts.getTestCreateReferralIngredients(referee, referrer)
	giveIngr := ts.getTestGiveReferralRewardRequestIngredients(referee, referrer, createIngr)
	referralCnt := 3

	ts.makeUserVadationPassed(referee, referrer, createIngr)

	ts.userRepo.On("GetUserByDevice", referee.DeviceID).Return(referee, nil).Once()
	ts.userRepo.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()
	ts.userRepo.On("GetReferralCountByUser", referrer.ID).Return(referralCnt, nil).Once()
	ts.userRepo.On("GiveReferralReward", *giveIngr).Return(userreferral.APICallError{}).Once()

	res, err := ts.useCase.CreateReferral(*createIngr)
	_, ok := err.(userreferral.APICallError)

	ts.False(res)
	ts.Error(err)
	ts.True(ok)
}

func (ts *UseCaseTestSuite) makeUserVadationPassed(referee *userreferral.DeviceUser, referrer *userreferral.DeviceUser, ingr *userreferral.CreateReferralIngredients) {
	referee.ReferrerID = 0
	referrer.Code += "1"
	referrer.ID = referee.ID + 1
	referrer.ReferrerID = referee.ID + 2
	ingr.VerifyURL = "https://verify.buzzvil.com"
	ingr.Code = referrer.Code
	referee.IsVerified = true
}

func (ts *UseCaseTestSuite) equalUser(user1 *userreferral.DeviceUser, user2 *userreferral.DeviceUser) {
	ts.True(user1.ID == user2.ID)
	ts.True(user1.DeviceID == user2.DeviceID)
	ts.True(user1.Code == user2.Code)
	ts.True(user1.ReferrerID == user2.ReferrerID)
	ts.True(user1.IsVerified == user2.IsVerified)
}

func (ts *UseCaseTestSuite) getTestUser() *userreferral.DeviceUser {
	var user userreferral.DeviceUser
	ts.NoError(faker.FakeData(&user))
	user.ID++
	user.DeviceID++
	return &user
}

func (ts *UseCaseTestSuite) getTestCreateReferralIngredients(referee *userreferral.DeviceUser, referrer *userreferral.DeviceUser) *userreferral.CreateReferralIngredients {
	var ingr userreferral.CreateReferralIngredients
	ts.NoError(faker.FakeData(&ingr))

	ingr.DeviceID = referee.DeviceID
	ingr.AppID++
	ingr.Code = referrer.Code
	ingr.MaxReferral = 65536
	return &ingr
}

func (ts *UseCaseTestSuite) getTestGiveReferralRewardRequestIngredients(referee *userreferral.DeviceUser, referrer *userreferral.DeviceUser, ingr *userreferral.CreateReferralIngredients) *userreferral.GiveReferralRewardRequestIngredients {
	return &userreferral.GiveReferralRewardRequestIngredients{
		RefereeDeviceID:  referee.DeviceID,
		RefereeReward:    ingr.RewardAmount,
		ReferrerDeviceID: referrer.DeviceID,
		ReferrerReward:   ingr.RewardAmount,
		JWT:              ingr.JWT,
		TitleForReferral: ingr.TitleForReferral,
	}
}

type mockUserRepo struct {
	mock.Mock
}

func (r *mockUserRepo) CreateUser(deviceID int64, code string, isVerified bool) (*userreferral.DeviceUser, error) {
	ret := r.Called(deviceID, code, isVerified)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*userreferral.DeviceUser), ret.Error(1)
}

func (r *mockUserRepo) UpdateUserReferrerID(userID int64, referrerID int64) error {
	ret := r.Called(userID, referrerID)
	return ret.Error(0)
}

func (r *mockUserRepo) GetUserByDevice(deviceID int64) (*userreferral.DeviceUser, error) {
	ret := r.Called(deviceID)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*userreferral.DeviceUser), ret.Error(1)
}

func (r *mockUserRepo) GetUserByCode(code string) (*userreferral.DeviceUser, error) {
	ret := r.Called(code)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*userreferral.DeviceUser), ret.Error(1)
}

func (r *mockUserRepo) GetReferralCountByUser(referrerID int64) (int, error) {
	ret := r.Called(referrerID)
	return ret.Int(0), ret.Error(1)
}

func (r *mockUserRepo) IsVerifiedDevice(verifyURL string, udt string) error {
	ret := r.Called(verifyURL, udt)
	return ret.Error(0)
}

func (r *mockUserRepo) GiveReferralReward(ingr userreferral.GiveReferralRewardRequestIngredients) error {
	ret := r.Called(ingr)
	return ret.Error(0)
}
