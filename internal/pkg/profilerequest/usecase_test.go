package profilerequest_test

import (
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/profilerequest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

func (ts *UseCaseTestSuite) Test_PopulateProfile() {
	account := profilerequest.Account{
		AppID:     12345,
		IFA:       "sample_ifa",
		AccountID: 54321,
		CookieID:  "sample_cookie_id",
		AppUserID: "sample_app_user_id",
	}

	ts.repo.On("PopulateProfile", mock.AnythingOfType("profilerequest.Account")).Return(nil).Once()

	err := ts.useCase.PopulateProfile(account)
	ts.NoError(err)
}

type UseCaseTestSuite struct {
	suite.Suite
	repo    *mockRepo
	useCase profilerequest.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.useCase = profilerequest.NewUseCase(ts.repo)
}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) PopulateProfile(account profilerequest.Account) error {
	ret := r.Called(account)
	return ret.Error(0)
}
