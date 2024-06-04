package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/recovery"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service/es"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"gopkg.in/olivere/elastic.v5"
)

const defaultContentSize = 30
const defaultPageLimit = 25

// V3ContentFetcher struct
type V3ContentFetcher struct {
	req          *dto.ContentArticlesRequest
	query        []elastic.Query
	orgQuery     []elastic.Query
	pageLimit    int
	searchSource *elastic.SearchSource
	scriptSort   *elastic.ScriptSort
}

func (f *V3ContentFetcher) buildReqWith(ctx context.Context, req *dto.ContentArticlesRequest) *V3ContentFetcher {
	f.req = req
	f.query = f.buildV3FilterQuery(ctx)
	f.pageLimit = f.buildPageLimit()
	f.searchSource, f.scriptSort = f.buildSortScriptAndSearchSource(ctx)
	return f
}

func (f *V3ContentFetcher) buildPageLimit() int {
	pageLimit := f.req.Size
	if pageLimit == 0 {
		pageLimit = defaultContentSize
	}
	return pageLimit
}

func (f *V3ContentFetcher) fetch(ctx context.Context) (*elastic.SearchResult, error) {
	defer recovery.LogRecoverWith(f.req)

	searchService := buzzscreen.Service.ES.Search().Index(env.Config.ElasticSearch.CampaignIndexName).FetchSource(true).
		SearchSource(f.searchSource).Type("content_campaign").TimeoutInMillis(1000).Preference("_local").Size(f.pageLimit)

	queryKey, queryErr := f.req.GetQueryKey()
	if queryErr != nil {
		return nil, queryErr
	}

	if queryKey != nil {
		searchService = searchService.From(queryKey.Index)
	}
	searchService = searchService.Query(elastic.NewBoolQuery().Filter(f.buildV3FilterQuery(ctx)...)).SortBy(f.scriptSort)

	if env.IsLocal() {
		searchService = searchService.Pretty(true)
	}
	searchResult, err := searchService.Do(context.Background())

	if err != nil {
		return nil, err
	}

	if searchResult.Hits == nil {
		return nil, errors.New("V3ContentFetcher buildReqWith() - es search failed")
	}

	return searchResult, err
}

func (f *V3ContentFetcher) getESCreativeType() string {
	if _, ok := (*f.req.GetTypes())["NATIVE"]; ok {
		return "R"
	} else if _, ok := (*f.req.GetTypes())["IMAGE"]; ok {
		return "A"
	} else {
		return ""
	}
}

func (f *V3ContentFetcher) buildV3FilterQuery(ctx context.Context) []elastic.Query {
	queryBuilder := es.NewESQueryBuilder().WithStartTime("now/h").WithEndTime("now/h").
		WithUnit(f.req.GetUnit(ctx).ID, f.req.GetUnit(ctx).ContentType == app.ContentTypeAll).
		WithAppID(f.req.Session.AppID).
		WithOrgID(f.req.GetUnit(ctx).OrganizationID).
		WithCountry(GetSupportedCountry(f.req.GetCountry(ctx))).
		WithCreativeTypes(f.getESCreativeType()).WithImageRatio(1.0).
		WithAge(f.req.GetAge()).WithGender(f.req.Gender).WithSdk(f.req.SdkVersion).
		WithRegisteredSeconds(f.req.Session.CreatedSeconds).
		WithWeekSlot(f.req.GetLocalTime(ctx)).
		WithCustomTargets(f.req.CustomTarget1, f.req.CustomTarget2, f.req.CustomTarget3).
		WithOsVersion(f.req.OsVersion).WithBatteryOptimization(f.req.IsInBatteryOpts)

	if profile := f.req.GetDynamoProfile(); profile != nil {
		if installedPackages := profile.InstalledPackages; installedPackages != nil {
			queryBuilder = queryBuilder.WithPackages(splitAndTrim(*installedPackages))
		}
	}

	queryKey, err := f.req.GetQueryKey()
	if err != nil {
		panic(err)
	}
	if queryKey != nil {
		lteTime := time.Unix(int64(queryKey.CreatedAt), 0)
		queryBuilder = queryBuilder.WithUpdatedTime(&lteTime)
	}

	if f.req.LandingTypes != "" {
		queryBuilder = queryBuilder.WithLandingTypes(splitAndTrim(f.req.LandingTypes)...)
	}

	// Category
	if f.req.Categories != "" {
		queryBuilder = queryBuilder.WithCategories(true, splitAndTrim(f.req.Categories)...)
	} else if f.req.CategoryID != "" {
		queryBuilder = queryBuilder.WithCategories(true, f.req.CategoryID)
	}

	if f.req.FilterCategories != "" {
		queryBuilder = queryBuilder.WithCategories(false, splitAndTrim(f.req.FilterCategories)...)
	}

	// ChannelID
	if f.req.ChannelID > 0 {
		queryBuilder = queryBuilder.WithChannels(true, strconv.FormatInt(f.req.ChannelID, 10))
	}

	if f.req.FilterChannelIDs != "" {
		queryBuilder = queryBuilder.WithChannels(false, splitAndTrim(f.req.FilterChannelIDs)...)
	}

	if f.req.GetUnit(ctx).FilteredProviders != nil {
		queryBuilder = queryBuilder.WithFilteredProviders(splitAndTrim(*f.req.GetUnit(ctx).FilteredProviders)...)
	}

	activity := f.req.GetDynamoActivity()
	if activity != nil {
		filterScript := es.GetScriptLoader().GetFilterScript()
		queryBuilder = queryBuilder.WithFrequencyCapping(filterScript, *activity)
	}

	queries := *queryBuilder.Build()

	statusQueries := make([]elastic.Query, 0)
	for _, status := range model.StatusesForLockscreen {
		statusQueries = append(statusQueries, elastic.NewTermQuery("status", status))
	}

	if f.req.GetUnit(ctx).UnitType == app.UnitTypeNative {
		for _, status := range model.StatusesForFeed {
			statusQueries = append(statusQueries, elastic.NewBoolQuery().Must(elastic.NewTermQuery("status", status), elastic.NewExistsQuery("related")))
		}
	}

	if len(statusQueries) > 0 {
		queries = append(queries, elastic.NewBoolQuery().Should(statusQueries...))
	}

	return queries
}

