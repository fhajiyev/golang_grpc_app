package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	rewardRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/reward/repo"
	"github.com/Buzzvil/go-test/mock"
	"github.com/Buzzvil/go-test/test"
	"github.com/bxcodec/faker"
	"github.com/guregu/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAds(t *testing.T) {
	newAdTestCase(t, &(BuzzAdMockRes{Types: []string{"WEB", "JS"}})).
		run(func(oriAd ad.AdV2, resAd map[string]interface{}) bool {
			return oriAd.Creative["htmlTag"] == resAd["creative"].(map[string]interface{})["html_tag"] &&
				oriAd.Creative["height"].(int) == int(resAd["creative"].(map[string]interface{})["height"].(float64)) &&
				oriAd.Creative["width"].(int) == int(resAd["creative"].(map[string]interface{})["width"].(float64)) &&
				oriAd.Creative["sizeType"] == resAd["creative"].(map[string]interface{})["size_type"] &&
				oriAd.Creative["bgUrl"] == resAd["creative"].(map[string]interface{})["bg_url"]
		})
	newAdTestCase(t, &(BuzzAdMockRes{Types: []string{"SDK"}})).
		run(func(oriAd ad.AdV2, resAd map[string]interface{}) bool {
			return oriAd.Creative["network"] == resAd["creative"].(map[string]interface{})["network"] &&
				oriAd.Creative["placementId"] == resAd["creative"].(map[string]interface{})["placement_id"] &&
				oriAd.Creative["publisherId"] == resAd["creative"].(map[string]interface{})["publisher_id"] &&
				oriAd.Creative["referrerUrl"] == resAd["creative"].(map[string]interface{})["referrer_url"]
		})
}

