package controller_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/cypher"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/Buzzvil/go-test/test"
)

type (
	// TestAllocV1Response type definition
	TestAllocV1Response struct {
		Code      int                      `json:"code"`
		Message   string                   `json:"msg"`
		Campaigns []map[string]interface{} `json:"campaigns"`
		Settings  map[string]interface{}   `json:"settings"`
	}
)

func TestPostAllocationV1Settings(t *testing.T) {
	var allocRes TestAllocV1Response

	params := buildV1BaseTestRequest()

	statusCode, err := (&network.Request{
		Method: "POST",
		Params: params,
		URL:    ts.URL + "/api/allocation/",
	}).GetResponse(&allocRes)
	if err != nil || statusCode != 200 {
		t.Fatal(statusCode, err, allocRes)
	} else {
		fsr := allocRes.Settings["first_display_ratio"].(string)
		adRatio, _ := strconv.Atoi(strings.Split(fsr, ":")[0])
		test.AssertEqual(t, adRatio, 100, "TestPostAllocationV1Settings")
		checkSettings(t, allocRes)
	}

	params.Set("sdk_version", "1280")

	statusCode, err = (&network.Request{
		Method: "POST",
		Params: params,
		URL:    ts.URL + "/api/allocation/",
	}).GetResponse(&allocRes)
	if err != nil || statusCode != 200 {
		t.Fatal(statusCode, err, allocRes)
	} else {
		fsr := allocRes.Settings["first_display_ratio"].(string)
		adRatio, _ := strconv.Atoi(strings.Split(fsr, ":")[0])
		test.AssertEqual(t, adRatio, 9, "TestPostAllocationV1Settings")
		checkSettings(t, allocRes)
	}
}

func checkSettings(t *testing.T, allocRes TestAllocV1Response) {
	test.AssertEqual(t, int(allocRes.Settings["base_hour_limit"].(float64)) > 0, true, "checkSettings-base_hour_limit")
	test.AssertEqual(t, int(allocRes.Settings["base_init_period"].(float64)) > 0, true, "checkSettings-base_init_period")
	test.AssertEqual(t, int(allocRes.Settings["hour_limit"].(float64)) > 0, true, "checkSettings-hour_limit")
	test.AssertEqual(t, int(allocRes.Settings["page_limit"].(float64)) > 10, true, "checkSettings-page_limit")
	test.AssertEqual(t, int(allocRes.Settings["request_trigger"].(float64)) > 10, true, "checkSettings-request_trigger")
	test.AssertEqual(t, int(allocRes.Settings["request_period"].(float64)) > 1800, true, "checkSettings-request_period")
	test.AssertEqual(t, int(allocRes.Settings["base_init_period"].(float64)) > 1800, true, "checkSettings-base_init_period")
}

func TestPostAllocationV1AdOnly(t *testing.T) {
	_, removeFunc := getPatchedHttpClient(os.Getenv("BUZZAD_URL"))
	defer removeFunc()

	ccs := createBaseContentCampaigns(2)
	insertContentCampaignsToESAndDB(t, ccs...)
	defer deleteContentCampaignsFromESAndDB(t, ccs...)

	var allocRes TestAllocV1Response
	params := buildV1BaseTestRequest()
	params.Set("app_id", strconv.FormatInt(tests.TestAppIDAdOnly, 10))
	statusCode, err := (&network.Request{
		Method: "POST",
		Params: params,
		URL:    ts.URL + "/api/allocation/",
	}).GetResponse(&allocRes)
	if err != nil || statusCode != 200 {
		t.Fatal(statusCode, err, allocRes)
	} else {
		lenCamp := len(allocRes.Campaigns)
		test.AssertEqual(t, lenCamp > 0, true, fmt.Sprintf("TestPostAllocationV1AdOnly - %d", lenCamp))
		for _, camp := range allocRes.Campaigns {
			test.AssertEqual(t, camp["is_ad"].(bool), true, fmt.Sprintf("TestPostAllocationV1AdOnly - %v", camp["id"]))
		}
	}
}

func TestPostAllocationV1ContentOnly(t *testing.T) {
	_, removeFunc := getPatchedHttpClient(os.Getenv("BUZZAD_URL"))
	defer removeFunc()

	ccs := createBaseContentCampaigns(2)
	insertContentCampaignsToESAndDB(t, ccs...)
	defer deleteContentCampaignsFromESAndDB(t, ccs...)

	var allocRes TestAllocV1Response
	params := buildV1BaseTestRequest()
	params.Set("app_id", strconv.FormatInt(tests.TestAppIDContentOnly, 10))
	statusCode, err := (&network.Request{
		Method: "POST",
		Params: params,
		URL:    ts.URL + "/api/allocation/",
	}).GetResponse(&allocRes)
	if err != nil || statusCode != 200 {
		t.Fatal(statusCode, err, allocRes)
	} else {
		lenCamp := len(allocRes.Campaigns)
		test.AssertEqual(t, lenCamp > 0, true, fmt.Sprintf("TestPostAllocationV1ContentOnly - %d", lenCamp))
		for _, camp := range allocRes.Campaigns {
			test.AssertEqual(t, camp["is_ad"].(bool), false, fmt.Sprintf("TestPostAllocationV1ContentOnly - %v", camp["id"]))
		}
	}
}

