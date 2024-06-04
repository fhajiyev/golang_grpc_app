package controller_test

import (
	"fmt"
	"math/rand"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"strconv"

	"net/http"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/Buzzvil/go-test/mock"
	uuid "github.com/satori/go.uuid"
)

var ts *httptest.Server

func buildV1BaseTestRequest() *url.Values {
	return &url.Values{
		"ifa":               {"069e6a97-b341-43a0-b9fc-df2556f06a25"},
		"app_id":            {strconv.FormatInt(tests.TestAppID1, 10)},
		"unit_device_token": {"1b66a654cd0d4c4fb471f4fb02b65015"},
		"device_id":         {strconv.Itoa(1092)},
		"device_name":       {"SHV-E210S"},
		"device_os":         {"21"},
		"year_of_birth":     {strconv.Itoa(1985)},
		"sex":               {"M"},
		"carrier":           {"KT"},
		"region":            {"서울시 관악구"},
		"sdk_version":       {"1820"},
		"country":           {"KR"},
		"timezone":          {"Asia/Tokyo"},
	}
}

func buildV2BaseTestRequest(t *testing.T) (*url.Values, *dto.CreateDeviceResponse) {
	adID := uuid.NewV4().String()
	device, _ := execTestDevice(t, tests.TestAppID1, fmt.Sprintf("TestUser-%d", rand.Intn(100)), adID)
	sessionKey := device.Result["session_key"].(string)
	return &url.Values{
		"adId":        {uuid.NewV4().String()},
		"sdkVersion":  {"20042"},
		"timeZone":    {"America/Los_Angeles"},
		"locale":      {"ko_KR"},
		"deviceName":  {"Nexus One"},
		"birthday":    {"1984-06-07"},
		"gender":      {"M"},
		"sessionKey":  {sessionKey},
		"networkType": {"wifi"},
		"osVersion":   {"21"},
		"types":       {"{\"IMAGE\":[\"INTERSTITIAL\"],\"SDK\":[\"FAN\",\"OUTBRAIN\"]}"},
		"userAgent":   {"Mozilla/5.0 (Linux; Android 4.0.4; Galaxy Nexus Build/IMM76B) AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.133 Mobile Safari/535.19"},
		"unitId":      {strconv.FormatInt(tests.TestAppID1, 10)},
	}, device
}

func buildV3BaseTestRequest(t *testing.T) (*url.Values, *dto.CreateDeviceResponse) {
	adID := uuid.NewV4().String()
	device, _ := execTestDevice(t, tests.TestAppID1, fmt.Sprintf("TestUser-%d", rand.Intn(100)), adID)
	sessionKey := device.Result["session_key"].(string)
	return &url.Values{
		"ad_id":        {uuid.NewV4().String()},
		"sdk_version":  {"20042"},
		"timezone":     {"America/Los_Angeles"},
		"locale":       {"ko_KR"},
		"device_name":  {"Nexus One"},
		"birthday":     {"1984-06-07"},
		"gender":       {"M"},
		"session_key":  {sessionKey},
		"network_type": {"wifi"},
		"os_version":   {"21"},
		"types":        {"{\"IMAGE\":[\"INTERSTITIAL\"],\"SDK\":[\"FAN\",\"OUTBRAIN\"]}"},
		"user_agent":   {"Mozilla/5.0 (Linux; Android 4.0.4; Galaxy Nexus Build/IMM76B) AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.133 Mobile Safari/535.19"},
		"unit_id":      {strconv.FormatInt(tests.TestAppID1, 10)},
	}, device
}

func getPatchedHttpClient(hosts ...string) (*http.Client, func()) {
	httpClient := network.DefaultHTTPClient

	mockServers := make([]*mock.TargetServer, 0)
	mockServers = append(mockServers, getInsightMockServer())

	for _, host := range hosts {
		switch host {
		case os.Getenv("BUZZAD_URL"):
			mockServers = append(mockServers, getBuzzAdMockServer())
		default:
			panic(fmt.Sprintf("getPatchedHttpClient() - don't know this host: %s ", host))
		}
	}

	clientPatcher := mock.PatchClient(httpClient, mockServers...)

	return httpClient, clientPatcher.RemovePatch
}

