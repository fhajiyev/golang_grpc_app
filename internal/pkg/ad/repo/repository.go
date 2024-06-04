package repo

import (
	"fmt"
	"net/http"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	"github.com/google/go-querystring/query"
)

// Repository struct
type Repository struct {
	redisCache rediscache.RedisSource
	m          *mapper
	buzzAdURL  string
}

const (
	pathAdsV1 = "/api/lockscreen/ads"
	pathAdsV2 = "/api/v2/lockscreen/ads"
)

// GetAdDetail returns ad.Detail from BuzzAd with 1 hour caching
func (r *Repository) GetAdDetail(adID int64, accessToken string) (*ad.Detail, error) {
	key := r.getCacheKey(adID)
	detail, err := r.getAdDetailFromCache(key)
	if err != nil {
		detail, err = r.getAdDetailFromBuzzAd(adID, accessToken)
		if err != nil {
			return nil, err
		}
		r.setAdDetailToCacheAsync(key, *detail)
	}

	return detail, nil
}

// GetAdsV1 definition
func (r *Repository) GetAdsV1(v1Req ad.V1AdsRequest) (*ad.V1AdsResponse, error) {
	values, err := query.Values(v1Req)
	if err != nil {
		return nil, err
	}

	req := &network.Request{
		Method: http.MethodGet,
		Params: &values,
		URL:    r.buzzAdURL + pathAdsV1,
	}

	var v1AdsResponse ad.V1AdsResponse
	statusCode, err := req.GetResponse(&v1AdsResponse)
	if err != nil {
		core.Logger.WithError(err).WithField("user", v1Req.UnitDeviceToken).Warnf("BuzzAd Request Error, req: %v, status: %v", v1Req, statusCode)
		return nil, err
	} else if statusCode/100 != 2 {
		core.Logger.WithError(err).WithField("user", v1Req.UnitDeviceToken).Warnf("BuzzAd code error, req: %v, status: %v", v1Req, statusCode)
		return nil, err
	} else if len(v1AdsResponse.Ads) == 0 {
		core.Logger.Warnf("GetV1AdFromBuzzAd() - Ads is empty. %s, %v", r.buzzAdURL, v1Req)
	}

	core.Logger.Debugf("GetV1AdFromBuzzAd() - req: %v, Ads: %v, NativeAds: %v, Settings: %#v", v1Req, len(v1AdsResponse.Ads), len(v1AdsResponse.NativeAds), v1AdsResponse.Settings)
	return &v1AdsResponse, nil
}

// GetAdsV2 definition
func (r *Repository) GetAdsV2(v2Req ad.V2AdsRequest) (*ad.V2AdsResponse, error) {
	values, err := query.Values(v2Req)
	if err != nil {
		return nil, err
	}

	req := &network.Request{
		Method: http.MethodGet,
		Params: &values,
		URL:    r.buzzAdURL + pathAdsV2,
	}

	var v2AdsResponse ad.V2AdsResponse
	statusCode, err := req.GetResponse(&v2AdsResponse)
	if statusCode != 200 || err != nil {
		if err == nil {
			err = fmt.Errorf("GetV2AdFromBuzzAd() - status code is %v", statusCode)
		}
		return nil, err
	}

	core.Logger.Debugf("GetAdsFromBuzzAd() - url %v, params: %+v", r.buzzAdURL, values)
	return &v2AdsResponse, nil
}

func (r *Repository) getAdDetailFromBuzzAd(adID int64, accessToken string) (*ad.Detail, error) {
	req := &network.Request{
		URL:    fmt.Sprintf("%s/adserver/orders/lineitems/%v", r.buzzAdURL, adID),
		Method: http.MethodGet,
		Header: &http.Header{"Authorization": {fmt.Sprintf("Token %s", accessToken)}},
	}

	detail := adDetail{}
	status, err := req.GetResponse(&detail)
	if err != nil || status != http.StatusOK {
		return nil, fmt.Errorf("failed to get lineitem from BA. err: %s, response: %+v", err, detail)
	}

	return r.m.toEntityDetail(detail), nil
}

// New returns Repository struct
func New(redisCache rediscache.RedisSource, buzzAdURL string) *Repository {
	return &Repository{redisCache, &mapper{}, buzzAdURL}
}
