package dto_test

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/impressiondata"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/cypher"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/Buzzvil/go-test/test"
	uuid "github.com/satori/go.uuid"
)

func TestESContentGetDoc(t *testing.T) {
	var channelID int64 = 2
	var id int64 = 1
	ipu := 10
	ecc := &dto.ESContentCampaign{
		ContentCampaign: model.ContentCampaign{
			Categories:        "news",
			ChannelID:         &channelID,
			CleanMode:         0,
			ClickURL:          "http://daum.net",
			CreatedAt:         "2006-01-02T15:04:05",
			Country:           "KR",
			DisplayType:       "A",
			EndDate:           "2099-01-03T00:00:00",
			JSON:              `{"imgW": 800, "imgH": 400}`,
			DisplayWeight:     10,
			ID:                id,
			Ipu:               &ipu,
			IsCtrFilterOff:    false,
			IsEnabled:         true,
			LandingReward:     2,
			LandingType:       0,
			Name:              "이것은 테스트 캠페인",
			OrganizationID:    0,
			OwnerID:           0,
			ProviderID:        0,
			PublishedAt:       "2006-01-02T15:04:06",
			Status:            model.StatusComplete,
			StartDate:         "2006-01-02T00:00:00",
			TargetAgeMin:      model.ESNullShortMin,
			TargetAgeMax:      model.ESNullShortMax,
			RegisteredDaysMin: model.ESNullShortMin,
			RegisteredDaysMax: model.ESNullShortMax,
			Timezone:          "Asia/Seoul",
			Title:             "abc",
			Tipu:              rand.Intn(100),
			Type:              "C",
			UpdatedAt:         "2006-01-02T15:04:05",
		},
		Related:       &id,
		Clicks:        10,
		CreativeLinks: map[string][]string{"A": {"http://abc-A.jpg"}, "R": {"http://abc-R.jpg"}},
		Impressions:   100,
		CreativeTypes: "A,R",
	}
	ccDoc := ecc.GetDocToCreate()
	test.AssertEqual(t, ccDoc.Categories, ecc.Categories, "TestESContentGetDoc - Categories")
	test.AssertEqual(t, ccDoc.TargetApp, model.ESGlobString, "TestESContentGetDoc - TargetApp")
	test.AssertEqual(t, ccDoc.TargetGender, model.ESGlobString, "TestESContentGetDoc - TargetGender")
	test.AssertEqual(t, ccDoc.TargetLanguage, model.ESGlobString, "TestESContentGetDoc - TargetLanguage")
	test.AssertEqual(t, ccDoc.TargetCarrier, model.ESGlobString, "TestESContentGetDoc - TargetCarrier")
	test.AssertEqual(t, ccDoc.TargetRegion, model.ESGlobString, "TestESContentGetDoc - TargetRegion")
	test.AssertEqual(t, ccDoc.CustomTarget1, model.ESGlobString, "TestESContentGetDoc - CustomTarget1")
	test.AssertEqual(t, ccDoc.CustomTarget2, model.ESGlobString, "TestESContentGetDoc - CustomTarget2")
	test.AssertEqual(t, ccDoc.CustomTarget3, model.ESGlobString, "TestESContentGetDoc - CustomTarget3")
	test.AssertEqual(t, ccDoc.WeekSlot, model.ESGlobString, "TestESContentGetDoc - WeekSlot")

	ecc2, err := ccDoc.EscapeNull()

	if err != nil {
		t.Fatal(err)
	}

	test.AssertEqual(t, ecc2.Categories, ecc.Categories, "TestESContentGetDoc - Categories")
	test.AssertEqual(t, ecc2.TargetApp, "", "TestESContentGetDoc - TargetApp")
	test.AssertEqual(t, ecc2.TargetGender, "", "TestESContentGetDoc - TargetGender")
	test.AssertEqual(t, ecc2.TargetLanguage, "", "TestESContentGetDoc - TargetLanguage")
	test.AssertEqual(t, ecc2.TargetCarrier, "", "TestESContentGetDoc - TargetCarrier")
	test.AssertEqual(t, ecc2.TargetRegion, "", "TestESContentGetDoc - TargetRegion")
	test.AssertEqual(t, ecc2.CustomTarget1, "", "TestESContentGetDoc - CustomTarget1")
	test.AssertEqual(t, ecc2.CustomTarget2, "", "TestESContentGetDoc - CustomTarget2")
	test.AssertEqual(t, ecc2.CustomTarget3, "", "TestESContentGetDoc - CustomTarget3")
	test.AssertEqual(t, ecc2.WeekSlot, "", "TestESContentGetDoc - WeekSlot")
}

func TestImpression(t *testing.T) {
	now := time.Now()

	contentID := rand.Int63n(now.Unix())
	country := "US"
	deviceID := rand.Int63n(now.Unix())
	year := now.Year()
	gender := "M"
	ifa := uuid.NewV1().String()
	unitDeviceToken := fmt.Sprintf("TestDevice-%d", rand.Int())
	unitID := int64(tests.HsKrAppID)

	caseImpression(t, &dto.ImpressionRequest{
		ImpressionData: impressiondata.ImpressionData{
			IFA:             ifa,
			CampaignID:      contentID,
			UnitID:          unitID,
			DeviceID:        deviceID,
			UnitDeviceToken: unitDeviceToken,
			Country:         country,
			Gender:          &gender,
			YearOfBirth:     &year,
		},
	})

	caseImpression(t, &dto.ImpressionRequest{
		ImpressionData: impressiondata.ImpressionData{
			IFA:             ifa,
			CampaignID:      contentID,
			UnitID:          unitID,
			DeviceID:        deviceID,
			UnitDeviceToken: unitDeviceToken,
			Country:         country,
		},
	})
}

