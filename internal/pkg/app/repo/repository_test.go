package repo_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbapp"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const cacheExpiration = time.Hour * 24

func (ts *RepoTestSuite) Test_GetUnitByID() {
	unit := &app.Unit{}
	ts.NoError(faker.FakeData(&unit))
}

func (ts *RepoTestSuite) TestGetAppByID() {
	app := &app.App{}
	ts.NoError(faker.FakeData(&app))
	dbApp := &dbapp.App{
		ID:               app.ID,
		LatestAppVersion: app.LatestAppVersion,
		IsEnabled:        app.IsEnabled,
	}
	cacheKey := fmt.Sprintf("CACHE_GO_APP-%v", app.ID)
	appCacheValidator := ts.appCacheValidator(app)

	ts.Run("GetAppFromDB", func() {
		ts.redisCache.On("GetCache", cacheKey, &repo.AppCache{}).Return(errors.New("empty")).Once()
		ts.dbSource.On("GetAppByID", app.ID).Return(dbApp, nil).Once()
		ts.redisCache.On("SetCacheAsync", cacheKey, mock.MatchedBy(appCacheValidator), cacheExpiration).Once()

		result, err := ts.repo.GetAppByID(context.Background(), app.ID)
		ts.NoError(err)
		ts.Equal(app, result)
	})

	ts.Run("CacheNilApp", func() {
		core.Logger.Infof("CacheNilApp")
		ts.redisCache.On("GetCache", cacheKey, &repo.AppCache{}).Return(errors.New("empty")).Once()
		ts.dbSource.On("GetAppByID", app.ID).Return(nil, nil).Once()
		ts.redisCache.On("SetCacheAsync", cacheKey, mock.MatchedBy(ts.appCacheValidator(nil)), cacheExpiration).Once()
		result, err := ts.repo.GetAppByID(context.Background(), app.ID)
		ts.NoError(err)
		ts.Nil(result)
	})

	ts.Run("GetAppFailed", func() {
		core.Logger.Infof("GetAppFailed")
		ts.redisCache.On("GetCache", cacheKey, &repo.AppCache{}).Return(errors.New("empty")).Once()
		ts.dbSource.On("GetAppByID", app.ID).Return(nil, errors.New("empty")).Once()

		result, err := ts.repo.GetAppByID(context.Background(), app.ID)
		ts.Error(err)
		ts.Nil(result)
	})

	ts.Run("GetAppFromCache", func() {
		setAppCache := func(_ string, obj interface{}) error {
			appCache := obj.(*repo.AppCache)
			appCache.App = app
			appCache.CreatedAt = time.Now()
			return nil
		}
		ts.redisCache.On("GetCache", cacheKey, &repo.AppCache{}).Return(setAppCache).Once()

		result, err := ts.repo.GetAppByID(context.Background(), app.ID)
		ts.NoError(err)
		ts.Equal(app, result)
	})
}

func (ts *RepoTestSuite) appCacheValidator(app *app.App) func(appCache repo.AppCache) bool {
	return func(appCache repo.AppCache) bool {
		core.Logger.Infof("app %v appcache %+v", app, appCache)
		ts.Equal(app, appCache.App)
		return true
	}
}

func (ts *RepoTestSuite) Test_GetRewardingWelcomeRewardConfigs() {
	var mockWRC *dbapp.WelcomeRewardConfig
	err := faker.FakeData(&mockWRC)
	mockWRCs := []dbapp.WelcomeRewardConfig{*mockWRC}
	ts.NoError(err)
	ts.dbSource.On("FindRewardingWelcomeRewardConfigs", mock.Anything).Return(mockWRCs, nil).Once()
	ts.redisCache.On("GetCache", mock.Anything, mock.Anything).Return(errors.New("cache is missing")).Once()
	ts.redisCache.On("SetCacheAsync", mock.Anything, mock.Anything, mock.Anything).Return().Once()

	var result app.WelcomeRewardConfigs
	result, err = ts.repo.GetRewardingWelcomeRewardConfigs(context.Background(), mockWRC.UnitID)

	ts.NoError(err)
	ts.Equal(mockWRC.Amount, (result)[0].Amount)
	ts.Equal(mockWRC.StartTime, (result)[0].StartTime)
	ts.Equal(mockWRC.EndTime, (result)[0].EndTime)
}

