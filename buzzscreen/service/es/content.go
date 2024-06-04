package es

import (
	"strconv"

	"strings"

	"time"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"gopkg.in/olivere/elastic.v5"
)

const (
	esTimeFormat = "2006-01-02T15:04:05-07:00"
)

type (
	// QueryBuilder type definition
	QueryBuilder struct {
		queries []elastic.Query
	}
)

// NewESQueryBuilder func definition
func NewESQueryBuilder() *QueryBuilder {
	queryBuilder := QueryBuilder{
		queries: make([]elastic.Query, 0),
	}
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewTermQuery("is_enabled", true))
	return &queryBuilder
}

// Build func definition
func (queryBuilder *QueryBuilder) Build() *[]elastic.Query {
	return &queryBuilder.queries
}

// WithStartTime func definition
func (queryBuilder *QueryBuilder) WithStartTime(lte string) *QueryBuilder {
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewRangeQuery("start_date").Lte(lte))
	return queryBuilder
}

// WithEndTime func definition
func (queryBuilder *QueryBuilder) WithEndTime(gt string) *QueryBuilder {
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewRangeQuery("end_date").Gt(gt))
	return queryBuilder
}

// WithCountry func definition
func (queryBuilder *QueryBuilder) WithCountry(cou string) *QueryBuilder {
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(
		elastic.NewTermQuery("country", model.ESGlobString),
		elastic.NewTermQuery("country", cou),
	))
	return queryBuilder
}

// WithGender func definition
func (queryBuilder *QueryBuilder) WithGender(gender string) *QueryBuilder {
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(
		elastic.NewBoolQuery().Should(
			elastic.NewTermQuery("target_gender", model.ESGlobString),
			elastic.NewTermQuery("target_gender", gender),
		),
	))
	return queryBuilder
}

// WithCarrier func definition
func (queryBuilder *QueryBuilder) WithCarrier(carrier string) *QueryBuilder {
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(
		elastic.NewTermQuery("target_carrier", model.ESGlobString),
		elastic.NewTermQuery("target_carrier", carrier),
	))
	return queryBuilder
}

// WithOrganization func definition
func (queryBuilder *QueryBuilder) WithOrganization(orgID int64) *QueryBuilder {
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Must(elastic.NewTermQuery("organization_id", orgID)))
	return queryBuilder
}

// WithUnit func definition
func (queryBuilder *QueryBuilder) WithUnit(unitID int64, withGlob bool) *QueryBuilder {
	unitQuery := elastic.NewBoolQuery().
		Should(elastic.NewTermQuery("target_unit", strconv.FormatInt(unitID, 10))).
		MustNot(elastic.NewTermQuery("detarget_unit", strconv.FormatInt(unitID, 10)))
	if withGlob {
		unitQuery.Should(elastic.NewTermQuery("target_unit", model.ESGlobString))
	}
	queryBuilder.queries = append(queryBuilder.queries, unitQuery)
	return queryBuilder
}

// WithAppID func definition
func (queryBuilder *QueryBuilder) WithAppID(appID int64) *QueryBuilder {
	appIDQuery := elastic.NewBoolQuery().
		Should(elastic.NewTermQuery("target_app_id", strconv.FormatInt(appID, 10))).
		Should(elastic.NewTermQuery("target_app_id", model.ESGlobString)).
		MustNot(elastic.NewTermQuery("detarget_app_id", strconv.FormatInt(appID, 10)))

	queryBuilder.queries = append(queryBuilder.queries, appIDQuery)
	return queryBuilder
}

// WithOrgID func definition
func (queryBuilder *QueryBuilder) WithOrgID(orgID int64) *QueryBuilder {
	orgQuery := elastic.NewBoolQuery().
		Should(elastic.NewTermQuery("target_org", strconv.FormatInt(orgID, 10))).
		Should(elastic.NewTermQuery("target_org", model.ESGlobString)).
		MustNot(elastic.NewTermQuery("detarget_org", strconv.FormatInt(orgID, 10)))

	queryBuilder.queries = append(queryBuilder.queries, orgQuery)
	return queryBuilder
}

// WithPackages func definition
func (queryBuilder *QueryBuilder) WithPackages(packages []string) *QueryBuilder {
	unitQuery := elastic.NewBoolQuery().
		Should(elastic.NewTermsQuery("target_app", append(stringsToInterfaces(packages), model.ESGlobString)...)).
		MustNot(elastic.NewTermsQuery("detarget_app", stringsToInterfaces(packages)...))
	queryBuilder.queries = append(queryBuilder.queries, unitQuery)
	return queryBuilder
}