func (f *V3ContentFetcher) buildSortScriptAndSearchSource(ctx context.Context) (*elastic.SearchSource, *elastic.ScriptSort) {
	modelArtifact := *f.req.GetModelArtifact(ctx)

	// 1. Build score script for ranking
	rawScript := es.GetScriptLoader().GetScoreScript(modelArtifact, es.JoinOpAdd)
	script := elastic.NewScript(rawScript)
	params := make(map[string]interface{})

	// Add preferred categories information for personalization
	if cs := f.req.GetCategoriesScores(); cs != nil {
		params["categoryProfile"] = *cs
	}

	// Add entity profiles for personalization
	if es := f.req.GetEntityScores(); es != nil {
		params["entityProfile"] = *es
	}

	// Add seen content ids
	if activity := f.req.GetDynamoActivity(); activity != nil {
		if len(activity.SeenCampaignIDs) != 0 {
			params["seenIDs"] = activity.SeenCampaignIDs
		}
	}
	script.Params(params)

	// 2. Set script to searchSource
	artifactScript := elastic.NewScript(fmt.Sprintf("'%s'", modelArtifact))
	searchSource := elastic.NewSearchSource().FetchSource(true).ScriptField(elastic.NewScriptField("modelArtifact", artifactScript))

	// 3. Set searchSource for debug scoring if needed
	// Check if device is required a detailed scoring analysis
	if f.req.GetIsDebugScore() {
		searchSource.ScriptFields(es.GetScriptLoader().GetScriptFields(modelArtifact, params)...)
	}

	scriptSort := elastic.NewScriptSort(script, "number").Desc()

	return searchSource, scriptSort
}

// V1ContentFetcher struct definition
type V1ContentFetcher struct {
	req          *dto.ContentAllocV1Request
	query        []elastic.Query
	orgQuery     []elastic.Query
	pageLimit    int
	searchSource *elastic.SearchSource
	scriptSort   *elastic.ScriptSort
}

func (f *V1ContentFetcher) buildReqWith(ctx context.Context, req *dto.ContentAllocV1Request) *V1ContentFetcher {
	f.req = req
	f.query = f.buildV1FilterQuery(ctx, false)
	f.pageLimit = f.buildPageLimit(ctx)
	f.searchSource, f.scriptSort = f.buildSortScriptAndSearchSource(ctx)
	return f
}

func (f *V1ContentFetcher) buildPageLimit(ctx context.Context) int {
	pageLimit := f.req.GetUnit(ctx).PageLimit
	if f.req.GetUnit(ctx).PageLimit < defaultPageLimit {
		pageLimit = defaultPageLimit
	}
	return pageLimit
}

func (f *V1ContentFetcher) fetch() (*elastic.SearchResult, error) {
	defer recovery.LogRecoverWith(f.req)
	ss := buzzscreen.Service.ES.Search().Index(env.Config.ElasticSearch.CampaignIndexName).FetchSource(true).
		SearchSource(f.searchSource).Type("content_campaign").Query(elastic.NewBoolQuery().Filter(f.query...)).
		SortBy(f.scriptSort).TimeoutInMillis(1000).Preference("_local").Size(f.pageLimit)

	if env.IsLocal() {
		ss.Pretty(true)
	}

	return ss.Do(context.Background())
}

