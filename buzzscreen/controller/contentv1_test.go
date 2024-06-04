package controller_test

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"

	"math"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/controller"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/cypher"
	"github.com/Buzzvil/go-test/test"
	"gopkg.in/olivere/elastic.v5"
)

func TestV1ContentAllocation(t *testing.T) {
	ccs := createBaseContentCampaigns(2)

	insertContentCampaignsToESAndDB(t, ccs...)
	defer deleteContentCampaignsFromESAndDB(t, ccs...)

	newContentV1testCase(t, "TestContentAllocationV1", func(testCase *ContentV1TestCase) {
	}).run(func(tc *ContentV1TestCase, res *TestContentV1Response) bool {
		for i := range ccs {
			compareESCampaignWithResCampaign(tc, ccs[i], res.Campaigns[i])
		}
		return len(res.Campaigns) == 2
	})
}

func compareESCampaignWithResCampaign(tc *ContentV1TestCase, esCamp *dto.ESContentCampaign, resCamp *dto.CampaignV1) {
	params := tc.params
	appID, err := strconv.ParseInt(params.Get("app_id"), 10, 64)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	appUseCase := buzzscreen.Service.AppUseCase
	unit, err := appUseCase.GetUnitByAppIDAndType(ctx, appID, app.UnitTypeLockscreen)
	if err != nil {
		panic(err)
	}
	ifa := params.Get("ifa")
	unitDeviceToken := params.Get("unit_device_token")
	yob, err := strconv.Atoi(params.Get("year_of_birth"))
	if err != nil {
		panic(err)
	}
	gender := params.Get("sex")
	deviceID, err := strconv.Atoi(params.Get("device_id"))
	if err != nil {
		panic(err)
	}
	country := params.Get("country")

	test.AssertEqual(tc.t, esCamp.Categories, resCamp.Category, "TestContentAllocationV1 - Category")
	test.AssertEqual(tc.t, resCamp.Extra["ic_type"].(string), "VIEW", "TestContentAllocationV1 - Extra")

	{ //ImpURLs
		impURL, err := url.Parse(resCamp.ImpressionURLs[0])
		if err != nil {
			panic(err)
		}
		data := impURL.Query().Get("data")
		data, err = url.QueryUnescape(data)
		if err != nil {
			panic(err)
		}

		values := cypher.DecryptAesBase64Dict(data, model.APIAesKey, model.APIAesIv, true)
		tc.t.Logf("compareESCampaignWithResCampaign - impressionURL impURL: %v, data: %v, values: %v", impURL, data, values)

		test.AssertEqual(tc.t, int64(values["c"].(float64)), resCamp.ID, "compareESCampaignWithResCampaign - ImpressionURLs")
		test.AssertEqual(tc.t, values["cou"].(string), country, "compareESCampaignWithResCampaign - ImpressionURLs")
		test.AssertEqual(tc.t, int(values["d"].(float64)), deviceID, "compareESCampaignWithResCampaign - ImpressionURLs")
		test.AssertEqual(tc.t, values["sex"].(string), gender, "compareESCampaignWithResCampaign - ImpressionURLs")
		test.AssertEqual(tc.t, values["i"].(string), ifa, "compareESCampaignWithResCampaign - ImpressionURLs")
		test.AssertEqual(tc.t, values["udt"].(string), unitDeviceToken, "compareESCampaignWithResCampaign - ImpressionURLs")
		test.AssertEqual(tc.t, int64(values["u"].(float64)), unit.ID, "compareESCampaignWithResCampaign - ImpressionURLs")
		test.AssertEqual(tc.t, int(values["yob"].(float64)), yob, "compareESCampaignWithResCampaign - ImpressionURLs")
	}

	{ //ClickURL
		clickURL, err := url.Parse(resCamp.ClickURL)
		if err != nil {
			panic(err)
		}
		data := clickURL.Query()

		deviceIDFromData, err := strconv.Atoi(data.Get("device_id"))
		if err != nil {
			panic(err)
		}
		test.AssertEqual(tc.t, deviceIDFromData, deviceID, "compareESCampaignWithResCampaign - ClickURL")
		test.AssertEqual(tc.t, data.Get("base_reward"), "__base_reward__", "compareESCampaignWithResCampaign - ClickURL")
		test.AssertEqual(tc.t, data.Get("ifa"), ifa, "compareESCampaignWithResCampaign - ClickURL")
		test.AssertEqual(tc.t, data.Get("campaign_type"), resCamp.Type, "compareESCampaignWithResCampaign - ClickURL")
		unitIDFromData, err := strconv.ParseInt(data.Get("unit_id"), 10, 64)
		if err != nil {
			panic(err)
		}
		test.AssertEqual(tc.t, unitIDFromData, unit.ID, "compareESCampaignWithResCampaign - ClickURL")
		test.AssertEqual(tc.t, data.Get("unit_device_token"), unitDeviceToken, "compareESCampaignWithResCampaign - ClickURL")
		tc.t.Logf("compareESCampaignWithResCampaign() - clickURL: %v, data: %v", clickURL, data)
	}

	test.AssertEqual(tc.t, esCamp.OrganizationID == unit.OrganizationID, resCamp.IsMedia, "compareESCampaignWithResCampaign - IsMedia")

	test.AssertEqual(tc.t, unit.BaseReward, resCamp.BaseReward, "compareESCampaignWithResCampaign - BaseReward")
	test.AssertEqual(tc.t, *esCamp.ChannelID, *resCamp.ChannelID, "compareESCampaignWithResCampaign - ChannelID")

	test.AssertEqual(tc.t, *esCamp.Channel, *resCamp.Channel, "compareESCampaignWithResCampaign - Channel")

	test.AssertEqual(tc.t, esCamp.CleanLink, resCamp.ClickURLClean, "compareESCampaignWithResCampaign - ClickURLClean")

	test.AssertEqual(tc.t, esCamp.DisplayType, resCamp.DisplayType, "compareESCampaignWithResCampaign - DisplayType")
	test.AssertEqual(tc.t, utils.ConvertToUnixTime(esCamp.EndDate), resCamp.EndedAt, "compareESCampaignWithResCampaign - EndedAt")

	test.AssertEqual(tc.t, int(math.Min(float64(esCamp.DisplayWeight*model.DisplayWeightMultiplier), 10000000)), int(resCamp.FirstDisplayWeight), "compareESCampaignWithResCampaign - FirstDisplayPriority")
	test.AssertEqual(tc.t, 10, int(resCamp.FirstDisplayPriority), "compareESCampaignWithResCampaign - FirstDisplayPriority")
	test.AssertEqual(tc.t, esCamp.ID, resCamp.ID, "compareESCampaignWithResCampaign - ID")
	test.AssertEqual(tc.t, esCamp.CreativeLinks[unit.Platform][0], resCamp.Image, "compareESCampaignWithResCampaign - Image")
	test.AssertEqual(tc.t, false, resCamp.IsAd, "compareESCampaignWithResCampaign - IsAd")

	test.AssertEqual(tc.t, controller.LandingTypeResponseMapping[esCamp.LandingType], resCamp.LandingType, "compareESCampaignWithResCampaign - LandingReward")
	test.AssertEqual(tc.t, esCamp.LandingReward, resCamp.LandingReward, "compareESCampaignWithResCampaign - LandingReward")
	test.AssertEqual(tc.t, esCamp.Name, resCamp.Name, "compareESCampaignWithResCampaign - Name")
	test.AssertEqual(tc.t, esCamp.OrganizationID, resCamp.OrganizationID, "compareESCampaignWithResCampaign - OrganizationId")
	test.AssertEqual(tc.t, esCamp.OwnerID, resCamp.OwnerID, "compareESCampaignWithResCampaign - OwnerId")
	test.AssertEqual(tc.t, esCamp.ProviderID, resCamp.ProviderID, "compareESCampaignWithResCampaign - ProviderId")
	test.AssertEqual(tc.t, esCamp.ClickURL, resCamp.SourceURL, "compareESCampaignWithResCampaign - SourceURL")
	test.AssertEqual(tc.t, utils.ConvertToUnixTime(esCamp.StartDate), resCamp.StartedAt, "compareESCampaignWithResCampaign - StartedAt")
	test.AssertEqual(tc.t, true, resCamp.SupportWebp, "compareESCampaignWithResCampaign - SupportWebp")
	test.AssertEqual(tc.t, esCamp.Timezone, resCamp.Timezone, "compareESCampaignWithResCampaign - Timezone")
	if esCamp.TargetApp != "" {
		test.AssertEqual(tc.t, esCamp.TargetApp, resCamp.TargetApp, "compareESCampaignWithResCampaign - TargetApp")
	}
	test.AssertEqual(tc.t, model.CampaignTypeCast, resCamp.Type, "compareESCampaignWithResCampaign - Type")
	test.AssertEqual(tc.t, 0, resCamp.UnitPrice, "compareESCampaignWithResCampaign - UnitPrice")
}

