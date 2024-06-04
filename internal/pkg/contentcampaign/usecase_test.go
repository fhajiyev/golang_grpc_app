package contentcampaign_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/stretchr/testify/suite"
)

func (ts *UseCaseTestSuite) Test_GetContentCampaignByID() {
	mockContentCampaign := &contentcampaign.ContentCampaign{
		ID: rand.Int63n(1000),
	}
	ts.repo.On("GetContentCampaignByID", mock.Anything).Return(mockContentCampaign, nil).Once()

	result, err := ts.useCase.GetContentCampaignByID(mockContentCampaign.ID)

	ts.NoError(err)
	ts.Equal(mockContentCampaign.ID, result.ID)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_ContentCamPaignExpired() {
	mockContentCampaign := &contentcampaign.ContentCampaign{
		ID:      rand.Int63n(1000),
		EndDate: time.Now().AddDate(0, 0, -10),
	}

	result := ts.useCase.IsContentCampaignExpired(mockContentCampaign)

	ts.Equal(true, result)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_ContentCamPaignAlived() {
	mockContentCampaign := &contentcampaign.ContentCampaign{
		ID:      rand.Int63n(1000),
		EndDate: time.Now().AddDate(0, 0, 1),
	}

	result := ts.useCase.IsContentCampaignExpired(mockContentCampaign)

	ts.Equal(false, result)
	ts.repo.AssertExpectations(ts.T())
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

type UseCaseTestSuite struct {
	suite.Suite
	repo    *mockRepo
	useCase contentcampaign.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.useCase = contentcampaign.NewUseCase(ts.repo)
}

var _ contentcampaign.Repository = &mockRepo{}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) GetContentCampaignByID(contentCampaignID int64) (*contentcampaign.ContentCampaign, error) {
	ret := r.Called(contentCampaignID)
	return ret.Get(0).(*contentcampaign.ContentCampaign), nil
}

func (r *mockRepo) IncreaseImpression(campaignID int64, unitID int64) error {
	ret := r.Called(campaignID, unitID)
	return ret.Error(0)
}

func (r *mockRepo) IncreaseClick(campaignID int64, unitID int64) error {
	ret := r.Called(campaignID, unitID)
	return ret.Error(0)
}
