package tests

import (
	"context"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	appRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/app/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	deviceRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/device/repo"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbapp"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"

	"strconv"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/jinzhu/gorm"
)

//noinspection GoUnusedConst
const (
	testMapping = `
{
	"settings":{
		"analysis": {
			"analyzer": {
				"comma_analyzer": {
					"type": "custom",
					"tokenizer": "comma_tokenizer"
				},
				"ngram_analyzer": {
					"type": "custom",
					"tokenizer": "ngram_tokenizer"
				},
				"i_ngram_analyzer": {
					"type": "custom",
					"tokenizer": "ngram_tokenizer",
					"filter": [
						"lowercase"
					]
				}
			},
			"tokenizer": {
				"comma_tokenizer": {
					"type": "pattern",
					"pattern": ","
				},
				"ngram_tokenizer": {
					"type": "nGram",
					"min_gram": "1",
					"max_gram": "50"
				}
			}
		},
		"index": {
			"number_of_shards": 1,
			"number_of_replicas": 1,
			"auto_expand_replicas": "0-all",
			"refresh_interval": "1s"
		}
	},

	"mappings": {
		"content_campaign": {
			"_all": {
				"enabled": "false"
			},
			"properties": {
				"id": { "type": "integer" },
				"name": { "type": "text", "index": false },
				"title": { "type": "text", "analyzer": "i_ngram_analyzer" },
				"description": { "type": "text", "index": false },
				"categories": { "type": "text", "analyzer": "comma_analyzer" },
				"tags": { "type": "text", "analyzer": "comma_analyzer" },
				"image": { "type": "text", "index": false },
				"click_url": { "type": "text", "index": false },
				"media_type": { "type": "short" },
				"landing_type": { "type": "short" },
				"clean_mode": { "type": "short" },
				"clean_link": { "type": "text", "index": false },
				"status": { "type": "short" },
				"score": { "type": "double" },
				"json": { "type": "text", "index": false },

				"start_date": { "type": "date", "format": "date_optional_time" },
				"end_date": { "type": "date", "format": "date_optional_time" },
				"week_slot": { "type": "text", "analyzer": "comma_analyzer" },
				"display_type": { "type": "keyword", "index": true },
				"display_weight": { "type": "integer" },
				"ipu": { "type": "integer" },
				"dipu": { "type": "integer" },
				"tipu": { "type": "integer" },

				"target_age_min": { "type": "short", "null_value": -32768 },
				"target_age_max": { "type": "short", "null_value": 32767 },
				"registered_days_min": { "type": "short", "null_value": -32768 },
				"registered_days_max": { "type": "short", "null_value": 32767 },
				"target_sdk_min": { "type": "integer", "null_value": -2147483648 },
				"target_sdk_max": { "type": "integer", "null_value": 2147483647 },
				"target_gender": { "type": "keyword", "index": true },
				"target_language": { "type": "keyword", "index": true },
				"target_carrier": { "type": "text", "analyzer": "comma_analyzer" },
				"target_region": { "type": "text", "analyzer": "comma_analyzer" },
				"custom_target_1": { "type": "text", "analyzer": "comma_analyzer" },
				"custom_target_2": { "type": "text", "analyzer": "comma_analyzer" },
				"custom_target_3": { "type": "text", "analyzer": "comma_analyzer" },

				"target_app": { "type": "text", "analyzer": "comma_analyzer" },
				"target_unit": { "type": "text", "analyzer": "comma_analyzer" },
				"target_app_id": { "type": "text", "analyzer": "comma_analyzer" },
				"target_org": { "type": "text", "analyzer": "comma_analyzer" },
				"target_os_min": { "type": "integer", "null_value": 2147483647 },
				"target_os_max": { "type": "integer", "null_value": 2147483647 },
				"target_battery_optimization": { "type": "boolean" },

				"detarget_app": { "type": "text", "analyzer": "comma_analyzer" },
				"detarget_unit": { "type": "text", "analyzer": "comma_analyzer" },
				"detarget_app_id": { "type": "text", "analyzer": "comma_analyzer" },
				"detarget_org": { "type": "text", "analyzer": "comma_analyzer" },

				"country": { "type": "keyword", "index": true },
				"timezone": { "type": "text", "index": false },
				"organization_id": { "type": "integer" },
				"owner_id": { "type": "integer" },
				"channel_id": { "type": "integer" },
				"provider_id": { "type": "integer" },
				"natural_id": { "type": "text", "index": false },
				"template_id": { "type": "short" },
				"is_enabled": { "type": "boolean" },
				"is_ctr_filter_off": { "type": "boolean" },
				"created_at": { "type": "date", "format": "date_optional_time" },
				"updated_at": { "type": "date", "format": "date_optional_time" },
				"published_at": { "type": "date", "format": "date_optional_time" },

				"creative_types": { "type": "text", "analyzer": "comma_analyzer" },
				"creative_links": { "type": "object", "enabled": "false" },
            	"channel": { "type": "object", "enabled": "false" },
            	"provider": {"type": "object", "enabled": "false"},
				"related": { "type": "integer" },
            	"related_count": {"type": "long"},
            	"image_width": { "type": "integer" },
            	"image_height": { "type": "integer" },
				"image_ratio": {"type": "double"},
				"impressions": { "type": "integer" },
				"clicks": { "type": "integer" }
			}
   		}
    }
}
`
	//TestAppIDUnknown
	TestAppID1 int64 = iota
	TestAppID2
	TestAppID3
	TestAppID4
	TestAppID5
	TestAppIDAdOnly
	TestAppIDContentOnly
	TestAppIDContentUnitOnly

	HsKrAppID           int64 = 100000043
	HsZzAppID           int64 = 100000060
	DeactivatedAppID    int64 = 1234567890
	HsKrFeedUnitID      int64 = 31605932142403
	HsKrAppPackage            = "com.buzzvil.adhours"
	SlidejoyAppPackage        = "com.slidejoy"
	ContentFilterWords        = "필터링단어,filteringWord,이거 뭐"
	FilteringProviderID int64 = 9999
)

