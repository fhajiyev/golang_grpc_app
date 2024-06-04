package repo

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/report"
	"github.com/google/go-querystring/query"
	"github.com/jinzhu/gorm"
)

const buzzAdCampaignIDOffset = 1000000000

// Repository struct definition
type Repository struct {
	db        *gorm.DB
	buzzAdURL string
}

/**
:reason code:
	REPORT_REASON_ETC = 1
	REPORT_REASON_SEXUALLY_INAPPROPRIATE = 2
	REPORT_REASON_VIOLENT_OR_PROHIBITED = 3
	REPORT_REASON_OFFENSIVE = 4
	REPORT_REASON_MISLEADING_OR_SCAM = 5
	REPORT_REASON_DISAGREE_WITH_IT = 6
	REPORT_REASON_IMAGE_IS_UNCLEAR = 7
	REPORT_REASON_BROKEN_WEB_PAGE = 8
	REPORT_REASON_SLOW_LANDING = 9
	REPORT_REASON_BY_TESTER = 10
*/

// SaveContentReport saves content campaign to db
func (r *Repository) SaveContentReport(camp report.Request) error {
	return r.db.Save(&ContentReported{
		CampaignID:      camp.CampaignID,
		CampaignName:    camp.CampaignName,
		Description:     camp.Description,
		DeviceID:        camp.DeviceID,
		HTML:            camp.HTML,
		IconURL:         camp.IconURL,
		IFA:             camp.IFA,
		ImageURL:        camp.ImageURL,
		LandingURL:      camp.LandingURL,
		ReportReason:    camp.ReportReason,
		Title:           camp.Title,
		UnitDeviceToken: camp.UnitDeviceToken,
		UnitID:          camp.UnitID,
	}).Error
}

// SaveAdReport saves ad to buzzad
func (r *Repository) SaveAdReport(camp report.Request) error {
	params, _ := query.Values(camp)
	lineitemID, err := strconv.ParseInt(params.Get("campaign_id"), 10, 64)
	if err != nil {
		return err
	}
	if lineitemID > buzzAdCampaignIDOffset {
		lineitemID = lineitemID - buzzAdCampaignIDOffset
	}

	params.Add("lineitem_id", strconv.FormatInt(lineitemID, 10))

	httpResponse, err := (&network.Request{
		Method: http.MethodPost,
		Params: &params,
		URL:    r.buzzAdURL + "/api/report/ads",
	}).MakeRequest()

	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != 200 {
		return errors.New("failed to report to buzzad")
	}
	return nil
}

// New returns Repository struct
func New(db *gorm.DB, buzzAdURL string) *Repository {
	return &Repository{db: db, buzzAdURL: buzzAdURL}
}
