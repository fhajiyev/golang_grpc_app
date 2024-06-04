package auth

import (
	"testing"

	"github.com/bxcodec/faker"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

func (ts *UseCaseTestSuite) TestCreateAuth() {
	identifier := Identifier{}
	err := faker.FakeData(&identifier)
	ts.Nil(err)
	var expectedToken string
	err = faker.FakeData(&expectedToken)
	ts.Nil(err)

	ts.repo.On("CreateAuth", identifier).Return(expectedToken, nil).Once()

	res, err := ts.useCase.CreateAuth(identifier)
	ts.Nil(err)
	ts.repo.AssertExpectations(ts.T())

	ts.Equal(expectedToken, res)
}

func (ts *UseCaseTestSuite) TestGetAuth() {
	var token string
	err := faker.FakeData(&token)
	ts.Nil(err)
	expectedAuth := Auth{}
	err = faker.FakeData(&expectedAuth)
	ts.Nil(err)

	ts.repo.On("GetAuth", token).Return(&expectedAuth, nil).Once()

	at, err := ts.useCase.GetAuth(token)
	ts.Nil(err)
	ts.NotNil(at)
	ts.repo.AssertExpectations(ts.T())

	ts.Equal(expectedAuth.AccountID, at.AccountID)
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

type UseCaseTestSuite struct {
	suite.Suite
	repo    *mockRepo
	useCase UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.useCase = &useCase{ts.repo}
}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) CreateAuth(identifier Identifier) (string, error) {
	args := r.Called(identifier)
	return args.Get(0).(string), args.Error(1)
}

func (r *mockRepo) GetAuth(token string) (*Auth, error) {
	args := r.Called(token)
	return args.Get(0).(*Auth), args.Error(1)
}