var (
	filteringProviderIDStr = strconv.FormatInt(FilteringProviderID, 10)
	// KoreaUnit is a Korean Unit
	KoreaUnit app.Unit
	// GlobalUnit is a global Unit
	GlobalUnit app.Unit
	// ContentChannelMap var definition
	ContentChannelMap = make(map[int64]*model.ContentChannel)
	// ContentProviderMap var definition
	ContentProviderMap = make(map[int64]*model.ContentProvider)
)

// SetupElasticSearch func definition
func SetupElasticSearch() func() {
	core.Logger.Println("setupElasticSearch")

	// 아주 혹시 모를 상황에 디비 날리지 못하도록
	if env.IsTest() == false {
		core.Logger.Fatal("Not in test env..")
	}

	client := buzzscreen.Service.ES
	client.DeleteIndex(env.Config.ElasticSearch.CampaignIndexName).Do(context.Background())

	createIndex, err := client.CreateIndex(env.Config.ElasticSearch.CampaignIndexName).Body(testMapping).Do(context.Background())
	if err != nil {
		core.Logger.Fatal(err)
	}
	if createIndex == nil {
		core.Logger.Fatalf("expected result to be != nil; got: %v", createIndex)
	}

	return func() {
		core.Logger.Println("teardown setupElasticSearch")
	}
}

// SetupDatabase func definition
func SetupDatabase() func() {
	core.Logger.Println("setupDatabase")

	// 아주 혹시 모를 상황에 디비 날리지 못하도록
	if env.IsTest() == false {
		core.Logger.Fatal("Not in test env..")
	}

	db := buzzscreen.Service.DB

	core.Logger.Infof("SetupDatabase() - Create tables...")

	dropAndCreateTables(db, &dbapp.App{}, &dbapp.Unit{}, &dbdevice.Device{}, &model.DeviceUser{},
		&dbdevice.DeviceUpdateHistory{}, &model.ContentCampaign{},
		&model.ContentChannel{}, &model.ContentProvider{}, &model.WelcomeReward{}, &dbapp.WelcomeRewardConfig{},
		&model.NotificationSchedule{})

	core.Logger.Infof("SetupDatabase() - Insert rows...")

	setupAppAndUnit(db)
	setupProviderAndChannelTable(db)

	core.Logger.Infof("SetupDatabase() - Dynamo DB")

	// TODO: Detach real dynamo db dependency from CI and remove this condition. The
	// dynamo table should be deleted and created again.
	if len(env.Config.DynamoHost) > 0 {
		deleteDynamoDBTable()
		createDynamoDBTable()
	}

	return func() {
		core.Logger.Println("tearDownDatabase")
	}
}

