package repo_test

import (
	"context"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth/repo"

	authsvc "github.com/Buzzvil/buzzapis/go/auth"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type RepoTestSuite struct {
	suite.Suite
	client *authServiceClient
	repo   auth.Repository
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

func (ts *RepoTestSuite) TestCreateAuth() {
	createdAuth := authsvc.CreateAuthResponse{}
	err := faker.FakeData(&createdAuth)
	ts.Nil(err)
	identifier := auth.Identifier{}
	err = faker.FakeData(&identifier)
	ts.Nil(err)

	ts.client.On("CreateAuth", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*buzzvil_auth_v1.CreateAuthRequest"), mock.AnythingOfType("[]grpc.CallOption")).Return(&createdAuth, nil).Once()

	token, err := ts.repo.CreateAuth(identifier)

	ts.Nil(err)
	ts.NotNil(token)
	ts.client.AssertExpectations(ts.T())

	ts.Equal(createdAuth.Token, token)
}

func (ts *RepoTestSuite) TestGetAuth() {
	returnedAuth := authsvc.Auth{}
	err := faker.FakeData(&returnedAuth)
	ts.Nil(err)
	var token string
	err = faker.FakeData(&token)
	ts.Nil(err)

	ts.client.On("GetAuth", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*buzzvil_auth_v1.GetAuthRequest"), mock.AnythingOfType("[]grpc.CallOption")).Return(&returnedAuth, nil).Once()

	at, err := ts.repo.GetAuth(token)

	ts.Nil(err)
	ts.NotNil(at)
	ts.client.AssertExpectations(ts.T())

	ts.Equal(returnedAuth.AccountId, at.AccountID)
	ts.Equal(returnedAuth.Identifier.AppId, at.AppID)
	ts.Equal(returnedAuth.Identifier.PublisherUserId, at.PublisherUserID)
	ts.Equal(returnedAuth.Identifier.Ifa, at.IFA)
}

func (ts *RepoTestSuite) SetupTest() {
	ts.client = &authServiceClient{}
	ts.repo = repo.New(ts.client)
}

type authServiceClient struct {
	mock.Mock
}

func (c *authServiceClient) CreateAuth(ctx context.Context, in *authsvc.CreateAuthRequest, opts ...grpc.CallOption) (*authsvc.CreateAuthResponse, error) {
	args := c.Called(ctx, in, opts)
	return args.Get(0).(*authsvc.CreateAuthResponse), args.Error(1)
}

func (c *authServiceClient) GetAuth(ctx context.Context, in *authsvc.GetAuthRequest, opts ...grpc.CallOption) (*authsvc.Auth, error) {
	args := c.Called(ctx, in, opts)
	return args.Get(0).(*authsvc.Auth), args.Error(1)
}

func (c *authServiceClient) BuildTestToken(ctx context.Context, in *authsvc.BuildTestTokenRequest, opts ...grpc.CallOption) (*authsvc.BuildTestTokenResponse, error) {
	args := c.Called(ctx, in, opts)
	return args.Get(0).(*authsvc.BuildTestTokenResponse), args.Error(1)
}