func (ts *RepoTestSuite) Test_GetReferralRewardConfig() {
	var mockRRC *dbapp.ReferralRewardConfig
	ts.NoError(faker.FakeData(&mockRRC))
	mockRRC.AppID++

	ts.dbSource.On("FindReferralRewardConfig", mockRRC.AppID).Return(mockRRC, nil).Once()

	result, err := ts.repo.GetReferralRewardConfig(context.Background(), mockRRC.AppID)

	ts.NoError(err)
	ts.equalRRC(mockRRC, result)
}

func (ts *RepoTestSuite) equalRRC(config1 *dbapp.ReferralRewardConfig, config2 *app.ReferralRewardConfig) {
	ts.True(config1.AppID == config2.AppID)
	ts.True(config1.Enabled == config2.Enabled)
	ts.True(config1.Amount == config2.Amount)
	ts.True(config1.MaxReferral == config2.MaxReferral)
	ts.True(*config1.StartDate == *config2.StartDate)
	ts.True(*config1.EndDate == *config2.EndDate)
	ts.True(config1.VerifyURL == config2.VerifyURL)
	ts.True(config1.TitleForReferee == config2.TitleForReferee)
	ts.True(config1.TitleForReferrer == config2.TitleForReferrer)
	ts.True(config1.TitleForMaxReferrer == config2.TitleForMaxReferrer)
	ts.True(config1.ExpireHours == config2.ExpireHours)
	ts.True(config1.MinSdkVersion == config2.MinSdkVersion)
}
func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

type RepoTestSuite struct {
	suite.Suite
	dbSource   *MockDBSource
	redisCache *MockRedisCache
	repo       app.Repository
}

func (ts *RepoTestSuite) SetupTest() {
	ts.dbSource = &MockDBSource{}
	ts.redisCache = &MockRedisCache{}
	ts.repo = repo.New(ts.dbSource, ts.redisCache)
}

var _ dbapp.DBSource = &MockDBSource{}

type MockDBSource struct {
	mock.Mock
}

func (ds *MockDBSource) GetAppByID(ctx context.Context, appID int64) (*dbapp.App, error) {
	ret := ds.Called(appID)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*dbapp.App), ret.Error(1)
}

func (ds *MockDBSource) FindRewardingWelcomeRewardConfigs(ctx context.Context, unitID int64) ([]dbapp.WelcomeRewardConfig, error) {
	ret := ds.Called(unitID)
	return ret.Get(0).([]dbapp.WelcomeRewardConfig), ret.Error(1)
}

func (ds *MockDBSource) FindReferralRewardConfig(ctx context.Context, appID int64) (*dbapp.ReferralRewardConfig, error) {
	ret := ds.Called(appID)
	return ret.Get(0).(*dbapp.ReferralRewardConfig), ret.Error(1)
}

func (ds *MockDBSource) GetUnit(ctx context.Context, unit *dbapp.Unit) (*dbapp.Unit, error) {
	ret := ds.Called(unit)
	return ret.Get(0).(*dbapp.Unit), ret.Error(1)
}

type MockRedisCache struct {
	mock.Mock
}

func (rc *MockRedisCache) SetCacheAsync(key string, obj interface{}, expiration time.Duration) {
	rc.Called(key, obj, expiration)
}
func (rc *MockRedisCache) SetCache(key string, obj interface{}, expiration time.Duration) error {
	ret := rc.Called(key, obj, expiration)
	return ret.Error(1)
}
func (rc *MockRedisCache) GetCache(key string, obj interface{}) error {
	ret := rc.Called(key, obj)
	f, ok := ret.Get(0).(func(_ string, _ interface{}) error)
	if ok {
		return f(key, obj)
	}
	return ret.Error(0)
}

func (rc *MockRedisCache) DeleteCache(key string) error {
	ret := rc.Called(key)
	return ret.Error(0)
}