// WithStatus func definition
func (queryBuilder *QueryBuilder) WithStatus(statuses ...model.Status) *QueryBuilder {
	statusQueries := make([]elastic.Query, 0)
	for _, status := range statuses {
		statusQueries = append(statusQueries, elastic.NewTermQuery("status", status))
	}
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(statusQueries...))
	return queryBuilder
}

// WithAge func definition
func (queryBuilder *QueryBuilder) WithAge(age int) *QueryBuilder {
	if age == 0 {
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewTermQuery("target_age_min", model.ESNullShortMin), elastic.NewTermQuery("target_age_max", model.ESNullShortMax))
	} else {
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewRangeQuery("target_age_min").Lte(age), elastic.NewRangeQuery("target_age_max").Gte(age))
	}
	return queryBuilder
}

// WithSdk func definition
func (queryBuilder *QueryBuilder) WithSdk(sdkVersion int) *QueryBuilder {
	if sdkVersion == 0 {
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewTermQuery("target_sdk_min", model.ESNullIntMin), elastic.NewTermQuery("target_sdk_max", model.ESNullIntMax))
	} else {
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewRangeQuery("target_sdk_min").Lte(sdkVersion), elastic.NewRangeQuery("target_sdk_max").Gte(sdkVersion))
	}
	return queryBuilder
}

// WithMostRecent func definition
func (queryBuilder *QueryBuilder) WithMostRecent() *QueryBuilder {
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewRangeQuery("published_date").Lte("now"), elastic.NewRangeQuery("published_date").Gte("now-1d/d"))
	return queryBuilder
}

// WithLanguage func definition
func (queryBuilder *QueryBuilder) WithLanguage(languages ...string) *QueryBuilder {
	if len(languages) > 0 {
		languageQueries := make([]elastic.Query, 0)
		for _, lang := range languages {
			languageQueries = append(languageQueries, elastic.NewTermQuery("target_language", lang))
		}
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(languageQueries...))
	}
	return queryBuilder
}

func (queryBuilder *QueryBuilder) withKeyAndFilteredItems(key string, positive bool, items ...string) *QueryBuilder {
	if len(items) > 0 {
		var query elastic.Query
		positiveQueries := make([]elastic.Query, 0)
		for _, item := range items {
			if positive {
				positiveQueries = append(positiveQueries, elastic.NewTermQuery(key, item))
			}
		}
		if positive {
			query = elastic.NewBoolQuery().Should(positiveQueries...)
		} else {
			query = elastic.NewBoolQuery().MustNot(elastic.NewTermsQuery(key, stringsToInterfaces(items)...))
		}

		queryBuilder.queries = append(queryBuilder.queries, query)
	}
	return queryBuilder
}

// WithoutCampaignIDs func definition
func (queryBuilder *QueryBuilder) WithoutCampaignIDs(positive bool, campaignIDs ...string) *QueryBuilder {
	return queryBuilder.withKeyAndFilteredItems("id", positive, campaignIDs...)
}

// WithCategories func definition
func (queryBuilder *QueryBuilder) WithCategories(positive bool, categories ...string) *QueryBuilder {
	return queryBuilder.withKeyAndFilteredItems("categories", positive, categories...)
}

// WithChannels func definition
func (queryBuilder *QueryBuilder) WithChannels(positive bool, channelIDs ...string) *QueryBuilder {
	return queryBuilder.withKeyAndFilteredItems("channel_id", positive, channelIDs...)
}

// WithRegisteredSeconds func definition
func (queryBuilder *QueryBuilder) WithRegisteredSeconds(registeredSeconds int64) *QueryBuilder {
	if registeredSeconds > 0 {
		registeredDays := utils.GetDaysFrom(registeredSeconds) + 1
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewRangeQuery("registered_days_min").Lte(registeredDays), elastic.NewRangeQuery("registered_days_max").Gte(registeredDays))
	} else {
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewTermQuery("registered_days_min", model.ESNullShortMin), elastic.NewTermQuery("registered_days_max", model.ESNullShortMax))
	}
	return queryBuilder
}

// WithWeekSlot func definition
func (queryBuilder *QueryBuilder) WithWeekSlot(localTime *time.Time) *QueryBuilder {
	weekSlotQueries := elastic.NewBoolQuery().Should(elastic.NewTermQuery("week_slot", model.ESGlobString))
	if localTime != nil {
		weekSlotQueries.Should(elastic.NewTermQuery("week_slot", strconv.Itoa(utils.GetWeekdayStartsFromMonday(localTime)*24+localTime.Hour())))
	}
	queryBuilder.queries = append(queryBuilder.queries, weekSlotQueries)
	return queryBuilder
}