func TestGetAdsLandingRewardSdk(t *testing.T) {
	testCase := &AdTestCase{AdTestConfig: V3AdTestConfig{}}
	testCase.setUp()
	defer testCase.tearDown()
	originalClickURL := "http://test/click.url"

	landingReward := 100
	eventLandingReward := 900
	actionReward := 200
	noReward := 0
	trackingURL := "http://buzzvil.com"

	ttl := 3600
	cases := []struct {
		Description string
		BAResponse  []ad.AdV2
		BSResponse  map[int][]map[string]interface{}
	}{
		{
			Description: "1. No events",
			BAResponse: []ad.AdV2{
				{
					ID: 1, // 1000000001
					Creative: map[string]interface{}{
						"bgUrl":       "https://buzzvil.akamaized.net/adfit.image/uploads/1508230314-ZI5BO.jpg",
						"landingType": 3,
						"network":     "OUTBRAIN",
						"placementId": "SDK_3",
						"publisherId": "SLIDE1L0QG0CFL3J7BP1L93J6",
						"referrerUrl": "http://www.getslidejoy.com/japan",
						"type":        "SDK",
						"clickUrl":    originalClickURL,
					},
					LandingReward: landingReward,
					ActionReward:  actionReward,
					Events:        ad.Events{},
					Meta:          map[string]interface{}{},
					TTL:           &ttl,
				},
			},
			BSResponse: map[int][]map[string]interface{}{
				20000: {{
					"id":             int64(1000000001),
					"landing_reward": landingReward,
					"reward_status":  "unknown",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000001&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=cae3915a-4732-40b8-9c99-50819e9a1f49&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&unit_device_token=TestUser-47&unit_id=1&use_clean_mode=__use_clean_mode__",
					},
					"meta": map[string]interface{}{},
				}},
				30302: {{
					"id":             int64(1000000001),
					"landing_reward": landingReward,
					"reward_status":  "unknown",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000001&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=cae3915a-4732-40b8-9c99-50819e9a1f49&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&unit_device_token=TestUser-47&unit_id=1&use_clean_mode=__use_clean_mode__",
					},
					"meta": map[string]interface{}{},
				}},
				30005: {{
					"id":             int64(1000000001),
					"landing_reward": landingReward,
					"reward_status":  "unknown",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000001&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=cae3915a-4732-40b8-9c99-50819e9a1f49&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&unit_device_token=TestUser-47&unit_id=1&use_clean_mode=__use_clean_mode__",
					},
					"meta": map[string]interface{}{},
				}},
			},
		},
		{
			Description: "2. Event exists(click reward)",
			BAResponse: []ad.AdV2{
				{
					ID: 1,
					Creative: map[string]interface{}{
						"bgUrl":       "https://buzzvil.akamaized.net/adfit.image/uploads/1508230314-ZI5BO.jpg",
						"landingType": 3,
						"network":     "OUTBRAIN",
						"placementId": "SDK_3",
						"publisherId": "SLIDE1L0QG0CFL3J7BP1L93J6",
						"referrerUrl": "http://www.getslidejoy.com/japan",
						"type":        "SDK",
						"clickUrl":    originalClickURL,
					},
					LandingReward: landingReward,
					ActionReward:  actionReward,
					Events: ad.Events{
						{
							Type:         "landed",
							TrackingURLs: []string{trackingURL},
							Reward: &ad.Reward{
								Amount: int64(eventLandingReward),
								Status: event.StatusReceivable,
								TTL:    86400,
								Extra: map[string]string{
									"minimum_stay_duration": "0",
								},
							},
							Extra: map[string]string{},
						},
					},
					Meta: map[string]interface{}{},
					TTL:  &ttl,
				},
			},
			BSResponse: map[int][]map[string]interface{}{
				20000: {{
					"id":             int64(1000000001),
					"landing_reward": eventLandingReward,
					"reward_status":  "unknown",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000002&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=18fb619f-8555-45c3-a919-1b644de3dbd6&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&tracking_url=http%3A%2F%2Fbuzzvil.com&unit_device_token=TestUser-13&unit_id=1&use_clean_mode=__use_clean_mode__&use_reward_api=true",
					},
					"meta": map[string]interface{}{},
				}},
				30302: {{
					"id":             int64(1000000001),
					"landing_reward": eventLandingReward,
					"reward_status":  "unknown",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000002&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=18fb619f-8555-45c3-a919-1b644de3dbd6&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&tracking_url=http%3A%2F%2Fbuzzvil.com&unit_device_token=TestUser-13&unit_id=1&use_clean_mode=__use_clean_mode__&use_reward_api=true",
					},
					"meta": map[string]interface{}{},
				}},
				30005: {{
					"id":             int64(1000000001),
					"landing_reward": eventLandingReward,
					"reward_status":  "unknown",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000002&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=18fb619f-8555-45c3-a919-1b644de3dbd6&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&tracking_url=http%3A%2F%2Fbuzzvil.com&unit_device_token=TestUser-13&unit_id=1&use_clean_mode=__use_clean_mode__&use_reward_api=true",
					},
					"meta": map[string]interface{}{},
				}},
			},
		},
		{
			Description: "3. Event exists(landing reward)",
			BAResponse: []ad.AdV2{
				{
					ID: 1,
					Creative: map[string]interface{}{
						"bgUrl":       "https://buzzvil.akamaized.net/adfit.image/uploads/1508230314-ZI5BO.jpg",
						"landingType": 3,
						"network":     "OUTBRAIN",
						"placementId": "SDK_3",
						"publisherId": "SLIDE1L0QG0CFL3J7BP1L93J6",
						"referrerUrl": "http://www.getslidejoy.com/japan",
						"type":        "SDK",
						"clickUrl":    originalClickURL,
					},
					LandingReward: landingReward,
					ActionReward:  actionReward,
					Events: ad.Events{
						{
							Type:         "landed",
							TrackingURLs: []string{trackingURL},
							Reward: &ad.Reward{
								Amount:         int64(eventLandingReward),
								Status:         event.StatusReceivable,
								StatusCheckURL: "http://statuscheckurl.com",
								TTL:            86400,
								Extra:          map[string]string{"minimum_stay_duration": "10"},
							},
							Extra: map[string]string{},
						},
					},
					Meta: map[string]interface{}{},
					TTL:  &ttl,
				},
			},
			BSResponse: map[int][]map[string]interface{}{
				20000: {{
					"id":             int64(1000000001),
					"landing_reward": eventLandingReward,
					"reward_status":  "unknown",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000002&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=18fb619f-8555-45c3-a919-1b644de3dbd6&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&tracking_url=http%3A%2F%2Fbuzzvil.com&unit_device_token=TestUser-13&unit_id=1&use_clean_mode=__use_clean_mode__&use_reward_api=true",
					},
					"meta": map[string]interface{}{},
				}},
				30302: {{
					"id":             int64(1000000001),
					"landing_reward": eventLandingReward,
					"reward_status":  "unknown",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000002&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=18fb619f-8555-45c3-a919-1b644de3dbd6&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&tracking_url=http%3A%2F%2Fbuzzvil.com&unit_device_token=TestUser-13&unit_id=1&use_clean_mode=__use_clean_mode__&use_reward_api=true",
					},
					"meta": map[string]interface{}{},
				}},
				30005: {{
					"id":             int64(1000000001),
					"landing_reward": noReward,
					"reward_status":  "unknown",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000001&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=cae3915a-4732-40b8-9c99-50819e9a1f49&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&unit_device_token=TestUser-47&unit_id=1&use_clean_mode=__use_clean_mode__",
					},
					"meta": map[string]interface{}{},
					"events": []interface{}{map[string]interface{}{
						"event_type": "landed",
						"reward": map[string]interface{}{
							"amount":           float64(900),
							"extra":            map[string]interface{}{"minimum_stay_duration": "10"},
							"issue_method":     "",
							"status":           "receivable",
							"status_check_url": "http://statuscheckurl.com",
							"ttl":              float64(86400)},
						"tracking_urls": []interface{}{"http://buzzvil.com"},
					}},
				}},
			},
		},
		{
			Description: "4. Event exists(landing reward already received)",
			BAResponse: []ad.AdV2{
				{
					ID: 1,
					Creative: map[string]interface{}{
						"bgUrl":       "https://buzzvil.akamaized.net/adfit.image/uploads/1508230314-ZI5BO.jpg",
						"landingType": 3,
						"network":     "OUTBRAIN",
						"placementId": "SDK_3",
						"publisherId": "SLIDE1L0QG0CFL3J7BP1L93J6",
						"referrerUrl": "http://www.getslidejoy.com/japan",
						"type":        "SDK",
						"clickUrl":    originalClickURL,
					},
					LandingReward: landingReward,
					ActionReward:  actionReward,
					Events: ad.Events{
						{
							Type:         "landed",
							TrackingURLs: []string{},
							Reward: &ad.Reward{
								Amount: int64(noReward),
								Status: event.StatusAlreadyReceived,
								TTL:    86400,
								Extra: map[string]string{
									"minimum_stay_duration": "10",
								},
							},
							Extra: map[string]string{},
						},
					},
					Meta: map[string]interface{}{},
					TTL:  &ttl,
				},
			},
			BSResponse: map[int][]map[string]interface{}{
				20000: {{
					"id":             int64(1000000001),
					"landing_reward": noReward,
					"reward_status":  "received",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000001&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=cae3915a-4732-40b8-9c99-50819e9a1f49&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&unit_device_token=TestUser-47&unit_id=1&use_clean_mode=__use_clean_mode__",
					},
					"meta": map[string]interface{}{},
				}},
				30302: {{
					"id":             int64(1000000001),
					"landing_reward": noReward,
					"reward_status":  "received",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000001&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=cae3915a-4732-40b8-9c99-50819e9a1f49&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&unit_device_token=TestUser-47&unit_id=1&use_clean_mode=__use_clean_mode__",
					},
					"meta": map[string]interface{}{},
				}},
				30005: {{
					"id":             int64(1000000001),
					"landing_reward": noReward,
					"reward_status":  "received",
					"creative": map[string]interface{}{
						"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000001&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=cae3915a-4732-40b8-9c99-50819e9a1f49&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&unit_device_token=TestUser-47&unit_id=1&use_clean_mode=__use_clean_mode__",
					},
					"meta": map[string]interface{}{},
					"events": []interface{}{map[string]interface{}{
						"event_type": "landed",
						"reward": map[string]interface{}{
							"amount":           float64(0),
							"extra":            map[string]interface{}{"minimum_stay_duration": "10"},
							"issue_method":     "",
							"status":           "already_received",
							"status_check_url": "",
							"ttl":              float64(86400)},
						"tracking_urls": []interface{}{},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		testCase.MockBuzzAds(c.BAResponse)

		for sdkVersion, expects := range c.BSResponse {
			requestParams, _ := testCase.getRequestParams(t)
			requestParams.Set("sdk_version", strconv.Itoa(sdkVersion))

			response, statusCode, err := testCase.GetAdsResponse(t, requestParams)
			require.Nil(t, err)
			require.Equal(t, statusCode, 200)

			assert.NotEmpty(t, response.Result.Ads)
			assert.Equal(t, len(expects), len(response.Result.Ads))

			expect := expects[0]
			actual := response.Result.Ads[0]

			sdkVersionMessage := fmt.Sprintf("description: %s sdkversion: %v", c.Description, sdkVersion)
			assert.Equal(t, expect["id"].(int64), int64(actual["id"].(float64)), "%+v does not equal %v", expect, actual)
			assert.Equal(t, expect["landing_reward"].(int), int(actual["landing_reward"].(float64)), fmt.Sprintf("%s wrong landing reward amount ", sdkVersionMessage))
			assert.Equal(t, expect["reward_status"].(string), actual["reward_status"].(string), fmt.Sprintf("%s wrong reward_status ", sdkVersionMessage))
			assert.Equal(t, expect["meta"].(map[string]interface{}), actual["meta"].(map[string]interface{}), fmt.Sprintf("%s wrong meta ", sdkVersionMessage))
			if expect["events"] != nil {
				assert.Equal(t, expect["events"], actual["events"], fmt.Sprintf("%s wrong events ", sdkVersionMessage))
			} else {
				assert.Nil(t, actual["events"], fmt.Sprintf("%s wrong events ", sdkVersionMessage))
			}

			expect_click_url, _ := url.Parse(expect["creative"].(map[string]interface{})["click_url"].(string))
			received_click_url := actual["creative"].(map[string]interface{})["click_url"].(string)
			actual_click_url, _ := url.Parse(received_click_url)
			assert.Equal(t, expect_click_url.Query().Get("use_reward_api"), actual_click_url.Query().Get("use_reward_api"), fmt.Sprintf("%s wrong use_reward_api %s", sdkVersionMessage, received_click_url))
			assert.Equal(t, expect_click_url.Query().Get("tracking_url"), actual_click_url.Query().Get("tracking_url"), fmt.Sprintf("%s wrong tracking_url %s", sdkVersionMessage, received_click_url))
		}

		testCase.buzzAdServer.ResponseHandlers = nil
	}
}

func TestGetAdsFiltered(t *testing.T) {
	testCase := &AdTestCase{AdTestConfig: V3AdTestConfig{}}
	testCase.setUp()
	defer testCase.tearDown()
	originalClickURL := "http://test/click.url"

	landingReward := 100
	eventLandingReward := 900
	actionReward := 200
	trackingURL := "http://buzzvil.com"

	url2048 := "http://"
	for i := 0; i < 2044; i++ {
		url2048 += "t"
	}
	url2048 += ".com"

	ttl := 3600
	cases := []struct {
		Description string
		Request     map[string]interface{}
		BAResponse  []ad.AdV2
		BSResponse  []map[string]interface{}
	}{
		{
			Description: "1. os version <= 20 with clickURL > 2048 should filter out ad",
			Request: map[string]interface{}{
				"os_version":  "20",
				"sdk_version": 30005,
			},
			BAResponse: []ad.AdV2{
				{
					ID: 1,
					Creative: map[string]interface{}{
						"bgUrl":       "https://buzzvil.akamaized.net/adfit.image/uploads/1508230314-ZI5BO.jpg",
						"landingType": 3,
						"network":     "OUTBRAIN",
						"placementId": "SDK_3",
						"publisherId": "SLIDE1L0QG0CFL3J7BP1L93J6",
						"referrerUrl": "http://www.getslidejoy.com/japan",
						"type":        "SDK",
						"clickUrl":    originalClickURL,
					},
					LandingReward: landingReward,
					ActionReward:  actionReward,
					Events: ad.Events{
						{
							Type:         "landed",
							TrackingURLs: []string{url2048},
							Reward: &ad.Reward{
								Amount: int64(eventLandingReward),
								Status: event.StatusReceivable,
								TTL:    86400,
								Extra: map[string]string{
									"minimum_stay_duration": "0",
								},
							},
							Extra: map[string]string{},
						},
					},
					Meta: map[string]interface{}{},
					TTL:  &ttl,
				},
			},
			BSResponse: []map[string]interface{}{},
		},
		{
			Description: "2. os version > 20 with clickURL > 2048 should not filter out ad",
			Request: map[string]interface{}{
				"os_version":  "30",
				"sdk_version": 30005,
			},
			BAResponse: []ad.AdV2{
				{
					ID: 1,
					Creative: map[string]interface{}{
						"bgUrl":       "https://buzzvil.akamaized.net/adfit.image/uploads/1508230314-ZI5BO.jpg",
						"landingType": 3,
						"network":     "OUTBRAIN",
						"placementId": "SDK_3",
						"publisherId": "SLIDE1L0QG0CFL3J7BP1L93J6",
						"referrerUrl": "http://www.getslidejoy.com/japan",
						"type":        "SDK",
						"clickUrl":    originalClickURL,
					},
					LandingReward: landingReward,
					ActionReward:  actionReward,
					Events: ad.Events{
						{
							Type:         "landed",
							TrackingURLs: []string{trackingURL},
							Reward: &ad.Reward{
								Amount: int64(eventLandingReward),
								Status: event.StatusReceivable,
								TTL:    86400,
								Extra: map[string]string{
									"minimum_stay_duration": "0",
								},
							},
							Extra: map[string]string{},
						},
					},
					Meta: map[string]interface{}{},
					TTL:  &ttl,
				},
			},
			BSResponse: []map[string]interface{}{{
				"id":             int64(1000000001),
				"landing_reward": eventLandingReward,
				"reward_status":  "unknown",
				"creative": map[string]interface{}{
					"click_url": "https://localhost/api/click_redirect/?app_id=100000043&base_reward=__base_reward__&campaign_id=2000000002&campaign_is_media=0&campaign_name=&campaign_owner_id=&campaign_payload=__campaign_payload__&campaign_type=&check=__check__&client_unit_device_token=__unit_device_token__&device_id=1&ifa=18fb619f-8555-45c3-a919-1b644de3dbd6&place=__place__&position=__position__&redirect_url=http%3A%2F%2Ftest%2Fclick.url&redirect_url_clean=&reward=__reward__&session_id=__session_id__&slot=__slot__&tracking_url=http%3A%2F%2Fbuzzvil.com&unit_device_token=TestUser-13&unit_id=1&use_clean_mode=__use_clean_mode__&use_reward_api=true",
				},
				"meta": map[string]interface{}{},
			}},
		},
	}

	for _, c := range cases {
		testCase.MockBuzzAds(c.BAResponse)

		requestParams, _ := testCase.getRequestParams(t)
		for key, value := range c.Request {
			requestParams.Set(key, fmt.Sprintf("%v", value))
		}

		response, statusCode, err := testCase.GetAdsResponse(t, requestParams)
		require.Nil(t, err)
		require.Equal(t, statusCode, 200)

		assert.Equal(t, len(c.BSResponse), len(response.Result.Ads), c.Description)
		if len(c.BSResponse) > 0 {
			expect := c.BSResponse[0]
			actual := response.Result.Ads[0]

			expect_click_url, _ := url.Parse(expect["creative"].(map[string]interface{})["click_url"].(string))
			received_click_url := actual["creative"].(map[string]interface{})["click_url"].(string)
			actual_click_url, _ := url.Parse(received_click_url)
			assert.Equal(t, expect_click_url.Query().Get("use_reward_api"), actual_click_url.Query().Get("use_reward_api"), fmt.Sprintf("%s wrong use_reward_api %s", c.Description, received_click_url))
			assert.Equal(t, expect_click_url.Query().Get("tracking_url"), actual_click_url.Query().Get("tracking_url"), fmt.Sprintf("%s wrong tracking_url %s", c.Description, received_click_url))
		}

		testCase.buzzAdServer.ResponseHandlers = nil
	}
}

func TestBuildingClickURL(t *testing.T) {
	testCase := &AdTestCase{AdTestConfig: V3AdTestConfig{}}
	testCase.setUp()
	defer testCase.tearDown()

	buzzAdID := 100001
	originalClickURL := "http://test/click.url"
	testCase.MockBuzzAds([]ad.AdV2{
		*buildBuzzAdV2From(func(ad *ad.AdV2) {
			ad.ID = int64(buzzAdID)
			ad.Creative["clickUrl"] = originalClickURL
		}),
	})

	requestParams, _ := testCase.getRequestParams(t)
	response, statusCode, err := testCase.GetAdsResponse(t, requestParams)
	require.Nil(t, err)
	require.Equal(t, statusCode, 200)

	ad := response.Result.Ads[0]
	clickURL := ad["creative"].(map[string]interface{})[testCase.getQueryName("click_url")].(string)

	if clickURL == "" {
		t.Fatal("Ad response should contain click url value: ", ad)
	}

	if clickURL == originalClickURL {
		t.Fatal("Click URL should not be equal to the original click URL: ", clickURL)
	}

	parsedURL, _ := url.Parse(clickURL)
	redirectURL := parsedURL.Query()["redirect_url"][0]
	if redirectURL != originalClickURL {
		t.Fatal("redirect_url should be equal to the original click URL: ", redirectURL)
	}

	ifa := parsedURL.Query()["ifa"][0]
	adID := requestParams.Get(testCase.getQueryName("ad_id"))
	if ifa != adID {
		t.Fatal("ifa should be equal to the request's ad id: ", ifa, adID)
	}

	campaignID, _ := strconv.Atoi(parsedURL.Query()["campaign_id"][0])
	if campaignID != buzzAdID+dto.BuzzAdCampaignIDOffset {
		t.Fatalf("Campaign ID should be %v but %v", buzzAdID+dto.BuzzAdCampaignIDOffset, campaignID)
	}
}

func TestGetAdsValidatingValidLandingReward(t *testing.T) {
	testGetAdsValidatingLandingReward(t, &AdTestCase{AdTestConfig: V3AdTestConfig{}}, reward.StatusUnknown)
}

func TestGetAdsValidatingInvalidLandingReward(t *testing.T) {
	testGetAdsValidatingLandingReward(t, &AdTestCase{AdTestConfig: V3AdTestConfig{}}, reward.StatusReceived)
}

func TestGetAdsAnonymousUser(t *testing.T) {
	testCase := &AdTestCase{AdTestConfig: V3AdTestConfig{}}
	testCase.setUp()
	defer testCase.tearDown()

	testCase.MockBuzzAds([]ad.AdV2{
		*buildBuzzAdV2From(func(ad *ad.AdV2) {
			ad.ID = int64(100001)
		}),
	})

	cases := []struct {
		name       string
		anonymous  string
		adsCount   int
		statusCode int
	}{
		{name: "anonymous", anonymous: "true", adsCount: 1, statusCode: 200},
		{name: "invalid session", anonymous: "false", adsCount: 0, statusCode: 401},
	}

	for _, tc := range cases {
		requestParams, _ := testCase.getRequestParams(t)
		requestParams.Del("session_key")
		requestParams.Del("gender")
		requestParams.Del("birthday")
		requestParams.Add("anonymous", tc.anonymous)

		t.Run(tc.name, func(t *testing.T) {
			response, statusCode, err := testCase.GetAdsResponse(t, requestParams)
			require.Nil(t, err)
			assert.Equal(t, statusCode, tc.statusCode)
			assert.Equal(t, len(response.Result.Ads), tc.adsCount, "wrong ads count")
		})
	}
}

func testGetAdsValidatingLandingReward(t *testing.T, testCase *AdTestCase, receivedStatus reward.ReceivedStatus) {
	testCase.setUp()
	defer testCase.tearDown()

	testCase.MockBuzzAds([]ad.AdV2{
		*buildBuzzAdV2From(func(ad *ad.AdV2) {
			ad.ID = int64(100001)
			ad.LandingReward = 100
		}),
	})

	requestParams, device := testCase.getRequestParams(t)
	if receivedStatus == reward.StatusReceived {
		testCase.createPoint(t, 100001+dto.BuzzAdCampaignIDOffset, requestParams.Get("ad_id"), int64(device.Result["device_id"].(float64)))
	}

	response, statusCode, err := testCase.GetAdsResponse(t, requestParams)
	require.Nil(t, err)
	require.Equal(t, statusCode, 200)

	if len(response.Result.Ads) != 1 {
		t.Fatal("All ad should be included.\n\t", response)
	}
}

func (tc *AdTestCase) GetAdsResponse(t *testing.T, requestParams *url.Values) (*TestAdsResponse, int, error) {
	if requestParams == nil {
		requestParams, _ = tc.getRequestParams(t)
	}

	var response TestAdsResponse
	statusCode, err := (&network.Request{
		Method: "GET",
		Header: &http.Header{
			"Buzz-App-ID": []string{"1000000001"},
		},
		Params: requestParams,
		URL:    ts.URL + tc.getPath(),
	}).GetResponse(&response)

	return &response, statusCode, err
}

func (tc *AdTestCase) setUp() {
	httpClient := network.DefaultHTTPClient
	tc.buzzAdServer = mock.NewTargetServer(network.GetHost(os.Getenv("BUZZAD_URL")))
	tc.clientPatcher = mock.PatchClient(httpClient, tc.buzzAdServer)
	tc.pointTable = tc.setupDynamoDBPointTable()
}

func (tc *AdTestCase) tearDown() {
	tc.clientPatcher.RemovePatch()
	tc.tearDownDynamoDBPointTable(tc.pointTable)
}

func (tc *AdTestCase) MockBuzzAds(ads []ad.AdV2) {
	adsResponse := ad.V2AdsResponse{Ads: ads}
	adsResponseString, _ := json.Marshal(&adsResponse)
	tc.buzzAdServer.AddResponseHandler(&mock.ResponseHandler{
		WriteToBody: func() []byte {
			return []byte(adsResponseString)
		},
		Path:   "/api/v2/lockscreen/ads",
		Method: http.MethodGet,
	})
}

func (tc *AdTestCase) setupDynamoDBPointTable() dynamo.Table {
	dyDB := env.GetDynamoDB()
	// Create the table
	if err := dyDB.CreateTable(env.Config.DynamoTablePoint, rewardRepo.DBPoint{}).Run(); err != nil {
		dyDB.Table(env.Config.DynamoTablePoint).DeleteTable().Run()
		if err := dyDB.CreateTable(env.Config.DynamoTablePoint, rewardRepo.DBPoint{}).Run(); err != nil {
			core.Logger.Fatalf("SetupTest failed with %v", err)
		}
	}
	return dyDB.Table(env.Config.DynamoTablePoint)
}

func (tc *AdTestCase) tearDownDynamoDBPointTable(pointTable dynamo.Table) {
	pointTable.DeleteTable().Run()
}

func (tc *AdTestCase) createPoint(t *testing.T, campaignID int64, adID string, deviceID int64) *rewardRepo.DBPoint {
	var p *rewardRepo.DBPoint
	faker.FakeData(&p)
	p.DeviceID = deviceID
	p.Type = "imp"
	p.ReferKey = strconv.FormatInt(campaignID, 10)
	now := time.Now().Unix()
	p.CreatedAt = now
	p.UpdatedAt = now
	p.Slot = now - now%3600

	err := tc.pointTable.Put(p).Run()
	if err != nil {
		t.Fatal("Faield to put record into dynamoDB")
	}
	return p
}

func newAdTestCase(t *testing.T, mockRes *BuzzAdMockRes) *AdTestCase {
	return &AdTestCase{
		t:            t,
		mockRes:      mockRes,
		AdTestConfig: V3AdTestConfig{},
	}
}

type (
	AdTestConfig interface {
		getRequestParams(t *testing.T) (*url.Values, *dto.CreateDeviceResponse)
		getQueryName(name string) string
		getPath() string
	}

	// AdTestCase type definition
	AdTestCase struct {
		mockRes       *BuzzAdMockRes
		t             *testing.T
		ver           int
		buzzAdServer  *mock.TargetServer
		clientPatcher *mock.ClientPatcher
		pointTable    dynamo.Table
		AdTestConfig
	}

	V3AdTestConfig struct{}
)

func (tc V3AdTestConfig) getRequestParams(t *testing.T) (*url.Values, *dto.CreateDeviceResponse) {
	return buildV3BaseTestRequest(t)
}

func (tc V3AdTestConfig) getPath() string {
	return "/api/v3/ads"
}

func (tc V3AdTestConfig) getQueryName(name string) string {
	return utils.CamelToUnderscore(name)
}

func (tc *AdTestCase) run(equalFuncs ...func(oriAd ad.AdV2, resAd map[string]interface{}) bool) {
	t := tc.t
	httpClient := network.DefaultHTTPClient
	buzzAdServer := mock.NewTargetServer(network.GetHost(os.Getenv("BUZZAD_URL"))).AddResponseHandler(&mock.ResponseHandler{
		WriteToBody: func() []byte {
			return []byte(tc.mockRes.getResponse())
		},
		Path:   "/api/v2/lockscreen/ads",
		Method: http.MethodGet,
	})

	var statuses []reward.ReceivedStatus
	for i := 0; i < len(tc.mockRes.AdsRes.Ads); i++ {
		statuses = append(statuses, reward.StatusUnknown)
	}

	clientPatcher := mock.PatchClient(httpClient, buzzAdServer, getInsightMockServer())

	defer clientPatcher.RemovePatch()

	var adsRes TestAdsResponse
	requestParams, _ := tc.getRequestParams(t)
	statusCode, err := (&network.Request{
		Method: "GET",
		Params: requestParams,
		URL:    ts.URL + tc.getPath(),
	}).GetResponse(&adsRes)

	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if err != nil || statusCode != 200 {
		t.Fatal(statusCode, err, adsRes)
	} else {
		for i, resAd := range adsRes.Result.Ads {
			oriAd := (*tc.mockRes.getAds())[i]
			test.AssertEqual(t, oriAd.ID+dto.BuzzAdCampaignIDOffset, int64(resAd["id"].(float64)), "TestGetAds() - id")
			if tc.ver == 2 {
				test.AssertEqual(t, oriAd.ActionReward, int(resAd["actionReward"].(float64)), "TestGetAds() - actionReward")
			} else {
				test.AssertEqual(t, oriAd.ActionReward, int(resAd["action_reward"].(float64)), "TestGetAds() - action_reward")
			}
			test.AssertEqual(t, (*oriAd.TTL-int(resAd["ttl"].(float64))) == 0, true, "TestGetAds() - ttl")

			for i, ct := range oriAd.ClickTrackers {
				test.AssertEqual(t, ct, oriAd.ClickTrackers[i], "TestGetAds() - ClickTrackers")
			}

			for i, it := range oriAd.ImpressionTrackers {
				test.AssertEqual(t, it, oriAd.ImpressionTrackers[i], "TestGetAds() - ImpressionTrackers")
			}

			//noinspection ALL
			test.AssertEqual(t, resAd["creative"].(map[string]interface{})["type"] == "", false, fmt.Sprintf("TestGetAds() - IntegrationType, resAd: %+v", resAd))
			test.AssertEqual(t, oriAd.Creative["type"], resAd["creative"].(map[string]interface{})["type"].(string), fmt.Sprintf("TestGetAds() - CreativeType, resAd: %+v", resAd))

			for _, assertCase := range equalFuncs {
				funcName := runtime.FuncForPC(reflect.ValueOf(assertCase).Pointer()).Name()
				test.AssertEqual(t, assertCase(oriAd, resAd), true, fmt.Sprintf("TestGetAds() - %v", funcName))
			}
		}
	}
	t.Logf("AdTestCase.run() - types: %v", tc.mockRes.Types)
}

func buildBuzzAdV2From(fun func(ad *ad.AdV2)) *ad.AdV2 {
	ad := createBuzzAdV2("SDK")
	fun(&ad)
	return &ad
}

func createBuzzAdV2(adType string) ad.AdV2 {
	ttl := 3600
	adV2 := ad.AdV2{
		ActionReward: 0,
		ClickTrackers: []string{
			"https://ad.buzzvil.com/api/v3/click?data=qvLA_ej9BLukC6vuOfvld-Kh0r58H-BU2vGUZIKQZPWNOsG9BxEvEdJKU68ckPElUshKcQcsiWqblruedC-qR4-EZsg3E0ipeaYVADg-SitSqU3oi5G5ZrVCCUEThHe_40uuYOVqopcEi0FkQf2w8A3l0ie2JvH_hc4-5O52rL-a5diLPMUlPCC0zWRHyxCzq01tmap8XWMyIPLObp6hSsT4cQGwIanfcpaHeLGm8h6eBNwtlQbzybr4ga44fv_2XdJGRCnzlTOljXQbpHkEgxlSdl0T7pjY4rsb9q4WqjgD1kkiC4b5wjj2K9F9KpVN",
			"https://ad.buzzvil.com/api/v3/lockscreen/slide_joy_tracker?data=BO0tBYJxVZexhVbaT9jyGgNncoy-brqQZw286gU2NussKFH8PQH-D0E1PELTCZ5chgsamcoOJzYJZuel6Jbhspuu2jLoEnZ6fBtd3JDYkYYYYlQsnv07ER-AJ3yAxWopsfXC8LHGiV82d7btFJNqAbIheQYOCmy1O0_sanMn1qj-7KWjf0vpLtu-_TzEnaHqx3ffy1DqBgy4CMiIyNqyLvn36_u93sIWTtVwHp2wWNBYFrdLzPvgS1nO1kbsdXKzARuuPqnQNJuXRz4H8MSzKKCS-CB48dHjVSsHvVFc5ZQ%3D",
		},
		CallToAction:  "더보기",
		LandingReward: 0,
		ImpressionTrackers: []string{
			"https://ad.buzzvil.com/api/lockscreen/impression?data=qvLA_ej9BLukC6vuOfvld-Kh0r58H-BU2vGUZIKQZPWNOsG9BxEvEdJKU68ckPElUshKcQcsiWqblruedC-qR4-EZsg3E0ipeaYVADg-SitSqU3oi5G5ZrVCCUEThHe_40uuYOVqopcEi0FkQf2w8A3l0ie2JvH_hc4-5O52rL-a5diLPMUlPCC0zWRHyxCzq01tmap8XWMyIPLObp6hSsT4cQGwIanfcpaHeLGm8h6eBNwtlQbzybr4ga44fv_2XdJGRCnzlTOljXQbpHkEgxlSdl0T7pjY4rsb9q4WqjgD1kkiC4b5wjj2K9F9KpVN",
			"https://ad.buzzvil.com/api/v3/lockscreen/slide_joy_tracker?data=2mljviCZ58fMWps1L7WT_yF5NH9Epfwi8mQqzYpIPAB8fQKVy_DY4i4u1qmCBthjLLsregVHH5d34eE6Gfz4ef7A3LcrX54oROFPZZRPTAhsoDHDobkdBIcdcrciATS_bwoGmTnA7R97Cv7ma3rRpP99RggGxwXeh2ShRrUDfetUMWE6-QFkH2AmDI5zqKX9TB70vjfsIFAgLMUjZwS_2C-QSaf6bcDlnsfm4pg22mxZ28hAtmHiHCU9r8f59vPMy0HTq3BCl1AYfbOmwGkCE7F8HiPcZxKQVhvUnGQ0H0w%3D",
		},
		TTL:            &ttl,
		UnlockReward:   0,
		OrganizationID: 1000,
		RewardPeriod:   7200,
	}
	switch adType {
	case "WEB":
		adV2.ID = 1502281
		adV2.Creative = map[string]interface{}{
			"clickUrl":    "https://api.getslidejoy.com/redirect?url=slide%3A%2F%2Fhome%3Fpage%3D2",
			"height":      480,
			"width":       320,
			"htmlTag":     "<!DOCTYPE html><html><head><meta charset=\"utf-8\"/><meta name=\"viewport\" content=\"width=device-width\"/><link href=\"https://s3-us-west-2.amazonaws.com/slidejoy-misc/mediation.css\"rel=\"stylesheet\"type=\"text/css\"/></head><body><div id=\"sj_wrapper\"class=\"sj_page\"><div id=\"sj_stage\"><img class=\"creative\"src=\"https://s3-us-west-2.amazonaws.com/slidejoy-misc/offers-tab-new.jpg\"/></div></div></body></html>",
			"bgUrl":       "https://buzzvil.akamaized.net/adfit.image/uploads/1498632095-KQODI.jpg",
			"sizeType":    "INTERSTITIAL",
			"landingType": 3,
			"type":        "WEB",
		}
	case "SDK":
		adV2.ID = 1502235
		adV2.Creative = map[string]interface{}{
			"bgUrl":       "https://buzzvil.akamaized.net/adfit.image/uploads/1508230314-ZI5BO.jpg",
			"landingType": 3,
			"network":     "OUTBRAIN",
			"placementId": "SDK_3",
			"publisherId": "SLIDE1L0QG0CFL3J7BP1L93J6",
			"referrerUrl": "http://www.getslidejoy.com/japan",
			"type":        "SDK",
		}
	case "JS":
		adV2.ID = 1378968
		adV2.Creative = map[string]interface{}{
			"bgUrl":        "https://buzzvil.akamaized.net/adfit.image/uploads/1485174717-54P8S.jpg",
			"height":       480,
			"htmlTag":      "<!DOCTYPE html>\n<html>\n<head>\n    <meta charset=\"utf-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=0\">\n    <script type=\"text/javascript\" src=\"https://cdn-ad-static.buzzvil.com/mediation/mediation-20170828-03.js\"></script>\n</head>\n<body style=\"background-color:black\">\n<div id=\"bz_wrapper\" class=\"bz_page\"><div id=\"bz_stage\"></div></div>\n<script>\n    var params = {\"relationship\": \"M\", \"ip\": \"220.118.92.225\", \"age\": 33, \"longitude\": 127.0, \"advertisingID\": \"e2b481bc-6f74-4985-a9ba-980a0553a9cc\", \"androidIDSHA1\": \"68807258c8fd781b3d3896a114bbba3f0e052981\", \"gender\": \"M\", \"latitude\": 36.0, \"androidIDMD5\": \"f876a422b20f02d990b84ef624c168a0\", \"ua\": \"Dalvik/1.6.0 (Linux; U; Android 4.4.2; LG-F400L Build/KVT49L.F400L11a)\"};\n    var network = {\"packageName\": \"com.slidejoy\", \"siteID\": \"8a809418015454bf83f9e32ca3b902e4\", \"name\": \"nexage\", \"placementID\": \"slidejoy_interstitials\"};\n    mediate(params, network);\n</script>\n</body>\n</html>\n",
			"landingType":  3,
			"sizeType":     "INTERSTITIAL",
			"support_webp": true,
			"type":         "JS",
			"width":        320,
		}
	}
	return adV2
}

type (
	// BuzzAdMockRes type definition
	BuzzAdMockRes struct {
		Types  []string
		AdsRes ad.V2AdsResponse
	}

	// TestAdsResponse type definition
	TestAdsResponse struct {
		Result struct {
			Ads []map[string]interface{} `json:"ads"`
		} `json:"result"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
)

func (mr *BuzzAdMockRes) getResponse() string {
	if len(mr.Types) != len(mr.AdsRes.Ads) {
		mr.build()
	}
	res, err := json.Marshal(&mr.AdsRes)
	if err != nil {
		panic(err)
	}
	return string(res)
}

func (mr *BuzzAdMockRes) getAds() *[]ad.AdV2 {
	return &mr.AdsRes.Ads
}

func (mr *BuzzAdMockRes) build() *BuzzAdMockRes {
	for _, adType := range mr.Types {
		mr.AdsRes.Ads = append(mr.AdsRes.Ads, createBuzzAdV2(adType))
	}
	return mr
}
