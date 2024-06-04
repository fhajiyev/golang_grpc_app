package service_test

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"gopkg.in/olivere/elastic.v5"
)

func intPtr(v int) *int {
	return &v
}

func createBaseContentCampaignWithID(id int64) *dto.ESContentCampaign {
	channelID := int64(1 + rand.Intn(len(tests.ContentChannelMap)))
	relatedCount := 1
	var channel model.ContentChannel
	err := buzzscreen.Service.DB.Where(&model.ContentChannel{ID: channelID}).First(&channel).Error
	if err != nil {
		panic(err)
	}
	now := time.Now()
	createdAt := now.Add(-time.Minute)
	timeFormat := "2006-01-02T15:04:05"
	imageRatio := 1.6

	ecc := &dto.ESContentCampaign{
		ContentCampaign: model.ContentCampaign{
			Categories:                "sports",
			ChannelID:                 &channelID,
			CleanMode:                 0,
			ClickURL:                  "http://daum.net",
			Country:                   "KR",
			CreatedAt:                 createdAt.Format(timeFormat),
			DisplayType:               "A",
			DisplayWeight:             10,
			EndDate:                   "2099-01-03T00:00:00",
			ID:                        id,
			Ipu:                       intPtr(5),
			Dipu:                      intPtr(10),
			IsEnabled:                 true,
			IsCtrFilterOff:            false,
			JSON:                      `{"imgW": 800, "imgH": 400, "unit": {"ic_type": "VIEW"}}`,
			LandingReward:             2,
			LandingType:               0,
			Name:                      "이것은 테스트 캠페인",
			OrganizationID:            0,
			OwnerID:                   0,
			ProviderID:                0,
			PublishedAt:               createdAt.Format(timeFormat),
			Status:                    model.StatusComplete,
			StartDate:                 "2017-01-02T00:00:00",
			Timezone:                  "Asia/Seoul",
			Title:                     "abc",
			Tipu:                      rand.Intn(100),
			Type:                      "C",
			UpdatedAt:                 now.Format(timeFormat),
			TargetAgeMin:              model.ESNullShortMin,
			TargetAgeMax:              model.ESNullShortMax,
			TargetSdkMin:              model.ESNullIntMin,
			TargetSdkMax:              model.ESNullIntMax,
			TargetUnit:                model.ESGlobString,
			RegisteredDaysMin:         model.ESNullShortMin,
			RegisteredDaysMax:         model.ESNullShortMax,
			TargetOsMin:               model.ESNullIntMin,
			TargetOsMax:               model.ESNullIntMax,
			TargetBatteryOptimization: false,
		},
		Channel: &dto.ESContentChannel{
			ID:   channel.ID,
			Name: channel.Name,
			Logo: channel.Logo,
		},
		Clicks:        10,
		CreativeTypes: "A,R",
		CreativeLinks: map[string][]string{"A": {"http://abc-A.jpg"}, "R": {"http://abc-R.jpg"}},
		Related:       &id,
		RelatedCount:  &relatedCount,
		ImageRatio:    &imageRatio,
		Impressions:   100,

		DetargetApp:   model.ESGlobString,
		DetargetUnit:  model.ESGlobString,
		DetargetAppID: model.ESGlobString,
		DetargetOrg:   model.ESGlobString,
	}

	return ecc
}

func createBaseContentCampaign() *dto.ESContentCampaign {
	return createBaseContentCampaignWithID(1)
}

func createBaseContentCampaigns(size int) []*dto.ESContentCampaign {
	ccs := make([]*dto.ESContentCampaign, 0)
	for i := 0; i < size; i++ {
		camp := createBaseContentCampaignWithID(int64(i + 1))
		camp.ID = int64(i + 1)
		ccs = append(ccs, camp)
	}
	return ccs
}

func deleteContentCampaignsFromESAndDB(t *testing.T, contentCampaigns ...*dto.ESContentCampaign) {
	client := buzzscreen.Service.ES
	bulkRequest := client.Bulk()
	db := buzzscreen.Service.DB

	for _, cc := range contentCampaigns {
		bulkRequest = bulkRequest.Add(elastic.NewBulkDeleteRequest().Index(env.Config.ElasticSearch.CampaignIndexName).
			Type("content_campaign").
			Id(strconv.FormatInt(cc.ID, 10)))
		err := db.Delete(&cc.ContentCampaign).Error
		if err != nil {
			t.Fatal(err)
		}
	}

	bulkResponse, err := bulkRequest.Do(context.Background())

	if err != nil {
		t.Fatal(err)
	} else {
		for _, item := range bulkResponse.Deleted() {
			t.Logf("deleteContentCampaignsFromESAndDB() - deleteRes: %+v", *item)
		}
	}
}

func insertContentCampaignsToESAndDB(t *testing.T, contentCampaigns ...*dto.ESContentCampaign) {
	client := buzzscreen.Service.ES
	bulkRequest := client.Bulk()
	db := buzzscreen.Service.DB

	for _, cc := range contentCampaigns {
		t.Logf("insertContentCampaignsToESAndDB() - ccID: %v", cc.ID)
		bulkRequest = bulkRequest.Add(elastic.NewBulkIndexRequest().
			Index(env.Config.ElasticSearch.CampaignIndexName).
			Type("content_campaign").
			Id(strconv.FormatInt(cc.ID, 10)).
			Doc(cc.GetDocToCreate())).
			Refresh("true")
		err := db.Save(&cc.ContentCampaign).Error
		if err != nil {
			t.Fatal(err)
		}
	}

	bulkResponse, err := bulkRequest.Do(context.Background())

	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("insertContentCampaignsToESAndDB() - bulkResponse: %v", *bulkResponse)
	}
}

func yearOfBirth(age int) int {
	return time.Now().Year() - age - 1
}
