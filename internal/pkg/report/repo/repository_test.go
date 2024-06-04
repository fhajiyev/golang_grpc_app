package repo_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/bxcodec/faker"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/report"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/report/repo"
	"github.com/Buzzvil/go-test/mock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
)

func (ts *RepoTestSuite) Test_SaveContentReport() {
	var cr report.Request
	faker.FakeData(&cr)
	req := "INSERT INTO `content_reported` (`campaign_id`,`campaign_name`,`description`,`device_id`,`html`,`icon_url`,`ifa`,`image_url`,`landing_url`,`report_reason`,`title`,`unit_id`,`unit_device_token`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	ts.mock.ExpectExec(fixedFullRe(req)).WithArgs(cr.CampaignID, cr.CampaignName, cr.Description, cr.DeviceID, cr.HTML, cr.IconURL, cr.IFA, cr.ImageURL, cr.LandingURL, cr.ReportReason, cr.Title, cr.UnitID, cr.UnitDeviceToken, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	err := ts.repo.SaveContentReport(cr)
	ts.NoError(err)
}

func (ts *RepoTestSuite) Test_SaveAdReport() {
	var cr report.Request
	faker.FakeData(&cr)
	httpClient := network.DefaultHTTPClient
	buzzAdServer := mock.NewTargetServer(network.GetHost(ts.buzzAdURL)).AddResponseHandler(&mock.ResponseHandler{
		WriteToBody: func() []byte {
			return []byte(`{"status": "ok"}`)
		},
		StatusCode: http.StatusOK,
		Path:       "/api/report/ads",
		Method:     http.MethodPost,
	})
	clientPatcher := mock.PatchClient(httpClient, buzzAdServer)
	defer clientPatcher.RemovePatch()

	err := ts.repo.SaveAdReport(cr)
	ts.NoError(err)
}

func fixedFullRe(s string) string {
	return fmt.Sprintf("^%s$", regexp.QuoteMeta(s))
}

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

var (
	_ suite.SetupTestSuite = &RepoTestSuite{}
)

type RepoTestSuite struct {
	suite.Suite
	mock      sqlmock.Sqlmock
	db        *gorm.DB
	repo      report.Repository
	buzzAdURL string
}

func (ts *RepoTestSuite) SetupTest() {
	ts.buzzAdURL = os.Getenv("BUZZAD_URL")
	db, mock, err := sqlmock.New()
	ts.NoError(err)
	ts.mock = mock
	ts.db, err = gorm.Open("mysql", db)
	ts.NoError(err)
	ts.db.LogMode(true)
	ts.db = ts.db.Set("gorm:update_column", true)
	ts.repo = repo.New(ts.db, ts.buzzAdURL)
}