// WithImageRatio func definition
// TODO: Remove BoolMustNotExistsConstraint after 2018-08-01
func (queryBuilder *QueryBuilder) WithImageRatio(lowerBound float32) *QueryBuilder {
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("image_ratio")), elastic.NewRangeQuery("image_ratio").Gte(lowerBound)))
	return queryBuilder
}

// WithCreativeTypes func definition
func (queryBuilder *QueryBuilder) WithCreativeTypes(creativeTypes string) *QueryBuilder {
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(elastic.NewTermQuery("creative_types", creativeTypes)))
	return queryBuilder
}

// WithRegion func definition
func (queryBuilder *QueryBuilder) WithRegion(region string) *QueryBuilder {
	regionState := ""
	if region != "" {
		regionState = strings.Fields(region)[0]
	}

	queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(
		elastic.NewTermQuery("target_region", model.ESGlobString),
		elastic.NewTermQuery("target_region", regionState),
		elastic.NewTermQuery("target_region", region),
	))
	return queryBuilder
}

func stringsToInterfaces(strs []string) []interface{} {
	inters := make([]interface{}, 0)
	for _, str := range strs {
		inters = append(inters, str)
	}
	return inters
}

// WithFilteredProviders func definition
func (queryBuilder *QueryBuilder) WithFilteredProviders(providers ...string) *QueryBuilder {
	if len(providers) > 0 {
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().MustNot(
			elastic.NewTermsQuery("provider_id", stringsToInterfaces(providers)...),
		))
	}
	return queryBuilder
}

// WithCustomTargets func definition
func (queryBuilder *QueryBuilder) WithCustomTargets(target1, target2, target3 string) *QueryBuilder {
	// Custom target 1,2,3
	customTargetMap := map[string]string{
		"custom_target_1": target1,
		"custom_target_2": target2,
		"custom_target_3": target3,
	}
	for key, targets := range customTargetMap {
		switch len(targets) {
		case 0:
			queryBuilder.queries = append(queryBuilder.queries, elastic.NewTermQuery(key, model.ESGlobString))
		default:
			customTargetQueries := []elastic.Query{elastic.NewTermQuery(key, model.ESGlobString)}
			for _, target := range strings.Split(targets, ",") {
				if target != "" {
					customTargetQueries = append(customTargetQueries, elastic.NewTermQuery(key, target))
				}
			}
			queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(customTargetQueries...))
		}
	}
	return queryBuilder
}

// WithUpdatedTime func definition
func (queryBuilder *QueryBuilder) WithUpdatedTime(lteTime *time.Time) *QueryBuilder {
	if lteTime != nil {
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewRangeQuery("updated_at").Lte(lteTime.Format(esTimeFormat)))
	}
	return queryBuilder
}

// WithLandingTypes func definition
func (queryBuilder *QueryBuilder) WithLandingTypes(landingTypes ...string) *QueryBuilder {
	landingTypeQueries := make([]elastic.Query, 0)
	for _, landingType := range landingTypes {
		if landingType != "" {
			landingTypeQueries = append(landingTypeQueries, elastic.NewTermQuery("landing_type", landingType))
		}
	}
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewBoolQuery().Should(landingTypeQueries...))
	return queryBuilder
}

// WithOsVersion func definition
func (queryBuilder *QueryBuilder) WithOsVersion(osVersion int) *QueryBuilder {
	if osVersion != 0 {
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewRangeQuery("target_os_min").Lte(osVersion), elastic.NewRangeQuery("target_os_max").Gte(osVersion))
	}
	return queryBuilder
}

// WithBatteryOptimization func definition
func (queryBuilder *QueryBuilder) WithBatteryOptimization(isInBatteryOpts bool) *QueryBuilder {
	if !isInBatteryOpts {
		queryBuilder.queries = append(queryBuilder.queries, elastic.NewTermQuery("target_battery_optimization", false))
	}
	return queryBuilder
}

// WithFrequencyCapping func definition
func (queryBuilder *QueryBuilder) WithFrequencyCapping(filterScript string, activity device.Activity) *QueryBuilder {
	params := map[string]interface{}{
		"seenCampaignCountForDay":  activity.SeenCampaignCountForDay,
		"seenCampaignCountForHour": activity.SeenCampaignCountForHour,
	}

	script := elastic.NewScript(filterScript)
	script.Params(params)
	queryBuilder.queries = append(queryBuilder.queries, elastic.NewScriptQuery(script))

	return queryBuilder
}
