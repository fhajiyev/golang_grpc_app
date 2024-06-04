package repo_test

import (
	"errors"
	"testing"
	"time"

	"github.com/bxcodec/faker"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbcontentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func (ts *RepoTestSuite) TestRepo_GetContentCampaign() {
	var mockContentCampaign *dbcontentcampaign.ContentCampaign
	err := faker.FakeData(&mockContentCampaign)

	ts.NoError(err)
	ts.redisCache.On("GetCache", mock.Anything, mock.Anything).Return(errors.New("cache: key is missing")).Once()
	ts.dbSource.On("GetContentCampaignByID", mock.Anything).Return(mockContentCampaign, nil).Once()
	ts.redisCache.On("SetCacheAsync", mock.Anything, mock.Anything, mock.Anything).Return().Once()

	var result *contentcampaign.ContentCampaign
	result, err = ts.repo.GetContentCampaignByID(mockContentCampaign.ID)

	ts.NoError(err)
	ts.NotNil(result)
	ts.Equal(mockContentCampaign.ID, result.ID)
	ts.Equal(mockContentCampaign.CreatedAt, result.CreatedAt)
	ts.Equal(mockContentCampaign.UpdatedAt, result.UpdatedAt)
	ts.dbSource.AssertExpectations(ts.T())
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

type RepoTestSuite struct {
	suite.Suite
	dbSource   *MockDBSource
	redisCache *MockRedisCache
	repo       contentcampaign.Repository
}

func (ts *RepoTestSuite) SetupTest() {
	ts.dbSource = &MockDBSource{}
	ts.redisCache = &MockRedisCache{}
	ts.repo = repo.New(ts.dbSource, ts.redisCache, nil)
}

var _ dbcontentcampaign.DBSource = &MockDBSource{}
var __ rediscache.RedisSource = &MockRedisCache{}

type MockDBSource struct {
	mock.Mock
}

func (ds *MockDBSource) GetContentCampaignByID(contentCampaignID int64) (*dbcontentcampaign.ContentCampaign, error) {
	ret := ds.Called(contentCampaignID)
	return ret.Get(0).(*dbcontentcampaign.ContentCampaign), ret.Error(1)
}

type MockRedisCache struct {
	mock.Mock
}

func (mrc *MockRedisCache) GetCache(key string, obj interface{}) error {
	ret := mrc.Called(key, obj)
	return ret.Error(0)
}

func (mrc *MockRedisCache) SetCacheAsync(key string, obj interface{}, expiration time.Duration) {
	mrc.Called(key, obj, expiration)
}

func (mrc *MockRedisCache) SetCache(key string, obj interface{}, expiration time.Duration) error {
	ret := mrc.Called(key, obj, expiration)
	return ret.Error(1)
}

func (mrc *MockRedisCache) DeleteCache(key string) error {
	ret := mrc.Called(key)
	return ret.Error(0)
}
