package repo_test

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbnotiplus"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/notiplus"
	notiplusRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/notiplus/repo"

	"github.com/bxcodec/faker"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

func (ts *RepoTestSuite) Test_GetConfigsByUnitID() {
	var configs []dbnotiplus.Config
	for i := 1; i < 3; i++ {
		var config dbnotiplus.Config
		ts.NoError(faker.FakeData(&config))
		config.UnitID = 1234
		config.ID = int64(i)
		configs = append(configs, config)
	}

	req := "SELECT * FROM `noti_plus_configs` WHERE (unit_id = ?) ORDER BY schedule_hour_minute ASC"
	ts.mock.ExpectQuery(fixedFullRe(req)).WillReturnRows(getRowsForNotiPlusConfigs(configs))

	res, err := ts.repo.GetConfigsByUnitID(1234)
	ts.NoError(err)
	ts.Equal(2, len(res))
}

func getRowsForNotiPlusConfigs(configs []dbnotiplus.Config) *sqlmock.Rows {
	names, _ := getListFields(configs[0])
	rows := sqlmock.NewRows(names)
	for _, config := range configs {
		_, fields := getListFields(config)
		rows = rows.AddRow(fields[:]...)
	}
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

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

type RepoTestSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
	db   *gorm.DB
	repo notiplus.Repository
}

func (ts *RepoTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	ts.NoError(err)
	ts.mock = mock
	ts.db, err = gorm.Open("mysql", db)
	ts.NoError(err)
	ts.db.LogMode(true)
	ts.db = ts.db.Set("gorm:update_column", true)
	ts.repo = notiplusRepo.New(ts.db)
}

func (ts *RepoTestSuite) AfterTest(suiteName, testName string) {
	_ = ts.db.Close()
}
