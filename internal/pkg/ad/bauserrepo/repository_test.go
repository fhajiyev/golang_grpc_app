package bauserrepo

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/bxcodec/faker"

	"gopkg.in/DATA-DOG/go-sqlmock.v2"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

func (ts *RepoTestSuite) Test_GetBAUserByID_FromCache() {
	var baUser ad.BAUser
	ts.NoError(faker.FakeData(&baUser))
	baUser.ID++

	getCacheFunc := func(key string, v interface{}) error {
		cache := v.(*baUserCache)
		cache.BAUser = baUser
		cache.CreatedAt = time.Now()
		return nil
	}

	cacheKey := ts.repo.getCacheKey(baUser.ID)
	ts.redisCache.On("GetCache", cacheKey, mock.AnythingOfType("*bauserrepo.baUserCache")).Return(getCacheFunc).Once()

	res, err := ts.repo.GetBAUserByID(baUser.ID)
	ts.NoError(err)

	ts.Equal(baUser.ID, res.ID)
	ts.Equal(baUser.AccessToken, res.AccessToken)
}

func (ts *RepoTestSuite) Test_GetBAUserByID_FromDB() {
	var baUser ad.BAUser
	ts.NoError(faker.FakeData(&baUser))
	baUser.ID++

	dbBAUser := BAUser{
		ID:          baUser.ID,
		AccessToken: baUser.AccessToken,
		Name:        baUser.Name,
		IsMedia:     baUser.IsMedia,
	}

	req := "SELECT * FROM `buzzad_user` WHERE (id = ?)"
	ts.mock.ExpectQuery(fixedFullRe(req)).WillReturnRows(getRowsForBAUser(dbBAUser))
	cacheKey := ts.repo.getCacheKey(baUser.ID)
	ts.redisCache.On("GetCache", cacheKey, mock.AnythingOfType("*bauserrepo.baUserCache")).Return(errors.New("Not cached")).Once()
	ts.redisCache.On("SetCacheAsync", cacheKey, mock.AnythingOfType("*bauserrepo.baUserCache"), time.Hour*24).Once()

	res, err := ts.repo.GetBAUserByID(baUser.ID)
	ts.NoError(err)

	log.Println(res)
	ts.Equal(baUser.ID, res.ID)
	ts.Equal(baUser.AccessToken, res.AccessToken)
}

func getRowsForBAUser(baUser BAUser) *sqlmock.Rows {
	names, _ := getListFields(baUser)
	rows := sqlmock.NewRows(names)
	_, fields := getListFields(baUser)
	rows = rows.AddRow(fields[:]...)
	return rows
}

func getListFields(a interface{}) ([]string, []driver.Value) {
	elements := reflect.ValueOf(a)
	var names []string
	var values []driver.Value

	for i := 0; i < elements.NumField(); i++ {
		names = append(names, gorm.ToDBName(elements.Type().Field(i).Name))
		values = append(values, elements.Field(i).Interface())
	}

	return names, values
}

func fixedFullRe(s string) string {
	return fmt.Sprintf("^%s$", regexp.QuoteMeta(s))
}

type RepoTestSuite struct {
	suite.Suite
	mock       sqlmock.Sqlmock
	db         *gorm.DB
	redisCache *MockRedisCache
	repo       *Repository
}

func (ts *RepoTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	ts.NoError(err)
	ts.mock = mock
	ts.db, err = gorm.Open("mysql", db)
	ts.NoError(err)
	ts.db.LogMode(true)
	ts.db = ts.db.Set("gorm:update_column", true)
	ts.redisCache = &MockRedisCache{}
	ts.repo = New(ts.db, ts.redisCache)
}

func (ts *RepoTestSuite) AfterTest(suiteName, testName string) {
	_ = ts.db.Close()
}

type MockRedisCache struct {
	mock.Mock
}

func (mrc *MockRedisCache) GetCache(key string, obj interface{}) error {
	ret := mrc.Called(key, obj)
	fun, ok := ret.Get(0).(func(string, interface{}) error)
	if ok {
		return fun(key, obj)
	}
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
