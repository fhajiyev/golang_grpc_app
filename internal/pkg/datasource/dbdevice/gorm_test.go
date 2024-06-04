package dbdevice_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/bxcodec/faker"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
)

func (ts *DSTestSuite) Test_GetByID() {
	var dbDevice dbdevice.Device
	ts.NoError(faker.FakeData(&dbDevice))
	dbDevice.ID += 1 // Random 으로 두면 가끔 0 으로 대입되면서 Test 가 실패함
	req := "SELECT * FROM `device` WHERE `device`.`id` = ?"
	ts.mock.ExpectQuery(fixedFullRe(req)).
		WillReturnRows(getRowsForWRCs(&dbDevice))

	result, err := ts.dbSource.GetByID(dbDevice.ID)

	ts.NoError(err)
	ts.Equal(dbDevice.ID, result.ID)
	ts.Equal(dbDevice.IFA, result.IFA)
	ts.Equal(dbDevice.UnitDeviceToken, result.UnitDeviceToken)
	ts.Equal(dbDevice.AppID, result.AppID)
}

func (ts *DSTestSuite) Test_GetByAppIDAndIFA() {
	var dbDevice dbdevice.Device
	ts.NoError(faker.FakeData(&dbDevice))
	dbDevice.AppID += 1 // Random 으로 두면 가끔 0 으로 대입되면서 Test 가 실패함
	req := "SELECT * FROM `device` WHERE (`device`.`app_id` = ?) AND (`device`.`ifa` = ?)"
	ts.mock.ExpectQuery(fixedFullRe(req)).
		WillReturnRows(getRowsForWRCs(&dbDevice))

	result, err := ts.dbSource.GetByAppIDAndIFA(dbDevice.AppID, dbDevice.IFA)

	ts.NoError(err)
	ts.Equal(dbDevice.ID, result.ID)
	ts.Equal(dbDevice.IFA, result.IFA)
	ts.Equal(dbDevice.UnitDeviceToken, result.UnitDeviceToken)
	ts.Equal(dbDevice.AppID, result.AppID)
}

func (ts *DSTestSuite) Test_GetByAppIDAndPubUserID() {
	var dbDevice dbdevice.Device
	ts.NoError(faker.FakeData(&dbDevice))
	dbDevice.AppID += 1 // Random 으로 두면 가끔 0 으로 대입되면서 Test 가 실패함
	req := "SELECT * FROM `device` WHERE (`device`.`app_id` = ?) AND (`device`.`unit_device_token` = ?)"
	ts.mock.ExpectQuery(fixedFullRe(req)).
		WillReturnRows(getRowsForWRCs(&dbDevice))

	result, err := ts.dbSource.GetByAppIDAndPubUserID(dbDevice.AppID, dbDevice.UnitDeviceToken)

	ts.NoError(err)
	ts.Equal(dbDevice.ID, result.ID)
	ts.Equal(dbDevice.IFA, result.IFA)
	ts.Equal(dbDevice.UnitDeviceToken, result.UnitDeviceToken)
	ts.Equal(dbDevice.AppID, result.AppID)
}

func getRowsForWRCs(d *dbdevice.Device) *sqlmock.Rows {
	var fieldNames = []string{"id", "app_id", "unit_device_token", "ifa", "address", "birthday", "carrier", "device_name", "resolution", "year_of_birth", "sdk_version", "sex", "packages", "package_name", "signup_ip", "serial_number"}
	rows := sqlmock.NewRows(fieldNames)
	rows = rows.AddRow(d.ID, d.AppID, d.UnitDeviceToken, d.IFA, d.Address, d.Birthday, d.Carrier, d.DeviceName, d.Resolution, d.YearOfBirth, d.SDKVersion, d.Sex, d.Packages, d.PackageName, d.SignupIP, d.SerialNumber)
	return rows
}

func fixedFullRe(s string) string {
	return fmt.Sprintf("^%s$", regexp.QuoteMeta(s))
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(DSTestSuite))
}

type DSTestSuite struct {
	suite.Suite
	mock     sqlmock.Sqlmock
	db       *gorm.DB
	dbSource dbdevice.DBSource
}

func (ts *DSTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	ts.NoError(err)
	ts.mock = mock
	ts.db, err = gorm.Open("mysql", db)
	ts.NoError(err)
	ts.db.LogMode(true)
	ts.db = ts.db.Set("gorm:update_column", true)
	ts.dbSource = dbdevice.NewSource(ts.db)
}

func (ts *DSTestSuite) AfterTest(suiteName, testName string) {
	_ = ts.db.Close()
}
