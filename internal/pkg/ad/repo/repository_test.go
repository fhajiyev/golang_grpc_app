package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/bxcodec/faker"

	networkmock "github.com/Buzzvil/go-test/mock"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const buzzAdURL = "http://ad-dev.buzzvil.com"

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

func (ts *RepoTestSuite) TestGetAdsV1() {
	var req ad.V1AdsRequest
	ts.NoError(faker.FakeData(&req))

	body, err := json.Marshal(buzzAdV1Res)
	ts.NoError(err)

	buzzAdServer := networkmock.NewTargetServer(network.GetHost(buzzAdURL)).AddResponseHandler(&networkmock.ResponseHandler{
		WriteToBody: func() []byte {
			return body
		},
		Path:   "/api/lockscreen/ads",
		Method: http.MethodGet,
	})
	clientPatcher := networkmock.PatchClient(network.DefaultHTTPClient, buzzAdServer)
	defer clientPatcher.RemovePatch()

	res, err := ts.repo.GetAdsV1(req)
	ts.NoError(err)
	ts.Equal(len(buzzAdV1Res.Ads), len(res.Ads))
	ts.Equal(len(buzzAdV1Res.NativeAds), len(res.NativeAds))
	ts.Equal(buzzAdV1Res.Code, res.Code)
	ts.Equal(buzzAdV1Res.Msg, res.Msg)
	ts.Equal(buzzAdV1Res.Settings, res.Settings)
}

func (ts *RepoTestSuite) TestGetAdsV2() {
	var req ad.V2AdsRequest
	ts.NoError(faker.FakeData(&req))

	body, err := json.Marshal(buzzAdV2Res)
	ts.NoError(err)

	buzzAdServer := networkmock.NewTargetServer(network.GetHost(buzzAdURL)).AddResponseHandler(&networkmock.ResponseHandler{
		WriteToBody: func() []byte {
			return body
		},
		Path:   "/api/v2/lockscreen/ads",
		Method: http.MethodGet,
	})
	clientPatcher := networkmock.PatchClient(network.DefaultHTTPClient, buzzAdServer)
	defer clientPatcher.RemovePatch()

	res, err := ts.repo.GetAdsV2(req)
	ts.NoError(err)
	ts.Equal(len(buzzAdV2Res.Ads), len(res.Ads))
	ts.Equal(buzzAdV2Res.Cursor, res.Cursor)
	ts.Equal(buzzAdV2Res.Settings, res.Settings)
}

func (ts *RepoTestSuite) TestGetAdDetailFromCache() {
	detail := ts.createAdDetail()

	key := ts.repo.getCacheKey(detail.ID)
	getCacheMock := func(_ string, obj interface{}) error {
		cache, ok := obj.(*adDetailCache)
		if !ok {
			return errors.New("cannot cast to *adDetailCache")
		}

		cache.CreatedAt = time.Now()
		cache.Detail = *detail
		log.Printf("%+v", cache)
		return nil
	}
	ts.redisCache.On("GetCache", key, mock.AnythingOfType("*repo.adDetailCache")).Return(getCacheMock).Once()

	resDetail, err := ts.repo.GetAdDetail(detail.ID, "")
	ts.NoError(err)
	ts.Equal(detail, resDetail)
	ts.redisCache.AssertExpectations(ts.T())
}

func (ts *RepoTestSuite) TestGetAdDetailFromBuzzAd() {
	detail := ts.createAdDetail()
	modelDetail := &adDetail{
		ID:             detail.ID,
		ItemName:       detail.Name,
		OrganizationID: detail.OrganizationID,
		RevenueType:    detail.RevenueType,
		ExtraData:      detail.Extra,
	}
	body, err := json.Marshal(&modelDetail)
	log.Println(string(body))
	ts.NoError(err)
	httpClient := network.DefaultHTTPClient
	buzzAdServer := networkmock.NewTargetServer(network.GetHost(buzzAdURL)).AddResponseHandler(&networkmock.ResponseHandler{
		WriteToBody: func() []byte {
			return body
		},
		Path:   fmt.Sprintf("/adserver/orders/lineitems/%d", detail.ID),
		Method: http.MethodGet,
	})
	clientPatcher := networkmock.PatchClient(httpClient, buzzAdServer)
	defer clientPatcher.RemovePatch()

	key := ts.repo.getCacheKey(detail.ID)
	ts.redisCache.On("GetCache", key, mock.AnythingOfType("*repo.adDetailCache")).Return(errors.New("failed to get cache")).Once()
	ts.redisCache.On("SetCacheAsync", key, mock.AnythingOfType("*repo.adDetailCache"), time.Hour).Once()

	resDetail, err := ts.repo.GetAdDetail(detail.ID, "")
	ts.NoError(err)
	ts.equalAdDetail(detail, resDetail)
	ts.redisCache.AssertExpectations(ts.T())
}

func (ts *RepoTestSuite) equalAdDetail(expected *ad.Detail, actual *ad.Detail) {
	ts.Equal(expected.ID, actual.ID)
	ts.Equal(expected.OrganizationID, actual.OrganizationID)
	ts.Equal(expected.Name, actual.Name)
	ts.Equal(expected.RevenueType, actual.RevenueType)
}

func (ts *RepoTestSuite) createAdDetail() *ad.Detail {
	return &ad.Detail{
		ID:             rand.Int63n(1000000) + 1,
		Name:           "TEST_AD_NAME",
		OrganizationID: rand.Int63n(1000000) + 1,
		RevenueType:    "cpc",
		Extra: map[string]interface{}{
			"unit": rand.Intn(100000) + 1,
		},
	}
}

type RepoTestSuite struct {
	suite.Suite
	redisCache *MockRedisCache
	repo       *Repository
}

func (ts *RepoTestSuite) SetupTest() {
	ts.redisCache = &MockRedisCache{}
	ts.repo = New(ts.redisCache, buzzAdURL)
}

var _ ad.Repository = &Repository{}

type MockRedisCache struct {
	mock.Mock
}

func (mrc *MockRedisCache) GetCache(key string, obj interface{}) error {
	ret := mrc.Called(key, obj)
	fun, ok := ret.Get(0).(func(string, interface{}) error)
	if ok {
		return fun(key, obj)
	}
	return ret.Error(0)
}

