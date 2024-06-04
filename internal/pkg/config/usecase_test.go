package config_test

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/config"
	"github.com/stretchr/testify/suite"
)

func (ts *UseCaseTestSuite) Test_GetConfigs() {
	mockConfigs := &[]config.Config{
		config.Config{
			Key:   "ArcadeEnabled",
			Value: "true",
		},
		config.Config{
			Key:   "RefreshRate",
			Value: "30",
		},
	}
	ts.repo.On("GetConfigs", mock.Anything).Return(mockConfigs, nil).Once()
	var configRequest config.RequestIngredients
	result := ts.useCase.GetConfigs(configRequest)

	ts.Equal((*mockConfigs)[0].Key, (*result)[0].Key)
	ts.Equal((*mockConfigs)[0].Value, (*result)[0].Value)
	ts.repo.AssertExpectations(ts.T())
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
	useCase config.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.useCase = config.NewUseCase(ts.repo)
}

var _ config.Repository = &mockRepo{}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) GetConfigs(configRequest config.RequestIngredients) *[]config.Config {
	ret := r.Called(configRequest)
	return ret.Get(0).(*[]config.Config)
}