// SetDeviceUnitRegisteredSeconds  Will remove registered seconds of other units if those record exist
func SetDeviceUnitRegisteredSeconds(deviceID int64, unitID int64, unitRegisteredSeconds int64) error {
	keyUnitRegisteredSecondsPrefix := "urd:" // copied from "internal/pkg/device/profilerepo.go"

	profileTable := buzzscreen.Service.DynamoDB.Table(env.Config.DynamoTableProfile)

	profileRecord := deviceRepo.DynamoDP{
		DeviceID:     deviceID,
		Profile:      keyUnitRegisteredSecondsPrefix + strconv.FormatInt(unitID, 10),
		ProfileValue: strconv.FormatInt(unitRegisteredSeconds, 10),
	}
	return profileTable.Put(&profileRecord).Run()
}

// SetDeviceRegisteredSeconds set to specified unit time stamp
func SetDeviceRegisteredSeconds(deviceID int64, unitRegisteredSeconds int64) error {
	keyUnitRegisteredSeconds := "rd" // copied from "internal/pkg/device/profilerepo.go"

	profileTable := buzzscreen.Service.DynamoDB.Table(env.Config.DynamoTableProfile)
	profileRecord := deviceRepo.DynamoDP{
		DeviceID:     deviceID,
		Profile:      keyUnitRegisteredSeconds,
		ProfileValue: strconv.FormatInt(unitRegisteredSeconds, 10),
	}
	return profileTable.Put(&profileRecord).Run()
}

// GetProfileByID get device profile
func GetProfileByID(deviceID int64) *device.Profile {
	profileTable := buzzscreen.Service.DynamoDB.Table(env.Config.DynamoTableProfile)
	profile, _ := deviceRepo.NewProfileRepo(&profileTable).GetByID(deviceID)
	return profile
}

func dropAndCreateTables(db *gorm.DB, models ...interface{}) {
	for _, dbModel := range models {
		//core.Logger.Printf("Drop %v table\n", reflect.TypeOf(dbModel))
		db = db.DropTableIfExists(dbModel)
		if db.Error != nil {
			core.Logger.Fatal(db.Error)
		}
		//core.Logger.Printf("Create %v table\n", reflect.TypeOf(dbModel))
		db = db.CreateTable(dbModel)
		if db.Error != nil {
			core.Logger.Fatal(db.Error)
		}
	}
}

// CreateDynamoDBTable creates dynamo DB table of Profile.
func createDynamoDBTable() {
	dyDB := buzzscreen.Service.DynamoDB
	if err := dyDB.CreateTable(env.Config.DynamoTableProfile, deviceRepo.DynamoDP{}).Run(); err != nil {
		core.Logger.Println(err)
	}
}

// DeleteDynamoDBTable deletes dynamo DB table of Profile.
func deleteDynamoDBTable() {
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String("ap-northeast-1"),
		Endpoint: aws.String(env.Config.DynamoHost),
	})

	if err != nil {
		core.Logger.Println(err)
	}

	svc := dynamodb.New(sess)
	if _, err := svc.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: &env.Config.DynamoTableProfile,
	}); err != nil {
		core.Logger.Println(err)
	}
}

var (
	adTypeAll, adTypeNone                                = dbapp.AdType(app.AdTypeAll), dbapp.AdType(app.AdTypeNone)
	contentTypeAll, contentTypeNone, contentTypeUnitOnly = dbapp.ContentType(app.ContentTypeAll), dbapp.ContentType(app.ContentTypeNone), dbapp.ContentType(app.ContentTypeUnitOnly)
)

func getApp(appID int64) *dbapp.App {
	return &dbapp.App{
		ID:        appID,
		IsEnabled: true,
	}
}

func getDBUnitWithAppID(appID int64) *dbapp.Unit {
	return &(dbapp.Unit{
		AppID:                appID,
		AdType:               &adTypeAll,
		BaseReward:           &(&struct{ x int }{2}).x,
		BaseInitPeriod:       &(&struct{ x int }{3600}).x,
		BuzzvilLandingReward: 1,
		ContentType:          &contentTypeAll,
		Country:              "KR",
		FilteredProviders:    &filteringProviderIDStr,
		FirstScreenRatio:     "9:1",
		FeedRatio:            "5:1",
		OrganizationID:       2,
		Platform:             "A",
		Timezone:             "Asia/Seoul",
		UnitType:             dbapp.UnitTypeLockscreen,
	})
}