func (mrc *MockRedisCache) SetCacheAsync(key string, obj interface{}, expiration time.Duration) {
	mrc.Called(key, obj, expiration)
}

func (mrc *MockRedisCache) SetCache(key string, obj interface{}, expiration time.Duration) error {
	ret := mrc.Called(key, obj, expiration)
	return ret.Error(0)
}

func (mrc *MockRedisCache) DeleteCache(key string) error {
	ret := mrc.Called(key)
	return ret.Error(0)
}

func stringPointer(s string) *string {
	return &s
}

var buzzAdV1Res = &ad.V1AdsResponse{
	Code: 200,
	Msg:  "ok",
	Settings: &ad.AdsV1Settings{
		FilteringWords: stringPointer("these|are|filtering|words,honeyscreen,cashslide,리워드,reward,성인,도박,카지노,sensitive,cash slide,캐시슬라이드,AliExpress"),
		WebUa:          stringPointer("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"),
	},
	Ads: []*ad.AdV1{
		{
			SupportWebp:           true,
			ClickURL:              "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOEOxhD0jzsi8kQaPV0Fb9LsmFvmmi_TZRp4k2j2pquj8SjQKZa73OTnRxVHrDzVzCQgyOmzbyLY7Dy7yR0qPf_fVArqbHLRzCzMqDIoM8e57wTiHJxFiArqofrGM-oC9wZj7l7tAofH88mmPGaRmi6ZylURcMDOSBUeDjiMJV5z2V0guOiPiTf_9kl6OMMH6G1F5bsHxIvLLquNwiFH8iGdXCYhPaEg1kpHFR66sa0A8f4UjKGGBnIYMhZ8mTn7lTJj4CFfjOP8Fxd280aiCKQIZkvD9wC9Cp5ILW-pQ-OnNxrB8kRzFoQZXzEUoccFjusrGu_M6KaJGAyTdX_ZTimx4s-ZVSL9oGgDKaM5QcN3wXljHXaheJElGmQBFiDGEWE%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
			Extra:                 map[string]interface{}{},
			FirstDisplayWeight:    10000000,
			Image:                 "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
			RemoveAfterImpression: true,
			Ipu:                   9999,
			StartedAt:             1534431600,
			Slot:                  "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
			FirstDisplayPriority:  10,
			AgeFrom:               0,
			UnitPrice:             21,
			ImpressionURLs:        []string{"https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOEOxhD0jzsi8kQaPV0Fb9LsmFvmmi_TZRp4k2j2pquj8SjQKZa73OTnRxVHrDzVzCQgyOmzbyLY7Dy7yR0qPf_fVArqbHLRzCzMqDIoM8e57wTiHJxFiArqofrGM-oC9wZj7l7tAofH88mmPGaRmi6ZylURcMDOSBUeDjiMJV5z2V0guOiPiTf_9kl6OMMH6G1F5bsHxIvLLquNwiFH8iGdXCYhPaEg1kpHFR66sa0A8f4UjKGGBnIYMhZ8mTn7lTJj4CFfjOP8Fxd280aiCKQIZkvD9wC9Cp5ILW-pQ-OnNxrB8kRzFoQZXzEUoccFjusrGu_M6KaJGAyTdX_ZTimx4s-ZVSL9oGgDKaM5QcN3wXljHXaheJElGmQBFiDGEWE%3D"},
			ActionReward:          0,
			UnlockReward:          0,
			Type:                  "cpc",
			UseWebUa:              false,
			ImageIos:              "",
			Dipu:                  9999,
			OrganizationID:        1,
			AdNetworkID:           0,
			Name:                  "Hmall::",
			LandingReward:         1,
			Sex:                   "",
			PreferredBrowser:      nil,
			ID:                    1383378,
			AgeTo:                 0,
			IsRtb:                 false,
			Creative: map[string]interface{}{
				"landing_type": "browser",
				"support_webp": true,
				"click_url":    "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOEOxhD0jzsi8kQaPV0Fb9LsmFvmmi_TZRp4k2j2pquj8SjQKZa73OTnRxVHrDzVzCQgyOmzbyLY7Dy7yR0qPf_fVArqbHLRzCzMqDIoM8e57wTiHJxFiArqofrGM-oC9wZj7l7tAofH88mmPGaRmi6ZylURcMDOSBUeDjiMJV5z2V0guOiPiTf_9kl6OMMH6G1F5bsHxIvLLquNwiFH8iGdXCYhPaEg1kpHFR66sa0A8f4UjKGGBnIYMhZ8mTn7lTJj4CFfjOP8Fxd280aiCKQIZkvD9wC9Cp5ILW-pQ-OnNxrB8kRzFoQZXzEUoccFjusrGu_M6KaJGAyTdX_ZTimx4s-ZVSL9oGgDKaM5QcN3wXljHXaheJElGmQBFiDGEWE%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
				"filterable":   true,
				"image":        "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
				"adchoice_url": nil,
				"type":         "IMAGE",
			},
			IsIncentive:  true,
			Tipu:         0,
			OwnerID:      2147,
			TargetApp:    "",
			DisplayType:  "A",
			Icon:         "",
			LandingType:  "browser",
			EndedAt:      1535080173,
			ClickBeacons: []string{"http://abc.def/adfwef?adef=1383378&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=Hmall%3A%3A&abc="},
			Region:       "",
			DeviceName:   "",
			Carrier:      "",
		},
		{
			SupportWebp:           true,
			ClickURL:              "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
			Extra:                 map[string]interface{}{},
			FirstDisplayWeight:    10000000,
			Image:                 "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
			RemoveAfterImpression: true,
			Ipu:                   9999,
			StartedAt:             1534431600,
			Slot:                  "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
			FirstDisplayPriority:  10,
			AgeFrom:               0,
			UnitPrice:             21,
			ImpressionURLs:        []string{"https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D"},
			ActionReward:          0,
			UnlockReward:          0,
			Type:                  "cpc",
			UseWebUa:              false,
			ImageIos:              "",
			Dipu:                  9999,
			OrganizationID:        1,
			AdNetworkID:           0,
			Name:                  "홈앤쇼핑::",
			LandingReward:         1,
			Sex:                   "",
			PreferredBrowser:      nil,
			ID:                    1383380,
			AgeTo:                 0,
			IsRtb:                 false,
			Creative: map[string]interface{}{
				"landing_type": "browser",
				"support_webp": true,
				"click_url":    "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
				"filterable":   true,
				"image":        "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
				"adchoice_url": nil,
				"type":         "IMAGE",
			},
			IsIncentive:  true,
			Tipu:         0,
			OwnerID:      2147,
			TargetApp:    "",
			DisplayType:  "A",
			Icon:         "",
			LandingType:  "browser",
			EndedAt:      1535080173,
			ClickBeacons: []string{"http://abc.def/adfwef?adef=1383378&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=Hmall%3A%3A&abc="},
			Region:       "",
			DeviceName:   "",
			Carrier:      "",
		},
		{
			SupportWebp:           true,
			ClickURL:              "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
			Extra:                 map[string]interface{}{},
			FirstDisplayWeight:    10000000,
			Image:                 "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
			RemoveAfterImpression: true,
			Ipu:                   9999,
			StartedAt:             1534431600,
			Slot:                  "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
			FirstDisplayPriority:  10,
			AgeFrom:               0,
			UnitPrice:             21,
			ImpressionURLs:        []string{"https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D"},
			ActionReward:          0,
			UnlockReward:          0,
			Type:                  "cpc",
			UseWebUa:              false,
			ImageIos:              "",
			Dipu:                  9999,
			OrganizationID:        1,
			AdNetworkID:           0,
			Name:                  "홈앤쇼핑::",
			LandingReward:         1,
			Sex:                   "",
			PreferredBrowser:      nil,
			ID:                    1383380,
			AgeTo:                 0,
			IsRtb:                 false,
			Creative: map[string]interface{}{
				"landing_type": "browser",
				"support_webp": true,
				"click_url":    "https://ad-dev.buzzvil.com/api/v1/click?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOFTRZX6WuSJf-shCcx2EF3KL3htBxXt41sIIe6bzBSxiI0OPdxbw11dKvFHWyhgHWAAXjNd0b3Wrqj8dDlO1cKdGpICTTck0FQa9N4kEQNt2-yTl4I6xyeNzWQt5OnlHLgd9Qkm3MNoadc4XwMZ9RcRXJQEyiqHziyIsd4rgvrBcP7CBCziYU3iw3hBOo6vDBngvVUaEERmndxB8dCNlPkGBAGRTQ_DJn1cTOEa85trv42jfc2Wgwu_CnSLkl2QgF7AqThiK8pPUVJqXVds4STjxEuCttmza8LpK0jqHEYR4nNqEFof17IP2Uoo9Du2dDtzjGnWQbVR-7CPuzXB4ff78RAsPZuVnejim17qc8SBIBCjBZ-QdYnfDwL6tJqZiLM%3D&redirect_url=https%3A%2F%2Fwww.buzzvil.com&direct=1",
				"filterable":   true,
				"image":        "https://buzzvil.akamaized.net/adfit.image/uploads/1534482576-MI7YU.jpg",
				"adchoice_url": nil,
				"type":         "IMAGE",
			},
			IsIncentive:  true,
			Tipu:         0,
			OwnerID:      2147,
			TargetApp:    "",
			DisplayType:  "A",
			Icon:         "",
			LandingType:  "browser",
			EndedAt:      1535080173,
			ClickBeacons: []string{"http://abc.def/adfwef?adef=1383378&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=Hmall%3A%3A&abc="},
			Region:       "",
			DeviceName:   "",
			Carrier:      "",
			Events: ad.Events{
				{
					TrackingURLs: []string{"http://tracking.url.com"},
					Type:         "landed",
					Reward: &ad.Reward{
						Amount: 1,
						Status: "RECEIVABLE",
					},
				},
			},
		},
	},
	NativeAds: []*ad.NativeAdV1{
		{
			AdV1: ad.AdV1{
				SupportWebp:           false,
				ClickURL:              "",
				Extra:                 map[string]interface{}{},
				FirstDisplayWeight:    10000000,
				Image:                 "https://cdn-ad-static.buzzvil.com/native_ad_bg/fan_bg_3.jpg",
				RemoveAfterImpression: false,
				Ipu:                   1,
				StartedAt:             1493264640,
				Slot:                  "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
				FirstDisplayPriority:  10,
				AgeFrom:               0,
				UnitPrice:             0,
				ImpressionURLs:        []string{"https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOHSDKETb7HA31RGtD3AC-dhcQArUCZDDydCFRcScFNvnlN_YL_I7-6q2Xn2BqFx8o1EU2Rsq3lvc-gsU3xVDKhlDiZDORayOh1ASUx7Aw9oV1MIkOyts_sqDnmjovYsyfB0GdkldbT28ekJy9yplTwYa4fdsfZCKIq8FNmONvGa8zXT7GMIj2ngdtzNFDh9lGMEHhqE2ycQ25CifKU4_sXQKZdabcyV1gZ6uFj-BRzTRk80ho_QSgpeaclDmimI-huST1k73R1lRV7xs1bOGXZRaUjlptKMH_P9xlGxCwI0hyq8syzdOrcMoClQq_V0ZowaPKeWAfytXFrI9qMRzs7rXDMnDQ7A-VSg6MLrlTswHjgzqnUr-oTvRfxs0Vbswng%3D"},
				ActionReward:          0,
				UnlockReward:          0,
				Type:                  "cpc",
				UseWebUa:              false,
				ImageIos:              "",
				Dipu:                  9999,
				OrganizationID:        1,
				AdNetworkID:           519,
				Name:                  "DDNtest",
				LandingReward:         0,
				Sex:                   "",
				PreferredBrowser:      nil,
				ID:                    1381015,
				AgeTo:                 0,
				IsRtb:                 false,
				Creative: map[string]interface{}{
					"support_webp": false,
					"placement_id": "",
					"period":       int64(7200),
					"background":   "https://buzzvil.akamaized.net/adfit.image/native_ad_bg/8.png",
					"lifetime":     int64(3600),
					"id":           "DAN-urumlgivnrdw",
					"network":      "ADFIT",
					"name":         "af",
					"filterable":   false,
					"type":         "BANNER_SDK",
					"referrer_url": "",
					"adchoice_url": nil,
					"publisher_id": "",
				},
				IsIncentive: false,
				Tipu:        0,
				OwnerID:     425,
				TargetApp:   "",
				DisplayType: "A",
				Icon:        "",
				LandingType: "browser",
				EndedAt:     1535080173,
				ClickBeacons: []string{
					"https://ad-dev.buzzvil.com/api/v1/click_track?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOHSDKETb7HA31RGtD3AC-dhcQArUCZDDydCFRcScFNvnlN_YL_I7-6q2Xn2BqFx8o1EU2Rsq3lvc-gsU3xVDKhlDiZDORayOh1ASUx7Aw9oV1MIkOyts_sqDnmjovYsyfB0GdkldbT28ekJy9yplTwYa4fdsfZCKIq8FNmONvGa8zXT7GMIj2ngdtzNFDh9lGMEHhqE2ycQ25CifKU4_sXQKZdabcyV1gZ6uFj-BRzTRk80ho_QSgpeaclDmimI-huST1k73R1lRV7xs1bOGXZRaUjlptKMH_P9xlGxCwI0hyq8syzdOrcMoClQq_V0ZowaPKeWAfytXFrI9qMRzs7rXDMnDQ7A-VSg6MLrlTswHjgzqnUr-oTvRfxs0Vbswng%3D",
					"http://abc.def/adfwef?adef=1381015&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=DDN%20test&abc=",
				},
				Region:     "",
				DeviceName: "",
				Carrier:    "",
			},
			Banner: &ad.NativeAdV1Settings{
				PlacementID: "",
				Period:      7200,
				Background:  "https://buzzvil.akamaized.net/adfit.image/native_ad_bg/8.png",
				LifeTime:    3600,
				ID:          "DAN-urumlgivnrdw",
				Network:     "ADFIT",
				Name:        "af",
				ReferrerURL: "",
				PublisherID: "",
			},
		},
		{
			AdV1: ad.AdV1{
				SupportWebp:           true,
				ClickURL:              "",
				Extra:                 map[string]interface{}{},
				FirstDisplayWeight:    41,
				Image:                 "https://buzzvil.akamaized.net/adfit.image/uploads/1492063831-UL5J6.jpg",
				RemoveAfterImpression: false,
				Ipu:                   1,
				StartedAt:             1493264640,
				Slot:                  "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23",
				FirstDisplayPriority:  10,
				AgeFrom:               0,
				UnitPrice:             0,
				ImpressionURLs:        []string{"https://ad-dev.buzzvil.com/api/lockscreen/impression?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOHSDKETb7HA31RGtD3AC-dhcQArUCZDDydCFRcScFNvnlN_YL_I7-6q2Xn2BqFx8o1EU2Rsq3lvc-gsU3xVDKhlDiZDORayOh1ASUx7Aw9oV1MIkOyts_sqDnmjovYsyfB0GdkldbT28ekJy9yplTwYa4fdsfZCKIq8FNmONvGa8zXT7GMIj2ngdtzNFDh9lGMEHhqE2ycQ25CifKU4_sXQKZdabcyV1gZ6uFj-BRzTRk80ho_QSgpeaclDmimI-huST1k73R1lRV7xs1bOGXZRaUjlptKMH_P9xlGxCwI0hyq8syzdOrcMoClQq_V0ZowaPKeWAfytXFrI9qMRzs7rXDMnDQ7A-VSg6MLrlTswHjgzqnUr-oTvRfxs0Vbswng%3D"},
				ActionReward:          0,
				UnlockReward:          0,
				Type:                  "cpc",
				UseWebUa:              false,
				ImageIos:              "",
				Dipu:                  9999,
				OrganizationID:        1,
				AdNetworkID:           503,
				Name:                  "mobvista",
				LandingReward:         0,
				Sex:                   "",
				PreferredBrowser:      nil,
				ID:                    1225041,
				AgeTo:                 0,
				IsRtb:                 false,
				Creative: map[string]interface{}{
					"support_webp": true,
					"placement_id": "",
					"period":       int64(14400),
					"background":   "https://buzzvil.akamaized.net/adfit.image/native_ad_bg/7.png",
					"lifetime":     int64(3600),
					"id":           "11230",
					"network":      "MOBVISTA",
					"name":         "mv",
					"filterable":   true,
					"type":         "SDK",
					"referrer_url": "",
					"adchoice_url": nil,
					"publisher_id": "",
				},
				IsIncentive: false,
				Tipu:        0,
				OwnerID:     704,
				TargetApp:   "",
				DisplayType: "A",
				Icon:        "",
				LandingType: "browser",
				EndedAt:     1535080173,
				ClickBeacons: []string{
					"https://ad-dev.buzzvil.com/api/v1/click_track?data=dBec7tf6gN70M2eDwQLlpLwJaus6HUG3RSPfdqSSrOHSDKETb7HA31RGtD3AC-dhcQArUCZDDydCFRcScFNvnlN_YL_I7-6q2Xn2BqFx8o1EU2Rsq3lvc-gsU3xVDKhlDiZDORayOh1ASUx7Aw9oV1MIkOyts_sqDnmjovYsyfB0GdkldbT28ekJy9yplTwYa4fdsfZCKIq8FNmONvGa8zXT7GMIj2ngdtzNFDh9lGMEHhqE2ycQ25CifKU4_sXQKZdabcyV1gZ6uFj-BRzTRk80ho_QSgpeaclDmimI-huST1k73R1lRV7xs1bOGXZRaUjlptKMH_P9xlGxCwI0hyq8syzdOrcMoClQq_V0ZowaPKeWAfytXFrI9qMRzs7rXDMnDQ7A-VSg6MLrlTswHjgzqnUr-oTvRfxs0Vbswng%3D",
					"http://abc.def/adfwef?adef=1381015&wefwef=a2f2a38a-398f-4b46-9c1e-e0d344408790&title=DDN%20test&abc=",
				},
				Region:     "",
				DeviceName: "",
				Carrier:    "",
			},
			Banner: &ad.NativeAdV1Settings{
				PlacementID: "",
				Period:      14400,
				Background:  "https://buzzvil.akamaized.net/adfit.image/native_ad_bg/7.png",
				LifeTime:    3000,
				ID:          "11230",
				Network:     "MOBVISTA",
				Name:        "mv",
				ReferrerURL: "",
				PublisherID: "",
			},
		},
	},
}