func caseImpression(t *testing.T, impReq *dto.ImpressionRequest) {
	impressionURLStr := impReq.BuildImpressionURL()
	impURL, err := url.Parse(impressionURLStr)
	if err != nil {
		panic(err)
	}
	data := impURL.Query().Get("data")
	data, err = url.QueryUnescape(data)
	if err != nil {
		panic(err)
	}
	values := cypher.DecryptAesBase64Dict(data, model.APIAesKey, model.APIAesIv, true)

	t.Log(fmt.Sprintf("TestImpression() - impURL: %v\ndata: %v\nvalues: %v\nimpReq: %v", impURL, data, values, impReq))
}

func TestClickRedirectURL(t *testing.T) {
	now := time.Now()

	id := rand.Int63n(now.Unix())
	deviceID := rand.Int63n(now.Unix())
	ifa := uuid.NewV1().String()
	unitDeviceToken := fmt.Sprintf("TestDevice-%d", rand.Int())
	unitID := int64(tests.HsKrAppID)

	clickReq := &dto.ClickRequest{
		ClickURL:        "http://www.buzzvil.com",
		ClickURLClean:   "http://www.buzzvil-clean.com",
		DeviceID:        deviceID,
		ID:              id,
		IFA:             ifa,
		Name:            "TestName",
		OrganizationID:  1,
		Type:            model.CampaignTypeCast,
		Unit:            &app.Unit{ID: unitID, OrganizationID: 1},
		UnitDeviceToken: unitDeviceToken,
	}
	clickURLStr := clickReq.BuildClickRedirectURL()
	clickURL, err := url.Parse(clickURLStr)
	if err != nil {
		panic(err)
	}

	data := clickURL.Query()

	t.Log(fmt.Sprintf("TestClickRedirectURL() - clickURL: %v\ndata: %v\nclickReq: %v", clickURL, data, clickReq))

	campID, err := strconv.ParseInt(data.Get("campaign_id"), 10, 64)
	if err != nil {
		panic(err)
	}
	test.AssertEqual(t, campID, id, "TestClickRedirectURL")
	deviceID2, err := strconv.ParseInt(data.Get("device_id"), 10, 64)
	if err != nil {
		panic(err)
	}
	test.AssertEqual(t, deviceID2, deviceID, "TestClickRedirectURL")

	test.AssertEqual(t, data.Get("base_reward"), "__base_reward__", "TestClickRedirectURL")
	test.AssertEqual(t, data.Get("ifa"), ifa, "TestClickRedirectURL")
	test.AssertEqual(t, data.Get("campaign_type"), clickReq.Type, "TestClickRedirectURL")
	unitID2, err := strconv.ParseInt(data.Get("unit_id"), 10, 64)
	if err != nil {
		panic(err)
	}
	test.AssertEqual(t, unitID, unitID2, "TestClickRedirectURL")
	test.AssertEqual(t, data.Get("unit_device_token"), unitDeviceToken, "TestClickRedirectURL")
}

func TestContentCreativeImage(t *testing.T) {
	defaultATypeImage := "http://abc-A_720_1560.jpg"
	expectedATypeImageOnLowSDK := "http://abc-A_720_1230.jpg"
	RTypeImage := "http://abc-R.jpg"

	ecc := &dto.ESContentCampaign{

		CreativeLinks: map[string][]string{"A": {defaultATypeImage}, "R": {RTypeImage}},
		CreativeTypes: "A,R",
	}

	test.AssertEqual(t, ecc.GetCDNImageURL("A", 3900), defaultATypeImage, "TestContentCreativeImage")
	test.AssertEqual(t, ecc.GetCDNImageURL("A", 3899), expectedATypeImageOnLowSDK, "TestContentCreativeImage")

	test.AssertEqual(t, ecc.GetCDNImageURL("R", 3900), RTypeImage, "TestContentCreativeImage")
	test.AssertEqual(t, ecc.GetCDNImageURL("R", 3899), RTypeImage, "TestContentCreativeImage")

	// Test case where image name does not contain the size
	ATypeImage := "http://abc-A.jpg"
	eccImageNameWithoutSize := &dto.ESContentCampaign{

		CreativeLinks: map[string][]string{"A": {ATypeImage}, "R": {RTypeImage}},
		CreativeTypes: "A,R",
	}

	test.AssertEqual(t, eccImageNameWithoutSize.GetCDNImageURL("A", 3900), ATypeImage, "TestContentCreativeImage")
	test.AssertEqual(t, eccImageNameWithoutSize.GetCDNImageURL("A", 3899), ATypeImage, "TestContentCreativeImage")

	test.AssertEqual(t, eccImageNameWithoutSize.GetCDNImageURL("R", 3900), RTypeImage, "TestContentCreativeImage")
	test.AssertEqual(t, eccImageNameWithoutSize.GetCDNImageURL("R", 3899), RTypeImage, "TestContentCreativeImage")

}