func setupAppAndUnit(db *gorm.DB) {
	core.Logger.Println("Create test app")
	appIDs := []int64{TestAppID1, TestAppID2, TestAppID3, TestAppID4, TestAppID5, TestAppIDAdOnly, TestAppIDContentOnly, TestAppIDContentUnitOnly, HsKrAppID, HsZzAppID, HsKrFeedUnitID}
	for _, appID := range appIDs {
		err := db.Create(getApp(appID)).Error
		if err != nil {
			core.Logger.Fatal(err)
		}
	}

	deactivatedApp := getApp(DeactivatedAppID)
	deactivatedApp.IsEnabled = false
	err := db.Create(deactivatedApp).Error
	if err != nil {
		core.Logger.Fatal(err)
	}

	core.Logger.Println("Create test unit")
	dbKoreaUnit := getDBUnitWithAppID(HsKrAppID)
	if err := db.Create(dbKoreaUnit).Error; err != nil {
		core.Logger.Fatal(err)
	}
	KoreaUnit = unitToEntity(dbKoreaUnit)

	dbKrFeedUnit := getDBUnitWithAppID(HsKrAppID)
	dbKrFeedUnit.ID = HsKrFeedUnitID
	dbKrFeedUnit.UnitType = dbapp.UnitTypeNative
	if err := db.Create(dbKrFeedUnit).Error; err != nil {
		core.Logger.Fatal(err)
	}

	dbGlobalUnit := getDBUnitWithAppID(HsZzAppID)
	dbGlobalUnit.Country = ""
	dbGlobalUnit.Timezone = "UTC"
	if err := db.Create(dbGlobalUnit).Error; err != nil {
		core.Logger.Fatal(err)
	}
	GlobalUnit = unitToEntity(dbGlobalUnit)

	unit := getDBUnitWithAppID(TestAppID1)
	unit.FirstScreenRatio = "100:1"
	db = db.Create(unit)

	db = db.Create(getDBUnitWithAppID(TestAppID2))
	db = db.Create(getDBUnitWithAppID(TestAppID3))
	db = db.Create(getDBUnitWithAppID(TestAppID4))
	db = db.Create(getDBUnitWithAppID(TestAppID5))

	adOnlyUnit := getDBUnitWithAppID(TestAppIDAdOnly)
	adOnlyUnit.AdType = &adTypeAll
	adOnlyUnit.ContentType = &contentTypeNone

	db = db.Create(adOnlyUnit)

	contentOnlyUnit := getDBUnitWithAppID(TestAppIDContentOnly)
	contentOnlyUnit.AdType = &adTypeNone
	contentOnlyUnit.ContentType = &contentTypeAll

	db = db.Create(contentOnlyUnit)

	contentUnitOnlyUnit := getDBUnitWithAppID(TestAppIDContentUnitOnly)
	contentUnitOnlyUnit.AdType = &adTypeNone
	contentUnitOnlyUnit.ContentType = &contentTypeUnitOnly

	db = db.Create(contentUnitOnlyUnit)

}

func setupProviderAndChannelTable(db *gorm.DB) {
	cc0 := &model.ContentChannel{
		ID:       1,
		Category: "sports",
		Name:     "스포츠채널",
	}
	cc1 := &model.ContentChannel{
		ID:       2,
		Category: "politics",
		Name:     "BBC",
	}
	cp0 := &model.ContentProvider{
		Categories: "sports",
		ChannelID:  cc0.ID,
		Country:    "KR",
		Enabled:    "Y",
		Name:       "스포츠채널",
	}
	cp1 := &model.ContentProvider{
		Categories: "politics",
		ChannelID:  cc1.ID,
		Country:    "US",
		Enabled:    "Y",
		Name:       "BBC",
	}
	core.Logger.Println("Create test channel")
	db = db.Create(cc0).Create(cc1)
	core.Logger.Println("Create test provider")
	db = db.Create(cp0).Create(cp1)

	ContentChannelMap[cc0.ID] = cc0
	ContentChannelMap[cc1.ID] = cc1
	ContentProviderMap[cp0.ID] = cp0
	ContentProviderMap[cp1.ID] = cp1
}

func unitToEntity(dbUnit *dbapp.Unit) app.Unit {
	em := appRepo.EntityMapper{}
	return em.UnitToEntity(dbUnit)
}