var buzzAdV2Res = &ad.V2AdsResponse{
	Cursor: "PlepwHVRoKbbja3OMeUSgJGyPzS9DkYyTK-bPRRTUfudsf1CbIL7c0mVtMaNr32HTulhetlSzN6nLFYT4bwBzg==",
	Settings: &ad.AdsV2Settings{
		"filteringWords": "these|are|filtering|words,honeyscreen,cashslide,리워드,reward,성인,도박,카지노,sensitive,cash slide,캐시슬라이드,AliExpress",
		"webUa":          "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5 Build/MOB30H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.98 Mobile Safari/537.36",
	},
	Ads: []ad.AdV2{
		{
			Extra:        map[string]interface{}{},
			RewardPeriod: 7200,
			Meta:         map[string]interface{}{},
			CallToAction: "더보기",
			TTL:                nil,
			ID:                 1744152,
			Network:            nil,
			ConversionCheckURL: nil,
			Creative: map[string]interface{}{
				"support_webp": true,
				"adchoiceUrl":  nil,
				"landingType":  1,
				"clickUrl":     "https://ad-dev.buzzvil.com/api/v1/click?position=__position__&ifv=__ifv__&data=hN4E1R2kfZ1SGPiSQZB3R2aQtSrJOS-X4dXsLN0zemAnGR3yGKASogX2zG2Z-EEEPiJ0r23CEZH5_HeAV7ZWjw4ejhcgesMHHM_qw3RmhDw5y6eAJzawcpDZ1Oh5cL6M0u2NV7kfKq0ztHtXw4lzpR6CkGhBeUfiJ61o9eWFp4kSXhb3ZsAvWpwoQhd60QZdj-QAh0t9LpA577t4-rsEPGffDchViWUVrC1pVlyeMdGucGHz0E75WvCq29YIxRy886kGdoMmmqKroTW_LZR4hfB0LXL097Kt9uMB5sbYGFK7TRUT8_HpCFdnn3xRKVxyPehVQ_2wZ27-EhM4r5b7LQ40mlQcItDHyQh55CQ-fnwYVQK5irM1o1YHX4tcOrkqCF-reR5z5NY5RwSEb2-VkSA8DU6fJcXj7oNvtYalzYV_8_PWuCbuXZvtZJssjBfOHjzwLU65y4si0LdwD8Csu-uMUHvVDaxEOP14HHIqCdiO0h9hjFp7iKgo36WfK2M_oSAhuI_XGODFs7WIfbfDg0cWkl1GYeQ_MQnACQUPKUOGjWl8onqAL7Si8b3izVeYKZR8HJ5Ime0aFFgf3KUOgroUNNq1avPjmu9N4E5MFNebNRKCaiSQdywMhIIfyQnnr7xpqwZWfV-ofYOhToAL8PqIo-1Rg8ZxdGQ-zWL_OPISw4WE-nn1BdTV7VtTI1e50SGwifpyP5i_Vdnf5THEYgwu8ONTJf-nAKXERbPPmTc%3D&session_id=__session_id__&redirect_url=https%3A%2F%2Fiag-0.tlnk.io%2Fserve%3Faction%3Dclick%26agency_id%3D978%26campaign_id_android%3D366628%26destination_id_android%3D525170%26google_aid%3D%7Badid%7D%26publisher_id%3D339430%26ref_id%3D%7Bbz_tracking_id%7D%26site_id_android%3D98222%26sub_publisher%3D%7Bunit_id%7D%26my_campaign%3Dmass%26my_publisher%3D2020_jul%26my_site%3Dhoneyscreen_M_da%26my_adgroup%3Dbehavior_11stapp%26my_keyword%3Dpromo%26my_placement%3D1001404678&direct=1",
				"height":       1230,
				"callToAction": "더보기",
				"sizeType":     "FULLSCREEN",
				"filterable":   true,
				"imageUrl":     "https://buzzvil.akamaized.net/adfit.image/uploads/1592295997-RS4FM.jpg",
				"width":        720,
				"type":         "IMAGE",
				"is_deeplink":  false,
			},
			UnlockReward:       0,
			FailTrackers:       []string{"https://ad-dev.buzzvil.com/api/lockscreen/fail_tracker?position=__position__&ifv=__ifv__&data=hN4E1R2kfZ1SGPiSQZB3R2aQtSrJOS-X4dXsLN0zemAnGR3yGKASogX2zG2Z-EEEPiJ0r23CEZH5_HeAV7ZWjw4ejhcgesMHHM_qw3RmhDw5y6eAJzawcpDZ1Oh5cL6M0u2NV7kfKq0ztHtXw4lzpR6CkGhBeUfiJ61o9eWFp4kSXhb3ZsAvWpwoQhd60QZdj-QAh0t9LpA577t4-rsEPGffDchViWUVrC1pVlyeMdGucGHz0E75WvCq29YIxRy886kGdoMmmqKroTW_LZR4hfB0LXL097Kt9uMB5sbYGFK7TRUT8_HpCFdnn3xRKVxyPehVQ_2wZ27-EhM4r5b7LQ40mlQcItDHyQh55CQ-fnwYVQK5irM1o1YHX4tcOrkqCF-reR5z5NY5RwSEb2-VkSA8DU6fJcXj7oNvtYalzYV_8_PWuCbuXZvtZJssjBfOHjzwLU65y4si0LdwD8Csu-uMUHvVDaxEOP14HHIqCdiO0h9hjFp7iKgo36WfK2M_oSAhuI_XGODFs7WIfbfDg0cWkl1GYeQ_MQnACQUPKUOGjWl8onqAL7Si8b3izVeYKZR8HJ5Ime0aFFgf3KUOgroUNNq1avPjmu9N4E5MFNebNRKCaiSQdywMhIIfyQnnr7xpqwZWfV-ofYOhToAL8PqIo-1Rg8ZxdGQ-zWL_OPISw4WE-nn1BdTV7VtTI1e50SGwifpyP5i_Vdnf5THEYgwu8ONTJf-nAKXERbPPmTc%3D&session_id=__session_id__"},
			ImpressionTrackers: []string{"https://ad-dev.buzzvil.com/api/lockscreen/impression?position=__position__&ifv=__ifv__&data=hN4E1R2kfZ1SGPiSQZB3R2aQtSrJOS-X4dXsLN0zemAnGR3yGKASogX2zG2Z-EEEPiJ0r23CEZH5_HeAV7ZWjw4ejhcgesMHHM_qw3RmhDw5y6eAJzawcpDZ1Oh5cL6M0u2NV7kfKq0ztHtXw4lzpR6CkGhBeUfiJ61o9eWFp4kSXhb3ZsAvWpwoQhd60QZdj-QAh0t9LpA577t4-rsEPGffDchViWUVrC1pVlyeMdGucGHz0E75WvCq29YIxRy886kGdoMmmqKroTW_LZR4hfB0LXL097Kt9uMB5sbYGFK7TRUT8_HpCFdnn3xRKVxyPehVQ_2wZ27-EhM4r5b7LQ40mlQcItDHyQh55CQ-fnwYVQK5irM1o1YHX4tcOrkqCF-reR5z5NY5RwSEb2-VkSA8DU6fJcXj7oNvtYalzYV_8_PWuCbuXZvtZJssjBfOHjzwLU65y4si0LdwD8Csu-uMUHvVDaxEOP14HHIqCdiO0h9hjFp7iKgo36WfK2M_oSAhuI_XGODFs7WIfbfDg0cWkl1GYeQ_MQnACQUPKUOGjWl8onqAL7Si8b3izVeYKZR8HJ5Ime0aFFgf3KUOgroUNNq1avPjmu9N4E5MFNebNRKCaiSQdywMhIIfyQnnr7xpqwZWfV-ofYOhToAL8PqIo-1Rg8ZxdGQ-zWL_OPISw4WE-nn1BdTV7VtTI1e50SGwifpyP5i_Vdnf5THEYgwu8ONTJf-nAKXERbPPmTc%3D&session_id=__session_id__"},
			LandingReward:      1,
			AdReportData:       nil,
			PreferredBrowser:   nil,
			Name:               "11번가",
			OrganizationID:     1,
			RevenueType:        "cpm",
			OwnerID:            6841,
			ClickTrackers:      []string{},
			ActionReward:       0,
		},
		{
			Extra:        map[string]interface{}{},
			RewardPeriod: 7200,
			Meta:         map[string]interface{}{},
			CallToAction: "더보기",
			TTL:                nil,
			ID:                 1383380,
			Network:            nil,
			ConversionCheckURL: nil,
			Creative: map[string]interface{}{
				"support_webp": true,
				"adchoiceUrl":  nil,
				"landingType":  1,
				"clickUrl":     "https://ad-dev.buzzvil.com/api/v1/click?position=__position__&ifv=__ifv__&data=hN4E1R2kfZ1SGPiSQZB3R2aQtSrJOS-X4dXsLN0zemAnGR3yGKASogX2zG2Z-EEEPiJ0r23CEZH5_HeAV7ZWjw4ejhcgesMHHM_qw3RmhDw5y6eAJzawcpDZ1Oh5cL6M0u2NV7kfKq0ztHtXw4lzpR6CkGhBeUfiJ61o9eWFp4kSXhb3ZsAvWpwoQhd60QZdj-QAh0t9LpA577t4-rsEPGffDchViWUVrC1pVlyeMdGucGHz0E75WvCq29YIxRy886kGdoMmmqKroTW_LZR4hfB0LXL097Kt9uMB5sbYGFK7TRUT8_HpCFdnn3xRKVxyPehVQ_2wZ27-EhM4r5b7LQ40mlQcItDHyQh55CQ-fnwYVQK5irM1o1YHX4tcOrkqCF-reR5z5NY5RwSEb2-VkSA8DU6fJcXj7oNvtYalzYV_8_PWuCbuXZvtZJssjBfOHjzwLU65y4si0LdwD8Csu-uMUHvVDaxEOP14HHIqCdiO0h9hjFp7iKgo36WfK2M_oSAhuI_XGODFs7WIfbfDg0cWkl1GYeQ_MQnACQUPKUOGjWl8onqAL7Si8b3izVeYKZR8HJ5Ime0aFFgf3KUOgroUNNq1avPjmu9N4E5MFNebNRKCaiSQdywMhIIfyQnnr7xpqwZWfV-ofYOhToAL8PqIo-1Rg8ZxdGQ-zWL_OPISw4WE-nn1BdTV7VtTI1e50SGwifpyP5i_Vdnf5THEYgwu8ONTJf-nAKXERbPPmTc%3D&session_id=__session_id__&redirect_url=https%3A%2F%2Fiag-0.tlnk.io%2Fserve%3Faction%3Dclick%26agency_id%3D978%26campaign_id_android%3D366628%26destination_id_android%3D525170%26google_aid%3D%7Badid%7D%26publisher_id%3D339430%26ref_id%3D%7Bbz_tracking_id%7D%26site_id_android%3D98222%26sub_publisher%3D%7Bunit_id%7D%26my_campaign%3Dmass%26my_publisher%3D2020_jul%26my_site%3Dhoneyscreen_M_da%26my_adgroup%3Dbehavior_11stapp%26my_keyword%3Dpromo%26my_placement%3D1001404678&direct=1",
				"height":       1230,
				"callToAction": "더보기",
				"sizeType":     "FULLSCREEN",
				"filterable":   true,
				"imageUrl":     "https://buzzvil.akamaized.net/adfit.image/uploads/1592295997-RS4FM.jpg",
				"width":        720,
				"type":         "IMAGE",
				"is_deeplink":  false,
			},
			UnlockReward:       0,
			FailTrackers:       []string{"https://ad-dev.buzzvil.com/api/lockscreen/fail_tracker?position=__position__&ifv=__ifv__&data=hN4E1R2kfZ1SGPiSQZB3R2aQtSrJOS-X4dXsLN0zemAnGR3yGKASogX2zG2Z-EEEPiJ0r23CEZH5_HeAV7ZWjw4ejhcgesMHHM_qw3RmhDw5y6eAJzawcpDZ1Oh5cL6M0u2NV7kfKq0ztHtXw4lzpR6CkGhBeUfiJ61o9eWFp4kSXhb3ZsAvWpwoQhd60QZdj-QAh0t9LpA577t4-rsEPGffDchViWUVrC1pVlyeMdGucGHz0E75WvCq29YIxRy886kGdoMmmqKroTW_LZR4hfB0LXL097Kt9uMB5sbYGFK7TRUT8_HpCFdnn3xRKVxyPehVQ_2wZ27-EhM4r5b7LQ40mlQcItDHyQh55CQ-fnwYVQK5irM1o1YHX4tcOrkqCF-reR5z5NY5RwSEb2-VkSA8DU6fJcXj7oNvtYalzYV_8_PWuCbuXZvtZJssjBfOHjzwLU65y4si0LdwD8Csu-uMUHvVDaxEOP14HHIqCdiO0h9hjFp7iKgo36WfK2M_oSAhuI_XGODFs7WIfbfDg0cWkl1GYeQ_MQnACQUPKUOGjWl8onqAL7Si8b3izVeYKZR8HJ5Ime0aFFgf3KUOgroUNNq1avPjmu9N4E5MFNebNRKCaiSQdywMhIIfyQnnr7xpqwZWfV-ofYOhToAL8PqIo-1Rg8ZxdGQ-zWL_OPISw4WE-nn1BdTV7VtTI1e50SGwifpyP5i_Vdnf5THEYgwu8ONTJf-nAKXERbPPmTc%3D&session_id=__session_id__"},
			ImpressionTrackers: []string{"https://ad-dev.buzzvil.com/api/lockscreen/impression?position=__position__&ifv=__ifv__&data=hN4E1R2kfZ1SGPiSQZB3R2aQtSrJOS-X4dXsLN0zemAnGR3yGKASogX2zG2Z-EEEPiJ0r23CEZH5_HeAV7ZWjw4ejhcgesMHHM_qw3RmhDw5y6eAJzawcpDZ1Oh5cL6M0u2NV7kfKq0ztHtXw4lzpR6CkGhBeUfiJ61o9eWFp4kSXhb3ZsAvWpwoQhd60QZdj-QAh0t9LpA577t4-rsEPGffDchViWUVrC1pVlyeMdGucGHz0E75WvCq29YIxRy886kGdoMmmqKroTW_LZR4hfB0LXL097Kt9uMB5sbYGFK7TRUT8_HpCFdnn3xRKVxyPehVQ_2wZ27-EhM4r5b7LQ40mlQcItDHyQh55CQ-fnwYVQK5irM1o1YHX4tcOrkqCF-reR5z5NY5RwSEb2-VkSA8DU6fJcXj7oNvtYalzYV_8_PWuCbuXZvtZJssjBfOHjzwLU65y4si0LdwD8Csu-uMUHvVDaxEOP14HHIqCdiO0h9hjFp7iKgo36WfK2M_oSAhuI_XGODFs7WIfbfDg0cWkl1GYeQ_MQnACQUPKUOGjWl8onqAL7Si8b3izVeYKZR8HJ5Ime0aFFgf3KUOgroUNNq1avPjmu9N4E5MFNebNRKCaiSQdywMhIIfyQnnr7xpqwZWfV-ofYOhToAL8PqIo-1Rg8ZxdGQ-zWL_OPISw4WE-nn1BdTV7VtTI1e50SGwifpyP5i_Vdnf5THEYgwu8ONTJf-nAKXERbPPmTc%3D&session_id=__session_id__"},
			LandingReward:      1,
			AdReportData:       nil,
			PreferredBrowser:   nil,
			Name:               "OUTBRAIN",
			OrganizationID:     1,
			RevenueType:        "cpm",
			OwnerID:            6841,
			ClickTrackers:      []string{},
			ActionReward:       0,
		},
		{
			Extra:              map[string]interface{}{},
			RewardPeriod:       7200,
			Meta:               map[string]interface{}{},
			CallToAction:       "もっと見る",
			TTL:                nil,
			ID:                 1641713,
			Network:            nil,
			ConversionCheckURL: nil,
			Creative: map[string]interface{}{
				"support_webp": true,
				"adchoiceUrl":  nil,
				"landingType":  1,
				"clickUrl":     "https://ad-dev.buzzvil.com/api/v1/click?position=__position__&ifv=__ifv__&data=hN4E1R2kfZ1SGPiSQZB3R2aQtSrJOS-X4dXsLN0zemAnGR3yGKASogX2zG2Z-EEEPiJ0r23CEZH5_HeAV7ZWjw4ejhcgesMHHM_qw3RmhDw5y6eAJzawcpDZ1Oh5cL6M0u2NV7kfKq0ztHtXw4lzpR6CkGhBeUfiJ61o9eWFp4kSXhb3ZsAvWpwoQhd60QZdj-QAh0t9LpA577t4-rsEPGffDchViWUVrC1pVlyeMdGucGHz0E75WvCq29YIxRy886kGdoMmmqKroTW_LZR4hfB0LXL097Kt9uMB5sbYGFK7TRUT8_HpCFdnn3xRKVxyPehVQ_2wZ27-EhM4r5b7LQ40mlQcItDHyQh55CQ-fnwYVQK5irM1o1YHX4tcOrkqCF-reR5z5NY5RwSEb2-VkSA8DU6fJcXj7oNvtYalzYV_8_PWuCbuXZvtZJssjBfOHjzwLU65y4si0LdwD8Csu-uMUHvVDaxEOP14HHIqCdiO0h9hjFp7iKgo36WfK2M_oSAhuI_XGODFs7WIfbfDg0cWkl1GYeQ_MQnACQUPKUOGjWl8onqAL7Si8b3izVeYKZR8HJ5Ime0aFFgf3KUOgroUNNq1avPjmu9N4E5MFNebNRKCaiSQdywMhIIfyQnnr7xpqwZWfV-ofYOhToAL8PqIo-1Rg8ZxdGQ-zWL_OPISw4WE-nn1BdTV7VtTI1e50SGwifpyP5i_Vdnf5THEYgwu8ONTJf-nAKXERbPPmTc%3D&session_id=__session_id__&redirect_url=https%3A%2F%2Fiag-0.tlnk.io%2Fserve%3Faction%3Dclick%26agency_id%3D978%26campaign_id_android%3D366628%26destination_id_android%3D525170%26google_aid%3D%7Badid%7D%26publisher_id%3D339430%26ref_id%3D%7Bbz_tracking_id%7D%26site_id_android%3D98222%26sub_publisher%3D%7Bunit_id%7D%26my_campaign%3Dmass%26my_publisher%3D2020_jul%26my_site%3Dhoneyscreen_M_da%26my_adgroup%3Dbehavior_11stapp%26my_keyword%3Dpromo%26my_placement%3D1001404678&direct=1",
				"height":       1230,
				"callToAction": "더보기",
				"sizeType":     "FULLSCREEN",
				"filterable":   true,
				"imageUrl":     "https://buzzvil.akamaized.net/adfit.image/uploads/1592295997-RS4FM.jpg",
				"width":        720,
				"type":         "IMAGE",
				"is_deeplink":  false,
			},
			UnlockReward:       0,
			FailTrackers:       []string{"https://ad-dev.buzzvil.com/api/lockscreen/fail_tracker?position=__position__&ifv=__ifv__&data=hN4E1R2kfZ1SGPiSQZB3R2aQtSrJOS-X4dXsLN0zemAnGR3yGKASogX2zG2Z-EEEPiJ0r23CEZH5_HeAV7ZWjw4ejhcgesMHHM_qw3RmhDw5y6eAJzawcpDZ1Oh5cL6M0u2NV7kfKq0ztHtXw4lzpR6CkGhBeUfiJ61o9eWFp4kSXhb3ZsAvWpwoQhd60QZdj-QAh0t9LpA577t4-rsEPGffDchViWUVrC1pVlyeMdGucGHz0E75WvCq29YIxRy886kGdoMmmqKroTW_LZR4hfB0LXL097Kt9uMB5sbYGFK7TRUT8_HpCFdnn3xRKVxyPehVQ_2wZ27-EhM4r5b7LQ40mlQcItDHyQh55CQ-fnwYVQK5irM1o1YHX4tcOrkqCF-reR5z5NY5RwSEb2-VkSA8DU6fJcXj7oNvtYalzYV_8_PWuCbuXZvtZJssjBfOHjzwLU65y4si0LdwD8Csu-uMUHvVDaxEOP14HHIqCdiO0h9hjFp7iKgo36WfK2M_oSAhuI_XGODFs7WIfbfDg0cWkl1GYeQ_MQnACQUPKUOGjWl8onqAL7Si8b3izVeYKZR8HJ5Ime0aFFgf3KUOgroUNNq1avPjmu9N4E5MFNebNRKCaiSQdywMhIIfyQnnr7xpqwZWfV-ofYOhToAL8PqIo-1Rg8ZxdGQ-zWL_OPISw4WE-nn1BdTV7VtTI1e50SGwifpyP5i_Vdnf5THEYgwu8ONTJf-nAKXERbPPmTc%3D&session_id=__session_id__"},
			ImpressionTrackers: []string{"https://ad-dev.buzzvil.com/api/lockscreen/impression?position=__position__&ifv=__ifv__&data=hN4E1R2kfZ1SGPiSQZB3R2aQtSrJOS-X4dXsLN0zemAnGR3yGKASogX2zG2Z-EEEPiJ0r23CEZH5_HeAV7ZWjw4ejhcgesMHHM_qw3RmhDw5y6eAJzawcpDZ1Oh5cL6M0u2NV7kfKq0ztHtXw4lzpR6CkGhBeUfiJ61o9eWFp4kSXhb3ZsAvWpwoQhd60QZdj-QAh0t9LpA577t4-rsEPGffDchViWUVrC1pVlyeMdGucGHz0E75WvCq29YIxRy886kGdoMmmqKroTW_LZR4hfB0LXL097Kt9uMB5sbYGFK7TRUT8_HpCFdnn3xRKVxyPehVQ_2wZ27-EhM4r5b7LQ40mlQcItDHyQh55CQ-fnwYVQK5irM1o1YHX4tcOrkqCF-reR5z5NY5RwSEb2-VkSA8DU6fJcXj7oNvtYalzYV_8_PWuCbuXZvtZJssjBfOHjzwLU65y4si0LdwD8Csu-uMUHvVDaxEOP14HHIqCdiO0h9hjFp7iKgo36WfK2M_oSAhuI_XGODFs7WIfbfDg0cWkl1GYeQ_MQnACQUPKUOGjWl8onqAL7Si8b3izVeYKZR8HJ5Ime0aFFgf3KUOgroUNNq1avPjmu9N4E5MFNebNRKCaiSQdywMhIIfyQnnr7xpqwZWfV-ofYOhToAL8PqIo-1Rg8ZxdGQ-zWL_OPISw4WE-nn1BdTV7VtTI1e50SGwifpyP5i_Vdnf5THEYgwu8ONTJf-nAKXERbPPmTc%3D&session_id=__session_id__"},
			LandingReward:      1,
			AdReportData:       nil,
			PreferredBrowser:   nil,
			Name:               "홈앤쇼핑",
			OrganizationID:     1,
			RevenueType:        "cpm",
			OwnerID:            6841,
			ClickTrackers:      []string{},
			ActionReward:       0,
			Events: ad.Events{
				{
					TrackingURLs: []string{"http://tracking.url.com"},
					Type:         "landed",
					Reward: &ad.Reward{
						Amount: 1,
						Status: "RECEIVABLE",
					},
				},
			},
		},
	},
}