func getBuzzAdMockServer() *mock.TargetServer {
	return mock.NewTargetServer(network.GetHost(os.Getenv("BUZZAD_URL"))).AddResponseHandler(&mock.ResponseHandler{
		WriteToBody: func() []byte {
			return []byte(BuzzAdV2Res)
		},
		Path:   "/api/v2/lockscreen/ads",
		Method: http.MethodGet,
	}).AddResponseHandler(&mock.ResponseHandler{
		WriteToBody: func() []byte {
			return []byte(BuzzAdV1Res)
		},
		Path:   "/api/lockscreen/ads",
		Method: http.MethodGet,
	})
}

func getInsightMockServer() *mock.TargetServer {
	rh := &mock.ResponseHandler{
		WriteToBody: func() []byte {
			return []byte(InsightLocationRes)
		},
		Path:   "/location",
		Method: http.MethodGet,
	}
	return mock.NewTargetServer(network.GetHost(env.Config.InsightURL)).AddResponseHandler(rh)
}

const InsightLocationRes = `
{
    "code": 0,
    "message": null,
    "result": {
        "location": {
            "country": "KR",
            "zipcode": null,
            "state": "11",
            "state2": null,
            "state3": null,
            "state4": null,
            "city": "Seoul",
            "timeZone": "Asia/Seoul",
            "latitude": 37.5111,
            "longitude": 126.9743,
            "ipAddress": "112.216.231.234"
        }
    }
}
`

