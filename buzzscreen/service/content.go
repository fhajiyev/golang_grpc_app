package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"gopkg.in/olivere/elastic.v5"
)

func getCacheKeyCategories(lang string) string {
	return fmt.Sprintf("CACHE_GO_GET_CATEGORIES_%v", lang)
}

// GetCategories func definition
func GetCategories(lang string) (*model.ContentCategories, error) {
	var categories model.ContentCategories

	cacheKey := getCacheKeyCategories(lang)
	err := env.GetCache(cacheKey, &categories)
	if len(categories) == 0 || err != nil {
		bcCategories := make([]map[string]string, 0)

		_, err = (&network.Request{
			URL: env.Config.BuzzconInternalURL + "/content/categories",
			Params: &url.Values{
				"language": {lang},
			},
		}).GetResponse(&bcCategories)

		if err == nil && len(bcCategories) > 0 {
			for _, bcCat := range bcCategories {
				cat := model.ContentCategory{
					ID:          bcCat["id"],
					Name:        bcCat["name"],
					Translation: bcCat["name"],
				}
				if val, ok := bcCat["icon_url"]; ok {
					cat.IconURL = &val
				}
				categories = append(categories, &cat)
			}

			env.SetCache(cacheKey, categories, time.Hour*24)
		}
	}

	return &categories, err
}

// GetCategoriesMap func definition
func GetCategoriesMap(lang string) (map[string]*model.ContentCategory, error) {
	categoriesMap := make(map[string]*model.ContentCategory)
	categories, err := GetCategories(lang)
	if err == nil {
		for _, category := range *categories {
			categoriesMap[category.ID] = category
		}
	}

	return categoriesMap, err
}

func getCacheKeyChannel(channelID int64) string {
	return fmt.Sprintf("CACHE_GO_CHANNEL_%v", channelID)
}

// GetChannel func definition
func GetChannel(channelID int64) *model.ContentChannel {
	channel := model.ContentChannel{
		ID: channelID,
	}
	cacheKey := getCacheKeyChannel(channelID)
	if err := env.GetCache(cacheKey, &channel); channel.ID == 0 || err != nil {
		buzzscreen.Service.DB.Where(&channel).First(&channel)
		if channel.ID != 0 {
			env.SetCache(cacheKey, channel, time.Hour*12)
		}
	}
	return &channel
}

// GetContentCampaignsFromDB func definition
func GetContentCampaignsFromDB(strIDs []string) (camps []*model.ContentCampaign) {
	articleIDs, err := utils.SliceAtoi(strIDs)
	if err != nil {
		panic(err)
	} else {
		buzzscreen.Service.DB.Where("id IN (?)", articleIDs).Order(fmt.Sprintf("FIELD(id, %s)", strings.Join(strIDs, ","))).Find(&camps)
	}
	return
}

// GetChannelMap func definition
func GetChannelMap(camps []*model.ContentCampaign) map[int64]*model.ContentChannel {
	channelMap := make(map[int64]*model.ContentChannel)
	for _, camp := range camps {
		cID := *camp.ChannelID
		if _, ok := channelMap[cID]; !ok {
			channelMap[cID] = nil
		}
	}
	cc := make(chan *model.ContentChannel, len(channelMap))
	for channelID := range channelMap {
		go func(channelID int64) {
			cc <- GetChannel(channelID)
		}(channelID)
	}
	chLen := len(channelMap)
	if chLen > 0 {
		for i := 0; i < chLen; i++ {
			channel := <-cc
			channelMap[channel.ID] = channel
		}
	}

	return channelMap
}

// GetContentCampaignByIDs func definition
func GetContentCampaignByIDs(deviceID int64, campIDs ...int64) ([]*dto.ESContentCampaign, error) {
	ss := buzzscreen.Service.ES.Search().Index(env.Config.ElasticSearch.CampaignIndexName).Type("content_campaign").TimeoutInMillis(1000).Preference("_local")
	idInterfaces := make([]interface{}, 0)
	for _, campID := range campIDs {
		idInterfaces = append(idInterfaces, campID)
	}
	ss = ss.Query(elastic.NewBoolQuery().Filter(elastic.NewBoolQuery().Must(elastic.NewTermsQuery("id", idInterfaces...))))

	if env.IsLocal() {
		ss.Pretty(true)
	}
	sr, err := ss.Do(context.Background())
	if err != nil {
		return nil, err
	}
	contentCampaigns := make([]*dto.ESContentCampaign, 0)
	if err != nil || sr.Hits == nil {
		errLog := core.Logger.WithError(err)
		if sr != nil && sr.Error != nil {
			errLog.Errorf("getContentCampaignsFromES() - error: %v", sr.Error.CausedBy)
		} else {
			errLog.Errorf("getContentCampaignsFromES() - searchResult: %v", sr)
		}
	} else {
		contentCampaigns = *parseESHitsToContentCampaigns(deviceID, sr)
	}

	return contentCampaigns, nil
}

func logDebugScore(did int64, ts string, ccs []*dto.ESContentCampaign) {
	for _, cc := range ccs {
		debugScoreForLog := DebugScoreForLog{
			LogType:           "DebugScore",
			DeviceID:          did,
			ESContentCampaign: *cc,
			TimeAllocated:     ts,
		}
		if mapForLog, err := debugScoreForLog.BuildMap(); err == nil {
			core.Loggers["general"].WithFields(mapForLog).Info("Log")
		}
	}
}

