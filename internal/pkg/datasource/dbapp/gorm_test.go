package dbapp_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbapp"
	"github.com/bxcodec/faker"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

func (ts *RepoTestSuite) Test_GetAppByID() {
	var dbApp dbapp.App
	ts.NoError(faker.FakeData(&dbApp))
	dbApp.ID = 1234
	req := "SELECT * FROM `apps` WHERE (`apps`.`id` = ?) ORDER BY `apps`.`id` ASC LIMIT 1"
	ts.mock.ExpectQuery(fixedFullRe(req)).WillReturnRows(getRowsForModel(dbApp))

	var res *dbapp.App
	res, err := ts.dbSource.GetAppByID(context.Background(), dbApp.ID)
	ts.NoError(err)
	ts.Equal(dbApp.ID, res.ID)
	ts.Equal(dbApp.LatestAppVersion, res.LatestAppVersion)
}

func (ts *RepoTestSuite) Test_GetUnitByID() {
	var unit dbapp.Unit
	ts.NoError(faker.FakeData(&unit))
	unit.ID = 1234

	query := &dbapp.Unit{ID: unit.ID}
	req := "SELECT * FROM `unit` WHERE (`unit`.`id` = ?) ORDER BY `unit`.`id` ASC LIMIT 1"
	ts.mock.ExpectQuery(fixedFullRe(req)).WillReturnRows(getRowsForModel(unit))

	res, err := ts.dbSource.GetUnit(context.Background(), query)
	ts.NoError(err)
	ts.Equal(unit.ID, res.ID)
}

func (ts *RepoTestSuite) Test_FindRewardingWelcomeRewardConfigs() {
	rewardConfigs := getTestWRCs(1)
	req := "SELECT * FROM `welcome_reward_config` WHERE (is_exhausted = ? and unit_id = ? and start_time < ? and (max_num_rewards is not null or ? < DATE_ADD(end_time, INTERVAL retention_days DAY)) and is_terminated = ?)"
	ts.mock.ExpectQuery(fixedFullRe(req)).
		WillReturnRows(getRowsForWRCs(rewardConfigs))

	wrc := rewardConfigs[0]
	var result []dbapp.WelcomeRewardConfig
	result, err := ts.dbSource.FindRewardingWelcomeRewardConfigs(context.Background(), wrc.UnitID)

	ts.NoError(err)
	ts.Equal(*rewardConfigs[0], (result)[0])
}

func (ts *RepoTestSuite) Test_FindReferralRewardConfig() {
	rrcs := ts.getTestRRCs(1)
	req := "SELECT * FROM `referral_reward_config` WHERE (`referral_reward_config`.`app_id` = ?) ORDER BY `referral_reward_config`.`app_id` ASC LIMIT 1"
	ts.mock.ExpectQuery(fixedFullRe(req)).WillReturnRows(ts.getRowsForRRCs(rrcs))

	rrc := rrcs[0]
	result, err := ts.dbSource.FindReferralRewardConfig(context.Background(), rrc.AppID)

	ts.NoError(err)
	ts.Equal(rrc, result)
}

func getRowsForModel(ap interface{}) *sqlmock.Rows {
	names, fields := getListFields(ap)
	rows := sqlmock.NewRows(names)
	rows = rows.AddRow(fields[:]...)
	return rows
}

func getTestWRCs(num int) []*dbapp.WelcomeRewardConfig {
	rewardConfigs := make([]*dbapp.WelcomeRewardConfig, 0)
	for i := 0; i < num; i++ {
		cou := "KR"
		n := &dbapp.WelcomeRewardConfig{
			UnitID:  1234,
			Amount:  500,
			Name:    "Campaign name",
			Country: &cou,
		}
		rewardConfigs = append(rewardConfigs, n)
	}
	return rewardConfigs
}

func (ts *RepoTestSuite) getTestRRCs(num int) []*dbapp.ReferralRewardConfig {
	rrcs := make([]*dbapp.ReferralRewardConfig, 0)
	for i := 0; i < num; i++ {
		var rrc dbapp.ReferralRewardConfig
		ts.NoError(faker.FakeData(&rrc))
		rrc.AppID++
		rrcs = append(rrcs, &rrc)
	}
	return rrcs
}

func getRowsForWRCs(wrcs []*dbapp.WelcomeRewardConfig) *sqlmock.Rows {
	var fieldNames = []string{"id", "name", "amount", "country", "unit_id"}
	rows := sqlmock.NewRows(fieldNames)
	for _, c := range wrcs {
		rows = rows.AddRow(c.ID, c.Name, c.Amount, *c.Country, c.UnitID)
	}
	return rows
}

func (ts *RepoTestSuite) getRowsForRRCs(rrcs []*dbapp.ReferralRewardConfig) *sqlmock.Rows {
	ts.NotEqual(len(rrcs), 0)
	rrc := rrcs[0]
	names, _ := getListFields(*rrc)
	rows := sqlmock.NewRows(names)
	for _, c := range rrcs {
		_, fields := getListFields(*c)
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
	mock     sqlmock.Sqlmock
	db       *gorm.DB
	dbSource dbapp.DBSource
}

func (ts *RepoTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	ts.NoError(err)
	ts.mock = mock
	ts.db, err = gorm.Open("mysql", db)
	ts.NoError(err)
	ts.db.LogMode(true)
	ts.db = ts.db.Set("gorm:update_column", true)
	ts.dbSource = dbapp.NewSource(ts.db)
}

func (ts *RepoTestSuite) AfterTest(suiteName, testName string) {
	_ = ts.db.Close()
}