func newContentV1testCase(t *testing.T, name string, builder func(*ContentV1TestCase)) *ContentV1TestCase {
	ctc := &ContentV1TestCase{
		t:      t,
		name:   name,
		params: buildV1BaseTestRequest(),
	}

	if builder != nil {
		builder(ctc)
	}

	return ctc
}

type (
	// ContentV1TestCase type definition
	ContentV1TestCase struct {
		t      *testing.T
		name   string
		params *url.Values
	}
)

type (
	// TestContentV1Response type definition
	TestContentV1Response struct {
		dto.ContentAllocV1Response
	}
)

func (tc *ContentV1TestCase) run(equalFunc func(tc *ContentV1TestCase, res *TestContentV1Response) bool) {
	var contentV1Res TestContentV1Response
	t := tc.t
	statusCode, err := (&network.Request{
		Method: "POST",
		Params: tc.params,
		URL:    ts.URL + "/api/content_allocation/",
	}).GetResponse(&contentV1Res)

	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if err != nil || statusCode != 200 {
		t.Fatal(statusCode, err, contentV1Res)
	} else {
		test.AssertEqual(t, equalFunc(tc, &contentV1Res), true, fmt.Sprintf("ContentTestCase - %v", tc.name))
	}
}

func TestElasticComputeDeltaDays(t *testing.T) {
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)
	allocReq := dto.ContentAllocV1Request{
		RegisteredSecondsReq: yesterday.Unix(),
	}
	deltaDays := utils.GetDaysFrom(allocReq.GetRegisteredSeconds())

	t.Log(fmt.Sprintf("TestElasticSearch() - today.Unix(): %v, yesterday.Unix(): %v, deltaDays: %v", today.Unix(), yesterday.Unix(), deltaDays))
	test.AssertEqual(t, deltaDays, 1, fmt.Sprintf("TestElasticSearch.execComputeDeltaDays() - %v", deltaDays))
}

