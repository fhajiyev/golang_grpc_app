package repo

import (
	"database/sql/driver"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/userreferral"
	gotestmock "github.com/Buzzvil/go-test/mock"
	"github.com/bxcodec/faker"
	"github.com/go-resty/resty"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
)

func TestUserRepoSuite(t *testing.T) {
	suite.Run(t, new(UserRepoTestSuite))
}

type UserRepoTestSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
	db   *gorm.DB
	repo userreferral.Repository

	proxyURL                  string
	giveReferralRewardBaseURL string
	giveReferralRewardPath    string

	verifyDeviceClient       *resty.Client
	giveReferralRewardClient *resty.Client
}

func (ts *UserRepoTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	ts.NoError(err)
	ts.mock = mock
	ts.db, err = gorm.Open("mysql", db)
	ts.NoError(err)
	ts.db.LogMode(true)
	ts.db = ts.db.Set("gorm:update_column", true)

	ts.verifyDeviceClient = resty.New()
	ts.proxyURL = "https://proxy.com"
	ts.verifyDeviceClient.SetProxy(ts.proxyURL)
	ts.giveReferralRewardClient = resty.New()

	ts.giveReferralRewardBaseURL = "https://buzzscreen.buzzvil.com"
	ts.giveReferralRewardPath = "/reward/referral-rewards"

	ts.repo = New(
		ts.db,
		ts.giveReferralRewardBaseURL,
		ts.giveReferralRewardClient,
		ts.verifyDeviceClient,
	)
}

func (ts *UserRepoTestSuite) AfterTest(suiteName, testName string) {
	_ = ts.db.Close()
}

