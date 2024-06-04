package dbcontentcampaign_test

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbcontentcampaign"

	"github.com/bxcodec/faker"

	"gopkg.in/DATA-DOG/go-sqlmock.v2"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
)

func (ts *RepoTestSuite) Test_GetContentCampaignByID() {
	var dbContentCampaign dbcontentcampaign.ContentCampaign
	ts.NoError(faker.FakeData(&dbContentCampaign))
	dbContentCampaign.ID += 1

	req := "SELECT * FROM `content_campaigns` WHERE (`content_campaigns`.`id` = ?)"
	ts.mock.ExpectQuery(fixedFullRe(req)).
		WillReturnRows(getRowsForContentCampaigns(&dbContentCampaign))

	var result *dbcontentcampaign.ContentCampaign
	result, err := ts.dbSource.GetContentCampaignByID(dbContentCampaign.ID)

	ts.NoError(err)
	fmt.Println(result)
	ts.Equal(dbContentCampaign.ID, result.ID)
	ts.Equal(dbContentCampaign.CreatedAt, result.CreatedAt)
	ts.Equal(dbContentCampaign.UpdatedAt, result.UpdatedAt)
}

func getRowsForContentCampaigns(cc *dbcontentcampaign.ContentCampaign) *sqlmock.Rows {
	names, fields := getListFields(*cc)
	rows := sqlmock.NewRows(names)
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

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

type RepoTestSuite struct {
	suite.Suite
	mock     sqlmock.Sqlmock
	db       *gorm.DB
	dbSource dbcontentcampaign.DBSource
}

func (ts *RepoTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	ts.NoError(err)
	ts.mock = mock
	ts.db, err = gorm.Open("mysql", db)
	ts.NoError(err)
	ts.db.LogMode(true)
	ts.db = ts.db.Set("gorm:update_column", true)
	ts.dbSource = dbcontentcampaign.NewSource(ts.db)
}

func (ts *RepoTestSuite) AfterTest(suiteName, testName string) {
	_ = ts.db.Close()
}