// GetV1ContentCampaignsFromES is Content allocation logic for V1
func GetV1ContentCampaignsFromES(ctx context.Context, allocReq *dto.ContentAllocV1Request) ([]*dto.ESContentCampaign, error) {
	if allocReq.GetUnit(ctx) == nil {
		return nil, errors.New("can't find the unit")
	}

	fetcher := (&V1ContentFetcher{}).buildReqWith(ctx, allocReq)
	searchResult, err := fetcher.fetch()

	var contentCampaigns []*dto.ESContentCampaign

	if err == nil && searchResult.Hits != nil {
		contentCampaigns = *parseESHitsToContentCampaigns(allocReq.DeviceID, searchResult)
	}

	// If device is required for intermediate logging
	if fetcher.req.GetIsDebugScore() {
		logDebugScore(fetcher.req.DeviceID, time.Now().UTC().String(), contentCampaigns)
	}

	return contentCampaigns, err
}

func parseESSourceToContentCampaign(source []byte) (*dto.ESContentCampaign, error) {
	var esContent dto.ESContentCampaign
	err := json.Unmarshal(source, &esContent)
	if err == nil {
		esContent, err = esContent.EscapeNull()
	}
	return &esContent, err
}

func parseESHitsToContentCampaigns(deviceID int64, searchResult *elastic.SearchResult) *[]*dto.ESContentCampaign {
	contentCampaigns := make([]*dto.ESContentCampaign, 0)
	ccIDs := make([]int64, 0)
	if searchResult.Hits.TotalHits > 0 {
		for _, hit := range searchResult.Hits.Hits {
			if esContent, err := parseESSourceToContentCampaign(*hit.Source); err != nil {
				core.Logger.WithError(err).Errorf("parseESHitsToContentCampaigns() - json parse error. _id: %v", hit.Id)
			} else {
				// Add score factors & model artifact
				var esScoreFactor = make(map[string]float64)

				for k, v := range hit.Fields {
					varray := reflect.ValueOf(v)
					if k == "modelArtifact" {
						modelArtifact := varray.Index(0).Interface().(string)
						esContent.ModelArtifact = modelArtifact
					} else {
						weight := varray.Index(0).Interface().(float64)
						esScoreFactor[k] = weight
					}
				}
				esContent.ScoreFactors = esScoreFactor

				// Add final score
				if len(hit.Sort) > 0 {
					score := (hit.Sort)[0].(float64)
					esContent.Score = &score
				}
				ccIDs = append(ccIDs, esContent.ID)
				contentCampaigns = append(contentCampaigns, esContent)

			}
		}
	}
	return &contentCampaigns
}

// GetContentCampaignsFromES is Content allocation logic for V3
func GetContentCampaignsFromES(ctx context.Context, contentReq *dto.ContentArticlesRequest) ([]*dto.ESContentCampaign, int, error) {
	fetcher := (&V3ContentFetcher{}).buildReqWith(ctx, contentReq)
	searchResult, err := fetcher.fetch(ctx)

	if err != nil {
		return nil, 0, err
	}

	if searchResult.Hits == nil {
		return nil, 0, errors.New("GetContentCampaignsFromES() - es search failed")
	}

	var contentCampaigns []*dto.ESContentCampaign

	if err == nil && searchResult.Hits != nil {
		contentCampaigns = *parseESHitsToContentCampaigns(contentReq.Session.DeviceID, searchResult)
	}

	// If device is required for intermediate logging
	if contentReq.GetIsDebugScore() {
		logDebugScore(contentReq.Session.DeviceID, time.Now().UTC().String(), contentCampaigns)
	}

	return contentCampaigns, int(searchResult.Hits.TotalHits), nil
}

func splitAndTrim(commaSeparatedString string) []string {
	splittedStrings := strings.Split(commaSeparatedString, ",")
	for i := range splittedStrings {
		splittedStrings[i] = strings.TrimSpace(splittedStrings[i])
	}
	return splittedStrings
}

// GetContentChannels func definition
func GetContentChannels(contentChannelQuery *model.ContentChannelsQuery) *model.ContentChannels {
	channels := contentChannelQuery.Get()
	return channels
}

// The list of supported countries are from https://github.com/Buzzvil/dash/blob/master/src/lib/constants.js
var countriesMap = map[string]struct{}{
	"AT": {},
	"AR": {},
	"AU": {},
	"BD": {},
	"BE": {},
	"BR": {},
	"CA": {},
	"CH": {},
	"CO": {},
	"CN": {},
	"DE": {},
	"DK": {},
	"ES": {},
	"FI": {},
	"FR": {},
	"GB": {},
	"HK": {},
	"ID": {},
	"IN": {},
	"IT": {},
	"JP": {},
	"KR": {},
	"LA": {},
	"MX": {},
	"MY": {},
	"NO": {},
	"NZ": {},
	"OM": {},
	"PH": {},
	"PK": {},
	"PL": {},
	"RU": {},
	"SE": {},
	"SG": {},
	"TH": {},
	"TW": {},
	"US": {},
	"VE": {},
	"VN": {},
	"ZA": {},
	"ZZ": {},
}

// DistCountry struct definition
type DistCountry struct {
	Country string
}

// GetSupportedCountry func definition
func GetSupportedCountry(country string) string {
	if _, ok := countriesMap[country]; ok {
		return country
	}
	return "US"
}
