package notiplus_test

import (
	"testing"

	"github.com/bxcodec/faker"

	"github.com/stretchr/testify/mock"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/notiplus"
	"github.com/stretchr/testify/suite"
)

func (ts *UseCaseTestSuite) Test_GetConfigsByUnitID() {
	unitID := int64(100000043)
	mockConfigs := make([]notiplus.Config, 0)

	var abc *notiplus.Config
	faker.FakeData(&abc)

	for i := 0; i < 2; i++ {
		var mockConfig notiplus.Config
		err := faker.FakeData(&mockConfig)
		mockConfig.UnitID = unitID
		ts.NoError(err)
		mockConfigs = append(mockConfigs, mockConfig)
	}

	ts.repo.On("GetConfigsByUnitID", mock.Anything).Return(mockConfigs, nil).Once()

	configs, _ := ts.useCase.GetConfigsByUnitID(unitID)
	ts.repo.AssertExpectations(ts.T())
	for i := 0; i < 2; i++ {
		ts.Equal(configs[i].UnitID, mockConfigs[i].UnitID)
		ts.Equal(configs[i].ID, mockConfigs[i].ID)
	}
}

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

type UseCaseTestSuite struct {
	suite.Suite
	repo    *mockRepo
	useCase notiplus.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.useCase = notiplus.NewUseCase(ts.repo)
}

var _ notiplus.Repository = &mockRepo{}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) GetConfigsByUnitID(unitID int64) ([]notiplus.Config, error) {
	ret := r.Called(unitID)
	return ret.Get(0).([]notiplus.Config), ret.Error(1)
}