const BuzzAdV1Res = `
{
  "native_ads": [
    {
      "support_webp": false,
      "click_url": "",
      "extra": {},
      "first_display_weight": 10000000,
      "image": "https://cdn-ad-static.buzzvil.com/native_ad_bg/fan_bg_3.jpg",
      "remove_after_impression": false,
      "ipu": 1,
      "started_at": 1493264640,
      "slot": "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
      "first_display_priority": 10,
      "age_from": null,
      "unit_price": 0,
      "impression_urls": [
        "https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOHSDKETb7HA31RGtD3AC-dhcQArUCZDDydCFRcScFNvnlN_YL_I7-6q2Xn2BqFx8o1EU2Rsq3lvc-gsU3xVDKhlDiZDORayOh1ASUx7Aw9oV1MIkOyts_sqDnmjovYsyfB0GdkldbT28ekJy9yplTwYa4fdsfZCKIq8FNmONvGa8zXT7GMIj2ngdtzNFDh9lGMEHhqE2ycQ25CifKU4_sXQKZdabcyV1gZ6uFj-BRzTRk80ho_QSgpeaclDmimI-huST1k73R1lRV7xs1bOGXZRaUjlptKMH_P9xlGxCwI0hyq8syzdOrcMoClQq_V0ZowaPKeWAfytXFrI9qMRzs7rXDMnDQ7A-VSg6MLrlTswHjgzqnUr-oTvRfxs0Vbswng%3D"
      ],
      "action_reward": 0,
      "unlock_reward": 0,
      "type": "cpc",
      "use_web_ua": false,
      "image_ios": null,
      "revenue_type": "cpc",
      "dipu": 9999,
      "organization_id": 1,
      "sdk_version_to": null,
      "banner": {
        "support_webp": false,
        "placement_id": "",
        "period": 7200,
        "background": "https://buzzvil.akamaized.net/adfit.image/native_ad_bg/8.png",
        "lifetime": 3600,
        "id": "DAN-urumlgivnrdw",
        "network": "ADFIT",
        "name": "af",
        "filterable": false,
        "type": "BANNER_SDK",
        "referrer_url": "",
        "adchoice_url": null,
        "publisher_id": ""
      },
      "adnetwork_id": 519,
      "name": "DDN test",
      "call_to_action": "더보기",
      "landing_reward": 0,
      "sdk_version_from": null,
      "sex": null,
      "preferred_browser": null,
      "id": 1381015,
      "age_to": null,
      "is_rtb": false,
      "creative": {
        "support_webp": false,
        "placement_id": "",
        "period": 7200,
        "background": "https://buzzvil.akamaized.net/adfit.image/native_ad_bg/8.png",
        "lifetime": 3600,
        "id": "DAN-urumlgivnrdw",
        "network": "ADFIT",
        "name": "af",
        "filterable": false,
        "type": "BANNER_SDK",
        "referrer_url": "",
        "adchoice_url": null,
        "publisher_id": ""
      },
      "is_incentive": false,
      "tipu": 0,
      "owner_id": 425,
      "target_app": null,
      "display_type": "A",
      "icon": "",
      "landing_type": "browser",
      "ended_at": 1535080173,
      "click_beacons": [
        "https://ad-dev.buzzvil.com/api/v1/click_track?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOHSDKETb7HA31RGtD3AC-dhcQArUCZDDydCFRcScFNvnlN_YL_I7-6q2Xn2BqFx8o1EU2Rsq3lvc-gsU3xVDKhlDiZDORayOh1ASUx7Aw9oV1MIkOyts_sqDnmjovYsyfB0GdkldbT28ekJy9yplTwYa4fdsfZCKIq8FNmONvGa8zXT7GMIj2ngdtzNFDh9lGMEHhqE2ycQ25CifKU4_sXQKZdabcyV1gZ6uFj-BRzTRk80ho_QSgpeaclDmimI-huST1k73R1lRV7xs1bOGXZRaUjlptKMH_P9xlGxCwI0hyq8syzdOrcMoClQq_V0ZowaPKeWAfytXFrI9qMRzs7rXDMnDQ7A-VSg6MLrlTswHjgzqnUr-oTvRfxs0Vbswng%3D",
        "http://abc.def/adfwef?adef=1381015&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=DDN%20test&abc="
      ],
      "region": null,
      "device_name": null,
      "carrier": null
    },
    {
      "support_webp": true,
      "click_url": "",
      "extra": {},
      "first_display_weight": 41,
      "image": "https://buzzvil.akamaized.net/adfit.image/uploads/1492063831-UL5J6.jpg",
      "remove_after_impression": false,
      "ipu": 1,
      "started_at": 1485248400,
      "slot": "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
      "first_display_priority": 10,
      "age_from": null,
      "unit_price": 0,
      "impression_urls": [
        "https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOGI7JF2cftyCzB0cK_QW-j1JRcoKlof9W6QPapcaYu49O_tBeUff0uDmK0rq1RWFDKUbZ03qzmpziavRyhy-FM37fYembbeQsjJNuNWWuqos_f_jd0SEzy5qhP5NDORBaA83gIvRZ5H-Xy9QWTaUBFu7Hp6CZhDLTv6GRhDBSCBW7y0RNCXkA4_lGdl8f62pZyLddX2YShMz_dFkK3Zsdz4D3kA6Zqoy4hTgd0C8zUx-w1f2k7ZjbAFdkvFYOe51YI9MbiGNZoRFOdDWBYVuoh26yoK_3YhKhyNJjPAxZQi3UphVqao2nh75Oob1RIxf1-qUcrYIrJONo6PLVdrxX1DwFSugTursISPPfeTMiGo7EZNs-Zc6ubL-TmuQroYzYE%3D"
      ],
      "action_reward": 0,
      "unlock_reward": 0,
      "type": "cpc",
      "use_web_ua": false,
      "image_ios": null,
      "revenue_type": "cpc",
      "dipu": 9999,
      "organization_id": 1,
      "sdk_version_to": null,
      "adnetwork_id": 503,
      "name": "mobvista",
      "call_to_action": "더보기",
      "landing_reward": 0,
      "native_ad": {
        "support_webp": true,
        "placement_id": "",
        "period": 14400,
        "background": "https://buzzvil.akamaized.net/adfit.image/native_ad_bg/7.png",
        "lifetime": 3000,
        "id": "11230",
        "network": "MOBVISTA",
        "name": "mv",
        "filterable": true,
        "type": "SDK",
        "referrer_url": "",
        "adchoice_url": null,
        "publisher_id": ""
      },
      "sdk_version_from": null,
      "sex": null,
      "preferred_browser": null,
      "id": 1225041,
      "age_to": null,
      "is_rtb": false,
      "creative": {
        "support_webp": true,
        "placement_id": "",
        "period": 14400,
        "background": "https://buzzvil.akamaized.net/adfit.image/native_ad_bg/7.png",
        "lifetime": 3000,
        "id": "11230",
        "network": "MOBVISTA",
        "name": "mv",
        "filterable": true,
        "type": "SDK",
        "referrer_url": "",
        "adchoice_url": null,
        "publisher_id": ""
      },
      "is_incentive": false,
      "tipu": 0,
      "owner_id": 704,
      "target_app": null,
      "display_type": "A",
      "icon": "",
      "landing_type": "browser",
      "ended_at": 1535080173,
      "click_beacons": [
        "https://ad-dev.buzzvil.com/api/v1/click_track?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOGI7JF2cftyCzB0cK_QW-j1JRcoKlof9W6QPapcaYu49O_tBeUff0uDmK0rq1RWFDKUbZ03qzmpziavRyhy-FM37fYembbeQsjJNuNWWuqos_f_jd0SEzy5qhP5NDORBaA83gIvRZ5H-Xy9QWTaUBFu7Hp6CZhDLTv6GRhDBSCBW7y0RNCXkA4_lGdl8f62pZyLddX2YShMz_dFkK3Zsdz4D3kA6Zqoy4hTgd0C8zUx-w1f2k7ZjbAFdkvFYOe51YI9MbiGNZoRFOdDWBYVuoh26yoK_3YhKhyNJjPAxZQi3UphVqao2nh75Oob1RIxf1-qUcrYIrJONo6PLVdrxX1DwFSugTursISPPfeTMiGo7EZNs-Zc6ubL-TmuQroYzYE%3D",
        "http://abc.def/adfwef?adef=1225041&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=mobvista&abc="
      ],
      "region": null,
      "device_name": null,
      "carrier": null
    }
  ],
  "code": 200,
  "ads": [
    {
      "support_webp": true,
      "click_url": "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOEOxhD0jzsi8kQaPV0Fb9LsmFvmmi_TZRp4k2j2pquj8SjQKZa73OTnRxVHrDzVzCQgyOmzbyLY7Dy7yR0qPf_fVArqbHLRzCzMqDIoM8e57wTiHJxFiArqofrGM-oC9wZj7l7tAofH88mmPGaRmi6ZylURcMDOSBUeDjiMJV5z2V0guOiPiTf_9kl6OMMH6G1F5bsHxIvLLquNwiFH8iGdXCYhPaEg1kpHFR66sa0A8f4UjKGGBnIYMhZ8mTn7lTJj4CFfjOP8Fxd280aiCKQIZkvD9wC9Cp5ILW-pQ-OnNxrB8kRzFoQZXzEUoccFjusrGu_M6KaJGAyTdX_ZTimx4s-ZVSL9oGgDKaM5QcN3wXljHXaheJElGmQBFiDGEWE%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
      "extra": {},
      "first_display_weight": 10000000,
      "image": "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
      "remove_after_impression": true,
      "ipu": 9999,
      "started_at": 1534431600,
      "slot": "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
      "first_display_priority": 10,
      "age_from": null,
      "unit_price": 21,
      "impression_urls": [
        "https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOEOxhD0jzsi8kQaPV0Fb9LsmFvmmi_TZRp4k2j2pquj8SjQKZa73OTnRxVHrDzVzCQgyOmzbyLY7Dy7yR0qPf_fVArqbHLRzCzMqDIoM8e57wTiHJxFiArqofrGM-oC9wZj7l7tAofH88mmPGaRmi6ZylURcMDOSBUeDjiMJV5z2V0guOiPiTf_9kl6OMMH6G1F5bsHxIvLLquNwiFH8iGdXCYhPaEg1kpHFR66sa0A8f4UjKGGBnIYMhZ8mTn7lTJj4CFfjOP8Fxd280aiCKQIZkvD9wC9Cp5ILW-pQ-OnNxrB8kRzFoQZXzEUoccFjusrGu_M6KaJGAyTdX_ZTimx4s-ZVSL9oGgDKaM5QcN3wXljHXaheJElGmQBFiDGEWE%3D"
      ],
      "action_reward": 0,
      "unlock_reward": 0,
      "type": "cpc",
      "use_web_ua": false,
      "image_ios": null,
      "revenue_type": "cpc",
      "dipu": 9999,
      "organization_id": 1,
      "sdk_version_to": null,
      "adnetwork_id": null,
      "name": "Hmall::",
      "call_to_action": "더보기",
      "landing_reward": 1,
      "sdk_version_from": null,
      "sex": null,
      "preferred_browser": null,
      "id": 1383378,
      "age_to": null,
      "is_rtb": false,
      "creative": {
        "landing_type": "browser",
        "support_webp": true,
        "click_url": "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOEOxhD0jzsi8kQaPV0Fb9LsmFvmmi_TZRp4k2j2pquj8SjQKZa73OTnRxVHrDzVzCQgyOmzbyLY7Dy7yR0qPf_fVArqbHLRzCzMqDIoM8e57wTiHJxFiArqofrGM-oC9wZj7l7tAofH88mmPGaRmi6ZylURcMDOSBUeDjiMJV5z2V0guOiPiTf_9kl6OMMH6G1F5bsHxIvLLquNwiFH8iGdXCYhPaEg1kpHFR66sa0A8f4UjKGGBnIYMhZ8mTn7lTJj4CFfjOP8Fxd280aiCKQIZkvD9wC9Cp5ILW-pQ-OnNxrB8kRzFoQZXzEUoccFjusrGu_M6KaJGAyTdX_ZTimx4s-ZVSL9oGgDKaM5QcN3wXljHXaheJElGmQBFiDGEWE%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
        "filterable": true,
        "image": "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
        "adchoice_url": null,
        "type": "IMAGE"
      },
      "is_incentive": true,
      "tipu": 0,
      "owner_id": 2147,
      "target_app": null,
      "display_type": "A",
      "icon": "",
      "landing_type": "browser",
      "ended_at": 1535080173,
      "click_beacons": [
        "http://abc.def/adfwef?adef=1383378&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=Hmall%3A%3A&abc="
      ],
      "region": null,
      "device_name": null,
      "carrier": null
    },
    {
      "support_webp": true,
      "click_url": "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
      "extra": {},
      "first_display_weight": 10000000,
      "image": "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
      "remove_after_impression": true,
      "ipu": 9999,
      "started_at": 1534431600,
      "slot": "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
      "first_display_priority": 10,
      "age_from": null,
      "unit_price": 21,
      "impression_urls": [
        "https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D"
      ],
      "action_reward": 0,
      "unlock_reward": 0,
      "type": "cpc",
      "use_web_ua": false,
      "image_ios": null,
      "revenue_type": "cpc",
      "dipu": 9999,
      "organization_id": 1,
      "sdk_version_to": null,
      "adnetwork_id": null,
      "name": "홈앤쇼핑::",
      "call_to_action": "더보기",
      "landing_reward": 1,
      "sdk_version_from": null,
      "sex": null,
      "preferred_browser": null,
      "id": 1383380,
      "age_to": null,
      "is_rtb": false,
      "creative": {
        "landing_type": "browser",
        "support_webp": true,
        "click_url": "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
        "filterable": true,
        "image": "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
        "adchoice_url": null,
        "type": "IMAGE"
      },
      "is_incentive": true,
      "tipu": 0,
      "owner_id": 2147,
      "target_app": null,
      "display_type": "A",
      "icon": "",
      "landing_type": "browser",
      "ended_at": 1535080173,
      "click_beacons": [
        "http://abc.def/adfwef?adef=1383380&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=%ED%99%88%EC%95%A4%EC%87%BC%ED%95%91%3A%3A&abc="
      ],
      "region": null,
      "device_name": null,
      "carrier": null
    },
    {
      "support_webp": true,
      "click_url": "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
      "extra": {},
      "first_display_weight": 10000000,
      "image": "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
      "remove_after_impression": true,
      "ipu": 9999,
      "started_at": 1534431600,
      "slot": "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
      "first_display_priority": 10,
      "age_from": null,
      "unit_price": 21,
      "impression_urls": [
        "https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D"
      ],
      "action_reward": 0,
      "unlock_reward": 0,
      "type": "cpc",
      "use_web_ua": false,
      "image_ios": null,
      "revenue_type": "cpc",
      "dipu": 9999,
      "organization_id": 1,
      "sdk_version_to": null,
      "adnetwork_id": null,
      "name": "홈앤쇼핑::",
      "call_to_action": "더보기",
      "landing_reward": 1,
      "sdk_version_from": null,
      "sex": null,
      "preferred_browser": null,
      "id": 1383380,
      "age_to": null,
      "is_rtb": false,
      "creative": {
        "landing_type": "browser",
        "support_webp": true,
        "click_url": "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
        "filterable": true,
        "image": "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
        "adchoice_url": null,
        "type": "IMAGE"
      },
      "is_incentive": true,
      "tipu": 0,
      "owner_id": 2147,
      "target_app": null,
      "display_type": "A",
      "icon": "",
      "landing_type": "browser",
      "ended_at": 1535080173,
      "click_beacons": [
        "http://abc.def/adfwef?adef=1383380&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=%ED%99%88%EC%95%A4%EC%87%BC%ED%95%91%3A%3A&abc="
      ],
      "region": null,
      "device_name": null,
      "carrier": null,
      "events": [
        {
          "tracking_urls": [
            "http://tracking.url.com"
          ],
          "event_type": "landed",
          "reward": {
            "amount": 1,
            "status": "RECEIVABLE"
          }
        }
      ]
    }
  ],
  "settings": {
    "filtering_words": "these|are|filtering|words,honeyscreen,cashslide,리워드,reward,성인,도박,카지노,sensitive,cash slide,캐시슬라이드,AliExpress",
    "web_ua": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"
  },
  "request_id": "be72dd5e-a742-11e8-b62c-0242ac110004",
  "msg": "ok"
}
`
const BuzzAdV2Res = `
{
    "msg": "ok",
    "code": 200,
    "ads": [
        {
            "support_webp": true,
            "landingReward": 0,
            "revenueType": "cpc",
            "callToAction": "See more",
            "ttl": 21600,
            "integrationType": "IMAGE",
            "id": 1668118,
            "network": "MOBPOWER",
            "name": "Mobpower",
            "creative": {
                "width": 720,
                "support_webp": true,
                "clickUrl": "https://ad.buzzvil.com/api/v1/click?data=nvypAq4ZcCO90sLS-wYLHZ8jWrqjKEZXXdv4rduVXAGhz4TncF90USS6t2FlHi8M7d9464GgoDSRzqdZ9kz99bQAHxSk-7UMihE-f3ODtBwVL8EErFW36ExnYx_z6XwQapw0Scbm5D84Xb01ySh_8lqF67ClEXeIh5_qEg51gWGwvxaXRUqmVmhX2slUEjvDA5aRCMLRCyv3KavoaykFjW9O9WCokUtpj52jUu4UgacWd8s_ARP2SXG8WIOSi057FE3Ea4wTakn_CG2kLK2zQq0d5C6Wdxr5u0vNvBpGOCQW987u0y3Y3mzSatVU8LnNAZm6XdYkhJH-WcJuT74Ji9fvARvbkVRN41bmLkeYfpv9IJNvkkAaWDh3T42BS3ZDjLNaqIjMPUuTzULeXyNpOCL55d446P5dAS6o0TBhEGZXw1Kumsa3bqZkALjxZqVi&redirect_url=http%3A%2F%2Ftknet.smardroid.com%2Fagentapi%2Fclick%3Fcid%3D245648129%26aid%3D5000036",
                "adchoiceUrl": null,
                "filterable": false,
                "imageUrl": "https://buzzvil.akamaized.net/adfit.image/uploads/1534993568-ZLN4J.jpg",
                "sizeType": "FULLSCREEN",
                "type": "IMAGE",
                "landingType": 1,
                "height": 1230
            },
            "unlockReward": 0,
            "failTrackers": [
                "https://ad.buzzvil.com/api/v2/lockscreen/fail_tracker?data=nvypAq4ZcCO90sLS-wYLHZ8jWrqjKEZXXdv4rduVXAGhz4TncF90USS6t2FlHi8M7d9464GgoDSRzqdZ9kz99bQAHxSk-7UMihE-f3ODtBwVL8EErFW36ExnYx_z6XwQapw0Scbm5D84Xb01ySh_8lqF67ClEXeIh5_qEg51gWGwvxaXRUqmVmhX2slUEjvDA5aRCMLRCyv3KavoaykFjW9O9WCokUtpj52jUu4UgacWd8s_ARP2SXG8WIOSi057FE3Ea4wTakn_CG2kLK2zQq0d5C6Wdxr5u0vNvBpGOCQW987u0y3Y3mzSatVU8LnNAZm6XdYkhJH-WcJuT74Ji9fvARvbkVRN41bmLkeYfpv9IJNvkkAaWDh3T42BS3ZDjLNaqIjMPUuTzULeXyNpOCL55d446P5dAS6o0TBhEGZXw1Kumsa3bqZkALjxZqVi"
            ],
            "clickTrackers": [
                "https://ad.buzzvil.com/api/v2/lockscreen/slide_joy_tracker?data=BO0tBYJxVZexhVbaT9jyGgNncoy-brqQZw286gU2NussKFH8PQH-D0E1PELTCZ5c3rkPGiyGvYYD9eCV5_YazLgU5MKf2wtgiJxXTr39rOQSjCSZx_h2-xzrRh12M_fhwLx-o_ruO6SyyJlIcKCifUjZk6pLrcpmwp71SKTVq8r0yJZj1SqtRGDuLO-9M7amnunE-jqP2xIOMp5YlOpiHUkpoG7p-jRaB7rX13JJL2Rz1HgNu5GRnywORUgDVM2hBKC2BJf9D9a4hSMD79tj1seWAMFMSkzSrnFmWZU2NdAHLg5Ur9a-DBPlFTCCv63W"
            ],
            "actionReward": 0,
            "impressionTrackers": [
                "https://ad.buzzvil.com/api/lockscreen/impression?data=nvypAq4ZcCO90sLS-wYLHZ8jWrqjKEZXXdv4rduVXAGhz4TncF90USS6t2FlHi8M7d9464GgoDSRzqdZ9kz99bQAHxSk-7UMihE-f3ODtBwVL8EErFW36ExnYx_z6XwQapw0Scbm5D84Xb01ySh_8lqF67ClEXeIh5_qEg51gWGwvxaXRUqmVmhX2slUEjvDA5aRCMLRCyv3KavoaykFjW9O9WCokUtpj52jUu4UgacWd8s_ARP2SXG8WIOSi057FE3Ea4wTakn_CG2kLK2zQq0d5C6Wdxr5u0vNvBpGOCQW987u0y3Y3mzSatVU8LnNAZm6XdYkhJH-WcJuT74Ji9fvARvbkVRN41bmLkeYfpv9IJNvkkAaWDh3T42BS3ZDjLNaqIjMPUuTzULeXyNpOCL55d446P5dAS6o0TBhEGZXw1Kumsa3bqZkALjxZqVi",
                "hhttps://ad.buzzvil.com/api/v2/lockscreen/slide_joy_tracker?data=2mljviCZ58fMWps1L7WT_yF5NH9Epfwi8mQqzYpIPAB8fQKVy_DY4i4u1qmCBthj1i1qXGL8O0SeotAll780m6LnV8oyY8He63dZ0Y9gl2hGg3wCVxK3UbTsqvKY2BNMKzzbNUY8egqLKlqB4Op35QbbSvYUK6jHn9eFbHHJbWopZmY7oUl14ilwRZ1VeDfX2rhtP5wvVxeFPNr-Dlx7GNa0GEG8gWD5n_4lBaAEBxVvBZ-LzjGvO6PuujtlXEkzVL01Z4Eeb59-AOy5c3cYPo6-vYT6zB1yAbvKFxdnYa6wl15mhEmGxGrNHWd1AUNm",
                "http://tknet.smardroid.com/agentapi/impression?cid=245648129&aid=5000036&pid=13528&sign=7b706e5e53d0f134d46c65ab78bd4664"
            ]
        },
        {
            "support_webp": true,
            "landingReward": 0,
            "revenueType": "cpm",
            "callToAction": "See more",
            "ttl": 21600,
            "integrationType": "SDK",
            "id": 1502186,
            "network": "OUTBRAIN",
            "name": "outbrain",
            "creative": {
                "support_webp": true,
                "placementId": "SDK_3",
                "adchoiceUrl": null,
                "landingType": 3,
                "clickUrl": "",
                "referrerUrl": "http://www.getslidejoy.com",
                "network": "OUTBRAIN",
                "filterable": true,
                "publisherId": "SLIDE1L0QG0CFL3J7BP1L93J6",
                "bgUrl": "https://buzzvil.akamaized.net/adfit.image/native_ad_bg/3.png",
                "type": "SDK"
            },
            "unlockReward": 0,
            "failTrackers": [
                "https://ad.buzzvil.com/api/v2/lockscreen/fail_tracker?data=nvypAq4ZcCO90sLS-wYLHXzUZF57aE_vtTMrgdg9Tahv9iPr6gzWJtsf4rmNivDhjxSv7M76WYS_oFB2zGFXIknMo_3UR7mYs7H2j10Izvf5NeYbogahRiq_iDJTCxY0vYU_iW89HTqSg0zj2RVbKjC3Ionn-9eJE4yvMKynM7caFGYm76pKZuNy-uMgGC5iZzB-AFNahME5jGz3CYzvh1Inm07NFigYeWLGDSBFjBhDMVClPcYwKlMNgUpmrm4RS64lxpRtjcw0mmAF9BE4RcAQxMvqwKYy3KJGV_3k7cfO6DlxNuyHWipb7M7sostffjDBxciJ00LauqTmZPAeI38mZFnykIQv_vWdc5T_ZoyJqMOMaT7zQT-P3Ud0oCEd4WOgk6DLr2Yfd8oK9Lxr8A%3D%3D"
            ],
            "clickTrackers": [
                "https://ad.buzzvil.com/api/v1/click_track?data=nvypAq4ZcCO90sLS-wYLHXzUZF57aE_vtTMrgdg9Tahv9iPr6gzWJtsf4rmNivDhjxSv7M76WYS_oFB2zGFXIknMo_3UR7mYs7H2j10Izvf5NeYbogahRiq_iDJTCxY0vYU_iW89HTqSg0zj2RVbKjC3Ionn-9eJE4yvMKynM7caFGYm76pKZuNy-uMgGC5iZzB-AFNahME5jGz3CYzvh1Inm07NFigYeWLGDSBFjBhDMVClPcYwKlMNgUpmrm4RS64lxpRtjcw0mmAF9BE4RcAQxMvqwKYy3KJGV_3k7cfO6DlxNuyHWipb7M7sostffjDBxciJ00LauqTmZPAeI38mZFnykIQv_vWdc5T_ZoyJqMOMaT7zQT-P3Ud0oCEd4WOgk6DLr2Yfd8oK9Lxr8A%3D%3D",
                "https://ad.buzzvil.com/api/v2/lockscreen/slide_joy_tracker?data=BO0tBYJxVZexhVbaT9jyGgNncoy-brqQZw286gU2NussKFH8PQH-D0E1PELTCZ5cLCChrUglrAk4sxFdq3IjlJhV0VxZL7w8XM7pgzEL9cMWl0b-OQtJDa68EZ9YuH8zcHLjceiMMHEFzIcr0ij5bA0KjkfgcwEeHq0kIQCJCB1klGnGFvJjcp4uWqgJJtlkPe8Dn4KJIHZfyF3eCRAseu7oFKfHxu8r6XToCtK-fBhWDt0RQDEvDPGOl0djH6fasTUsZwIKqd5u2OyjeOXpu6M0AddPVkOBSXNWUpAM2Zv2kyKIUwscPuWGR4GqJKnI"
            ],
            "actionReward": 0,
            "impressionTrackers": [
                "https://ad.buzzvil.com/api/lockscreen/impression?data=nvypAq4ZcCO90sLS-wYLHXzUZF57aE_vtTMrgdg9Tahv9iPr6gzWJtsf4rmNivDhjxSv7M76WYS_oFB2zGFXIknMo_3UR7mYs7H2j10Izvf5NeYbogahRiq_iDJTCxY0vYU_iW89HTqSg0zj2RVbKjC3Ionn-9eJE4yvMKynM7caFGYm76pKZuNy-uMgGC5iZzB-AFNahME5jGz3CYzvh1Inm07NFigYeWLGDSBFjBhDMVClPcYwKlMNgUpmrm4RS64lxpRtjcw0mmAF9BE4RcAQxMvqwKYy3KJGV_3k7cfO6DlxNuyHWipb7M7sostffjDBxciJ00LauqTmZPAeI38mZFnykIQv_vWdc5T_ZoyJqMOMaT7zQT-P3Ud0oCEd4WOgk6DLr2Yfd8oK9Lxr8A%3D%3D",
                "hhttps://ad.buzzvil.com/api/v2/lockscreen/slide_joy_tracker?data=2mljviCZ58fMWps1L7WT_yF5NH9Epfwi8mQqzYpIPAB8fQKVy_DY4i4u1qmCBthj1i1qXGL8O0SeotAll780m6LnV8oyY8He63dZ0Y9gl2hGg3wCVxK3UbTsqvKY2BNMKzzbNUY8egqLKlqB4Op35QbbSvYUK6jHn9eFbHHJbWopZmY7oUl14ilwRZ1VeDfX2rhtP5wvVxeFPNr-Dlx7GNa0GEG8gWD5n_4lBaAEBxVvBZ-LzjGvO6PuujtlXEkzVL01Z4Eeb59-AOy5c3cYPo6-vYT6zB1yAbvKFxdnYa6wl15mhEmGxGrNHWd1AUNm"
            ]
        }
    ],
    "settings": {
        "filteringWords": "queer,AliExpress"
    }
}
`

func TestMain(m *testing.M) {
	ts = tests.GetTestServer(m)
	tearDownElasticSearch := tests.SetupElasticSearch()
	tearDownDatabase := tests.SetupDatabase()

	code := m.Run()

	tearDownDatabase()
	tearDownElasticSearch()
	tests.DeleteLogFiles()
	ts.Close()

	os.Exit(code)
}
