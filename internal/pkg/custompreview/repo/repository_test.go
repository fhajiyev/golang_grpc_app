package repo

import (
	"database/sql/driver"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview"
	"github.com/bxcodec/faker"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
)

const (
	dateFormat = "2006-01-02"
)

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

type RepoTestSuite struct {
	suite.Suite
	mock             sqlmock.Sqlmock
	db               *gorm.DB
	repo             custompreview.Repository
	mapper           entityMapper
	customColumnName map[string]string
}

func (ts *RepoTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	ts.NoError(err)
	ts.mock = mock
	ts.db, err = gorm.Open("mysql", db)
	ts.NoError(err)
	ts.db.LogMode(true)
	ts.db = ts.db.Set("gorm:update_column", true)
	ts.repo = New(ts.db)

	ts.customColumnName = make(map[string]string)
	ts.customColumnName["DIPU"] = "dipu"
	ts.customColumnName["TIPU"] = "tipu"
	ts.customColumnName["DCPU"] = "dcpu"
	ts.customColumnName["TCPU"] = "tcpu"
}

func (ts *RepoTestSuite) AfterTest() {
	_ = ts.db.Close()
}

func (ts *RepoTestSuite) Test_Get() {
	config, dbconfig := ts.getTestConfig()
	targetTime := config.StartDate.Add(time.Hour * time.Duration(24))

	req := "SELECT * FROM `custom_preview_config` WHERE (unit_id = ? AND is_active = ? AND start_date <= ? AND end_date >= ? ) ORDER BY `custom_preview_config`.`id` DESC LIMIT 1"
	ts.mock.ExpectQuery(ts.fixedFullRe(req)).WithArgs(config.UnitID, true, targetTime, targetTime).WillReturnRows(ts.getRowsForConfig(dbconfig))

	res, err := ts.repo.GetConfigByUnitID(config.UnitID, true, targetTime)
	ts.NoError(err)
	ts.equalConfig(config, *res)
}

func (ts *RepoTestSuite) getTestConfig() (custompreview.Config, DBConfig) {
	var dbconfig DBConfig
	ts.NoError(faker.FakeData(&dbconfig))

	startDate, err := time.Parse(dateFormat, dbconfig.StartDate.Format(dateFormat))
	ts.NoError(err)
	dbconfig.StartDate = startDate

	endDate := startDate.Add(time.Hour * time.Duration(24*rand.Intn(100)+1))
	dbconfig.EndDate = endDate

	dbconfig.ID++
	dbconfig.UnitID++

	*dbconfig.DIPU++
	*dbconfig.TIPU++
	*dbconfig.DCPU++
	*dbconfig.TCPU++

	dbconfig.IsActive = true

	return *ts.mapper.dbConfigToConfig(dbconfig), dbconfig
}

func (ts *RepoTestSuite) equalConfig(config1 custompreview.Config, config2 custompreview.Config) {
	ts.Equal(config1.UnitID, config2.UnitID)
	ts.Equal(config1.Message, config2.Message)
	ts.Equal(config1.LandingURL, config2.LandingURL)

	ts.Equal(config1.Period.StartDate, config2.Period.StartDate)
	ts.Equal(config1.Period.EndDate, config2.Period.EndDate)
	ts.Equal(config1.Period.StartHourMinute, config2.Period.StartHourMinute)
	ts.Equal(config1.Period.EndHourMinute, config2.Period.EndHourMinute)

	ts.Equal(config1.FrequencyLimit.DIPU, config2.FrequencyLimit.DIPU)
	ts.Equal(config1.FrequencyLimit.TIPU, config2.FrequencyLimit.TIPU)
	ts.Equal(config1.FrequencyLimit.DCPU, config2.FrequencyLimit.DCPU)
	ts.Equal(config1.FrequencyLimit.TCPU, config2.FrequencyLimit.TCPU)

	ts.Equal(*config1.Icon, *config2.Icon)
}

func (ts *RepoTestSuite) fixedFullRe(s string) string {
	return fmt.Sprintf("^%s$", regexp.QuoteMeta(s))
}

func (ts *RepoTestSuite) getRowsForConfig(config DBConfig) *sqlmock.Rows {
	names, fields := ts.getListFields(config)
	rows := sqlmock.NewRows(names)
	rows = rows.AddRow(fields[:]...)
	return rows
}

func (ts *RepoTestSuite) getListFields(a interface{}) ([]string, []driver.Value) {
	elements := reflect.ValueOf(a)
	var names []string
	var values []driver.Value
	for i := 0; i < elements.NumField(); i++ {
		fieldName := elements.Type().Field(i).Name
		val, exists := ts.customColumnName[fieldName]
		dbName := val

		if !exists {
			dbName = gorm.ToDBName(fieldName)
		}

		names = append(names, dbName)
		values = append(values, elements.Field(i).Interface())
	}
	return names, values
}