func TestPostAllocationV1(t *testing.T) {
	_, removeFunc := getPatchedHttpClient(os.Getenv("BUZZAD_URL"))
	defer removeFunc()

	var allocRes TestAllocV1Response

	params := buildV1BaseTestRequest()

	appUseCase := buzzscreen.Service.AppUseCase

	ctx := context.Background()

	appID, err := strconv.ParseInt(params.Get("app_id"), 10, 64)
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
	statusCode, err := (&network.Request{
		Method: "POST",
		Params: params,
		URL:    ts.URL + "/api/allocation/",
	}).GetResponse(&allocRes)

	t.Log("TestPostAllocationV1() - response code:", allocRes.Code)

	if err != nil || statusCode != 200 {
		t.Fatal(statusCode, err, allocRes)
	} else {
		lenCamp := len(allocRes.Campaigns)
		test.AssertEqual(t, lenCamp > 0, true, fmt.Sprintf("TestPostAllocationV1 - %v", lenCamp))
		if len(allocRes.Campaigns) != 0 {
			for _, camp := range allocRes.Campaigns {
				if val, ok := camp["sdk_version_to"]; ok && val != 0 {
					t.Error("TestPostAllocationV1() - SDK Targeting shouldn't be provided.")
				}
				if _, ok := camp["clean_mode"]; !ok {
					t.Fatal("TestPostAllocationV1() - clean mode is Empty.")
				}

				if impressionURLs, ok := camp["impression_urls"]; ok {
					if camp["is_ad"].(bool) == false {
						impURL, err := url.Parse(impressionURLs.([]interface{})[0].(string))
						if err != nil {
							panic(err)
						}
						data := impURL.Query().Get("data")
						data, err = url.QueryUnescape(data)
						if err != nil {
							panic(err)
						}

						values := cypher.DecryptAesBase64Dict(data, model.APIAesKey, model.APIAesIv, true)
						t.Logf("TestPostAllocationV1() - impURL: %v, data: %v, values: %v", impURL, data, values)

						u, err := appUseCase.GetUnitByAppIDAndType(ctx, appID, app.UnitTypeLockscreen)
						if err != nil {
							panic(err)
						}
						test.AssertEqual(t, values["c"].(float64), camp["id"].(float64), "TestImpression")
						test.AssertEqual(t, values["cou"].(string), country, "TestImpression")
						test.AssertEqual(t, int(values["d"].(float64)), deviceID, "TestImpression")
						test.AssertEqual(t, values["sex"].(string), gender, "TestImpression")
						test.AssertEqual(t, values["i"].(string), ifa, "TestImpression")
						test.AssertEqual(t, values["udt"].(string), unitDeviceToken, "TestImpression")
						test.AssertEqual(t, int64(values["u"].(float64)), u.ID, "TestImpression")
						test.AssertEqual(t, int(values["yob"].(float64)), yob, "TestImpression")
					}
				} else {
					t.Fatal("TestPostAllocationV1() - impressionURLs is Empty.")
				}

				if clickURLStr, ok := camp["click_url"]; ok {
					clickURL, err := url.Parse(clickURLStr.(string))
					if err != nil {
						panic(err)
					}
					data := clickURL.Query()

					campID, err := strconv.ParseInt(data.Get("campaign_id"), 10, 64)
					if err != nil {
						panic(err)
					}
					test.AssertEqual(t, campID == 0, false, "TestClickRedirectURL")
					deviceIDFromData, err := strconv.Atoi(data.Get("device_id"))
					if err != nil {
						panic(err)
					}
					test.AssertEqual(t, deviceIDFromData, deviceID, "TestClickRedirectURL")
					test.AssertEqual(t, data.Get("base_reward"), "__base_reward__", "TestClickRedirectURL")
					test.AssertEqual(t, data.Get("ifa"), ifa, "TestClickRedirectURL")
					test.AssertEqual(t, data.Get("campaign_type"), camp["type"].(string), "TestClickRedirectURL")
					unitIDFromData, err := strconv.ParseInt(data.Get("unit_id"), 10, 64)
					if err != nil {
						panic(err)
					}
					u, err := appUseCase.GetUnitByAppIDAndType(ctx, appID, app.UnitTypeLockscreen)
					if err != nil {
						panic(err)
					}
					test.AssertEqual(t, unitIDFromData, u.ID, "TestClickRedirectURL")
					test.AssertEqual(t, data.Get("unit_device_token"), unitDeviceToken, "TestClickRedirectURL")
					if data.Get("tracking_url") != "" {
						test.AssertEqual(t, data.Get("tracking_url") != "", true, "TestTrackingURL")
					}
					t.Logf("TestPostAllocationV1() - clickURL: %v, data: %v", clickURL, data)
				} else {
					t.Fatal("TestPostAllocationV1() - clickURL is Empty.")
				}
			}
		}
	}
}
