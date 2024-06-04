package repo_test

import (
	"context"
	"testing"

	pb "github.com/Buzzvil/buzzapis/go/profile"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/profilerequest"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/profilerequest/repo"
	"github.com/bxcodec/faker"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

func (ts *RepoTestSuite) TestPopulateProfile() {
	account := profilerequest.Account{
		AppID:     12345,
		IFA:       "sample_ifa",
		AccountID: 54321,
		CookieID:  "sample_cookie_id",
		AppUserID: "sample_app_user_id",
	}

	profileIDStruct := &pb.ProfileID{}
	err := faker.FakeData(&profileIDStruct)
	ts.NoError(err)

	ts.grpcClient.On("GetProfileID", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("*buzzvil_profile_v1.GetProfileIDRequest"), []grpc.CallOption(nil)).Return(profileIDStruct, nil).Once()

	err = ts.r.PopulateProfile(account)
	ts.Nil(err)
}

type RepoTestSuite struct {
	suite.Suite
	grpcClient *mockProfileServiceClient
	r          *repo.Repository
}

func (ts *RepoTestSuite) SetupTest() {
	ts.grpcClient = new(mockProfileServiceClient)
	ts.r = repo.New(ts.grpcClient)
}

type mockProfileServiceClient struct {
	mock.Mock
}

func (r *mockProfileServiceClient) GetProfileID(ctx context.Context, in *pb.GetProfileIDRequest, opts ...grpc.CallOption) (*pb.ProfileID, error) {
	ret := r.Called(ctx, in, opts)
	return ret.Get(0).(*pb.ProfileID), ret.Error(1)
}

func (r *mockProfileServiceClient) ListProfileIDs(ctx context.Context, in *pb.ListProfileIDsRequest, opts ...grpc.CallOption) (*pb.ListProfileIDsResponse, error) {
	ret := r.Called(ctx, in, opts)
	return ret.Get(0).(*pb.ListProfileIDsResponse), ret.Error(1)
}

func (r *mockProfileServiceClient) GetProfileEmbedding(ctx context.Context, in *pb.GetProfileEmbeddingRequest, opts ...grpc.CallOption) (*pb.ProfileEmbedding, error) {
	ret := r.Called(ctx, in, opts)
	return ret.Get(0).(*pb.ProfileEmbedding), ret.Error(1)
}

func (r *mockProfileServiceClient) CompileProfileEmbeddings(ctx context.Context, in *pb.CompileProfileEmbeddingsRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	ret := r.Called(ctx, in, opts)
	return ret.Get(0).(*empty.Empty), ret.Error(1)
}
func (r *mockProfileServiceClient) MergeIdenticalProfiles(ctx context.Context, in *pb.MergeIdenticalProfilesRequest, opts ...grpc.CallOption) (*pb.MergeIdenticalProfilesResponse, error) {
	ret := r.Called(ctx, in, opts)
	return ret.Get(0).(*pb.MergeIdenticalProfilesResponse), ret.Error(1)
}
func (r *mockProfileServiceClient) ListUserIDs(ctx context.Context, in *pb.ListUserIDsRequest, opts ...grpc.CallOption) (*pb.ListUserIDsResponse, error) {
	ret := r.Called(ctx, in, opts)
	return ret.Get(0).(*pb.ListUserIDsResponse), ret.Error(1)
}