func TestElasticElasticQueries(t *testing.T) {
	allocReq := dto.ContentAllocV1Request{
		CustomTarget1: "A,B,C",
		CustomTarget2: "D",
		Language:      "en,kr",
	}
	customTargetMap := map[string]string{
		"custom_target_1": allocReq.CustomTarget1,
		"custom_target_2": allocReq.CustomTarget2,
		"custom_target_3": allocReq.CustomTarget3,
	}
	for key, targets := range customTargetMap {
		customTargetQueries := make([]*elastic.TermQuery, 0)
		for _, target := range strings.Split(targets, ",") {
			if target != "" {
				customTargetQueries = append(customTargetQueries, elastic.NewTermQuery(key, target))
			}
		}
		t.Log(fmt.Sprintf("TestElasticSearch().execElasticQueries() - key: %v, targets: %v, customTargetQueries: %v", key, targets, customTargetQueries))
	}

	languageQueries := make([]elastic.Query, 0)
	for _, lang := range strings.Split(allocReq.Language, ",") {
		languageQueries = append(languageQueries, elastic.NewTermQuery("language", lang))
	}

	t.Log(fmt.Sprintf("TestElasticSearch.execElasticQueries() - languageQueries: %v", languageQueries))
	test.AssertEqual(t, len(languageQueries), 2, "TestElasticSearch.execElasticQueries() - languageQueries: "+allocReq.Language)

	allocReq.Language = "en"
	languageQueries = make([]elastic.Query, 0)
	for _, lang := range strings.Split(allocReq.Language, ",") {
		languageQueries = append(languageQueries, elastic.NewTermQuery("language", lang))
	}
	test.AssertEqual(t, len(languageQueries), 1, "TestElasticSearch.execElasticQueries() - languageQueries: "+allocReq.Language)
}

func TestGetEntityScores(t *testing.T) {
	// sample data
	deviceID := int64(123)
	entityScores := map[string]float64{"a": 0.1, "b": 0.2}

	// Insert entity profile for good device 123
	dp := device.Profile{ID: deviceID, EntityScores: &entityScores}
	buzzscreen.Service.DeviceUseCase.SaveProfile(dp)

	allocReq := dto.ContentAllocV1Request{DeviceID: deviceID}
	es := *allocReq.GetEntityScores()
	test.AssertEqual(t, reflect.DeepEqual(entityScores, es), true, "TestGetEntityScores - Wrong device profile found!")

	allocReq = dto.ContentAllocV1Request{DeviceID: int64(456)}
	test.AssertEqual(t, allocReq.GetEntityScores() == nil, true, "TestGetEntityScores - Entity profile should not be found!")
}