func (f *V1ContentFetcher) buildV1FilterQuery(ctx context.Context, addOrgFilter bool) []elastic.Query {
	allocReq := f.req
	targetCarrier := allocReq.Carrier
	if carrierInMap, ok := model.CarrierMap[allocReq.Carrier]; ok {
		targetCarrier = string(carrierInMap)
	}

	queryBuilder := es.NewESQueryBuilder().
		WithStartTime("now/h").WithEndTime("now+1h/h").
		WithCountry(GetSupportedCountry(allocReq.GetCountry(ctx))).
		WithGender(allocReq.Gender).WithCarrier(targetCarrier).
		WithUnit(allocReq.GetUnit(ctx).ID, allocReq.GetUnit(ctx).ContentType != app.ContentTypeUnitOnly). //동아닷컴 (255009465857996, 363944316301025)
		WithAppID(allocReq.GetAppID(ctx)).
		WithOrgID(allocReq.GetOrganizationID(ctx)).
		WithAge(allocReq.GetTargetAge()).WithSdk(allocReq.SdkVersion).
		WithRegisteredSeconds(allocReq.GetRegisteredSeconds()).
		WithWeekSlot(allocReq.GetLocalTime(ctx)).WithCreativeTypes(allocReq.GetCreativeType(ctx)).
		WithCustomTargets(allocReq.CustomTarget1, allocReq.CustomTarget2, allocReq.CustomTarget3).
		WithOsVersion(allocReq.DeviceOs).WithBatteryOptimization(allocReq.IsInBatteryOpts).
		WithStatus(model.StatusesForLockscreen...).WithRegion(allocReq.Region)

	if profile := allocReq.GetDynamoProfile(); profile != nil {
		if installedPackages := profile.InstalledPackages; installedPackages != nil {
			queryBuilder = queryBuilder.WithPackages(splitAndTrim(*installedPackages))
		}
	}

	if allocReq.Language != "" {
		queryBuilder = queryBuilder.WithLanguage(splitAndTrim(allocReq.Language)...)
	}

	// Category
	if f.req.Categories != "" {
		queryBuilder = queryBuilder.WithCategories(true, splitAndTrim(f.req.Categories)...)
	}

	if allocReq.FilterCategories != "" {
		queryBuilder = queryBuilder.WithCategories(false, splitAndTrim(allocReq.FilterCategories)...)
	}

	if allocReq.FilterChannelIDs != "" {
		queryBuilder = queryBuilder.WithChannels(false, splitAndTrim(allocReq.FilterChannelIDs)...)
	}

	if allocReq.GetUnit(ctx).FilteredProviders != nil {
		queryBuilder = queryBuilder.WithFilteredProviders(splitAndTrim(*allocReq.GetUnit(ctx).FilteredProviders)...)
	}

	if addOrgFilter {
		queryBuilder = queryBuilder.WithOrganization(allocReq.GetOrganizationID(ctx))
	}

	activity := allocReq.GetDynamoActivity()
	if activity != nil {
		filterScript := es.GetScriptLoader().GetFilterScript()
		queryBuilder = queryBuilder.WithFrequencyCapping(filterScript, *activity)
	}

	return *queryBuilder.Build()
}

func (f *V1ContentFetcher) buildSortScriptAndSearchSource(ctx context.Context) (*elastic.SearchSource, *elastic.ScriptSort) {
	modelArtifact := *f.req.GetModelArtifact(ctx)

	// 1. Build score script for ranking
	joinOp := es.JoinOpMultiply
	if modelArtifact != "v1" {
		joinOp = es.JoinOpAdd
	}
	rawScript := es.GetScriptLoader().GetScoreScript(modelArtifact, joinOp)
	script := elastic.NewScript(rawScript)
	params := make(map[string]interface{})

	// Add preferred categories information for personalization
	if cs := f.req.GetCategoriesScores(); cs != nil {
		params["categoryProfile"] = *cs
	}

	// Add entity profiles for personalization
	if es := f.req.GetEntityScores(); es != nil {
		params["entityProfile"] = *es
	}

	// Add seen content ids
	if activity := f.req.GetDynamoActivity(); activity != nil {
		if len(activity.SeenCampaignIDs) != 0 {
			params["seenIDs"] = activity.SeenCampaignIDs
		}
	}
	script.Params(params)

	// 2. Set searchSource for debug scoring if needed
	artifactScript := elastic.NewScript(fmt.Sprintf("'%s'", modelArtifact))
	searchSource := elastic.NewSearchSource().FetchSource(true).ScriptField(elastic.NewScriptField("modelArtifact", artifactScript))
	// Check if device is required a detailed scoring analysis
	if f.req.GetIsDebugScore() {
		searchSource.ScriptFields(es.GetScriptLoader().GetScriptFields(modelArtifact, params)...)
	}

	scriptSort := elastic.NewScriptSort(script, "number").Desc()

	return searchSource, scriptSort
}