func (ts *UserRepoTestSuite) Test_CreateUser() {
	user, _ := ts.getTestUser()

	req := "INSERT INTO `device_user` (`device_id`,`code`,`referrer_id`,`is_verified`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?)"

	ts.mock.ExpectExec(ts.fixedFullRe(req)).WithArgs(user.DeviceID, user.Code, sqlmock.AnyArg(), user.IsVerified, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	dbuser, err := ts.repo.CreateUser(user.DeviceID, user.Code, user.IsVerified)
	ts.Equal(user.DeviceID, dbuser.DeviceID)
	ts.Equal(user.Code, dbuser.Code)
	ts.Equal(user.IsVerified, dbuser.IsVerified)
	ts.NoError(err)
}

func (ts *UserRepoTestSuite) Test_UpdateUserReferralID() {
	refereeID := int64(1 + rand.Intn(65536))
	referrerID := int64(1 + rand.Intn(65536))

	// TODO Why updated_at isn't inside sql ?
	req := "UPDATE `device_user` SET `referrer_id` = ? WHERE `device_user`.`id` = ?"

	ts.mock.ExpectExec(ts.fixedFullRe(req)).WithArgs(referrerID, refereeID).WillReturnResult(sqlmock.NewResult(1, 1))

	err := ts.repo.UpdateUserReferrerID(refereeID, referrerID)
	ts.NoError(err)
}

func (ts *UserRepoTestSuite) Test_GetUserByDevice() {
	user, dbuser := ts.getTestUser()

	req := "SELECT * FROM `device_user` WHERE (`device_user`.`device_id` = ?) ORDER BY `device_user`.`id` ASC LIMIT 1"
	ts.mock.ExpectQuery(ts.fixedFullRe(req)).WillReturnRows(ts.getRowsForUser(*dbuser))

	res, err := ts.repo.GetUserByDevice(user.DeviceID)
	ts.NoError(err)
	ts.equalUser(user, res)
}

func (ts *UserRepoTestSuite) Test_GetUserByCode() {
	user, dbuser := ts.getTestUser()

	req := "SELECT * FROM `device_user` WHERE (`device_user`.`code` = ?) ORDER BY `device_user`.`id` ASC LIMIT 1"
	ts.mock.ExpectQuery(ts.fixedFullRe(req)).WillReturnRows(ts.getRowsForUser(*dbuser))

	res, err := ts.repo.GetUserByCode(user.Code)
	ts.NoError(err)
	ts.equalUser(user, res)
}

func (ts *UserRepoTestSuite) Test_GetReferralCountByUser() {
	referrerID := int64(1 + rand.Intn(65536))
	cnt := 1 + rand.Intn(65536)

	req := "SELECT count(*) FROM `device_user` WHERE (`device_user`.`referrer_id` = ?)"
	ts.mock.ExpectQuery(ts.fixedFullRe(req)).WillReturnRows(ts.getRowsForCount(cnt))

	dbcnt, err := ts.repo.GetReferralCountByUser(referrerID)
	ts.NoError(err)
	ts.Equal(cnt, dbcnt)
}

func (ts *UserRepoTestSuite) Test_IsVerifiedDevice() {
	verifyURL := "http://verify.com"
	udt := "abcd"

	ts.Run("success", func() {
		mockServer := ts.getVerifyDeviceMockServer(verifyURL, http.StatusOK)
		clientPatcher := gotestmock.PatchClient(ts.verifyDeviceClient.GetClient(), mockServer)
		defer clientPatcher.RemovePatch()

		err := ts.repo.IsVerifiedDevice(verifyURL, udt)
		ts.NoError(err)
	})

	ts.Run("fail", func() {
		mockServer := ts.getVerifyDeviceMockServer(verifyURL, http.StatusNotFound)
		clientPatcher := gotestmock.PatchClient(ts.verifyDeviceClient.GetClient(), mockServer)
		defer clientPatcher.RemovePatch()

		err := ts.repo.IsVerifiedDevice(verifyURL, udt)
		_, ok := err.(userreferral.APICallError)

		ts.Error(err)
		ts.True(ok)
	})
}

func (ts *UserRepoTestSuite) Test_GiveReferralReward_Success() {
	ingr := ts.getTestGiveReferralRewardRequestIngredients()

	ts.Run("success", func() {
		mockServer := ts.getGiveReferralMockServer(http.StatusOK)
		clientPatcher := gotestmock.PatchClient(ts.giveReferralRewardClient.GetClient(), mockServer)
		defer clientPatcher.RemovePatch()

		err := ts.repo.GiveReferralReward(*ingr)
		ts.NoError(err)
	})

	ts.Run("fail", func() {
		mockServer := ts.getGiveReferralMockServer(http.StatusNotFound)
		clientPatcher := gotestmock.PatchClient(ts.giveReferralRewardClient.GetClient(), mockServer)
		defer clientPatcher.RemovePatch()

		err := ts.repo.GiveReferralReward(*ingr)
		_, ok := err.(userreferral.APICallError)

		ts.Error(err)
		ts.True(ok)
	})
}

func (ts *UserRepoTestSuite) getVerifyDeviceMockServer(verifyURL string, statusCode int) *gotestmock.TargetServer {
	server := gotestmock.NewTargetServer(network.GetHost(verifyURL)).AddResponseHandler(&gotestmock.ResponseHandler{
		WriteToBody: func() []byte {
			return []byte(string(""))
		},
		Path:       "",
		Method:     http.MethodGet,
		StatusCode: statusCode,
	})
	return server
}

func (ts *UserRepoTestSuite) getGiveReferralMockServer(statusCode int) *gotestmock.TargetServer {
	server := gotestmock.NewTargetServer(network.GetHost(ts.giveReferralRewardBaseURL)).AddResponseHandler(&gotestmock.ResponseHandler{
		WriteToBody: func() []byte {
			return []byte(string(""))
		},
		Path:       ts.giveReferralRewardPath,
		Method:     http.MethodPost,
		StatusCode: statusCode,
	})
	return server
}

func (ts *UserRepoTestSuite) fixedFullRe(s string) string {
	return fmt.Sprintf("^%s$", regexp.QuoteMeta(s))
}

func (ts *UserRepoTestSuite) equalUser(user1 *userreferral.DeviceUser, user2 *userreferral.DeviceUser) {
	ts.True(user1.ID == user2.ID)
	ts.True(user1.DeviceID == user2.DeviceID)
	ts.True(user1.Code == user2.Code)
	ts.True(user1.ReferrerID == user2.ReferrerID)
	ts.True(user1.IsVerified == user2.IsVerified)
}

func (ts *UserRepoTestSuite) getTestUser() (*userreferral.DeviceUser, *DBDeviceUser) {
	var user userreferral.DeviceUser
	ts.NoError(faker.FakeData(&user))
	user.ID++
	user.DeviceID++

	return &user, &DBDeviceUser{
		ID:         user.ID,
		DeviceID:   user.DeviceID,
		Code:       user.Code,
		ReferrerID: user.ReferrerID,
		IsVerified: user.IsVerified,
	}
}

func (ts *UserRepoTestSuite) getTestGiveReferralRewardRequestIngredients() *userreferral.GiveReferralRewardRequestIngredients {
	var ingr userreferral.GiveReferralRewardRequestIngredients
	ts.NoError(faker.FakeData(&ingr))
	ingr.RefereeDeviceID++
	ingr.ReferrerDeviceID++
	return &ingr
}

func (ts *UserRepoTestSuite) getRowsForCount(cnt int) *sqlmock.Rows {
	var fieldNames = []string{"count"}
	rows := sqlmock.NewRows(fieldNames)
	rows.AddRow(cnt)
	return rows
}

func (ts *UserRepoTestSuite) getRowsForUser(user DBDeviceUser) *sqlmock.Rows {
	names, _ := ts.getListFields(user)
	rows := sqlmock.NewRows(names)
	_, fields := ts.getListFields(user)
	rows = rows.AddRow(fields[:]...)
	return rows
}

func (ts *UserRepoTestSuite) getListFields(a interface{}) ([]string, []driver.Value) {
	elements := reflect.ValueOf(a)
	var names []string
	var values []driver.Value
	for i := 0; i < elements.NumField(); i++ {
		names = append(names, gorm.ToDBName(elements.Type().Field(i).Name))
		values = append(values, elements.Field(i).Interface())
	}

	return names, values
}
