package service_test

import (
	"strconv"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/Buzzvil/go-test/test"
)

func TestGetContentChannelsByID(t *testing.T) {
	var strIDs string
	for k := range tests.ContentChannelMap {
		strIDs += strconv.FormatInt(k, 10) + ","
	}
	strIDs = strIDs[:len(strIDs)-1]
	channelQuery := model.NewContentChannelsQuery().WithIDs(strIDs)
	channels := service.GetContentChannels(channelQuery)
	for _, channelFromRes := range *channels {
		channelFromMap := tests.ContentChannelMap[channelFromRes.ID]
		assertChannelEqual(t, channelFromRes, channelFromMap)
	}
}

func TestGetContentChannelsByCountry(t *testing.T) {
	testGetContentChannelsByCountry(t, "KR")
	testGetContentChannelsByCountry(t, "US")
}

func testGetContentChannelsByCountry(t *testing.T, country string) {
	krChannelMap := make(map[int64]*model.ContentChannel)
	for _, v := range tests.ContentProviderMap {
		if v.Country == country {
			krChannelMap[v.ChannelID] = tests.ContentChannelMap[v.ChannelID]
		}
	}
	channelQuery := model.NewContentChannelsQuery().WithCountryAndCategoryID(country, "")
	channels := service.GetContentChannels(channelQuery)
	for _, channelFromRes := range *channels {
		channelFromMap := tests.ContentChannelMap[channelFromRes.ID]
		assertChannelEqual(t, channelFromRes, channelFromMap)
	}
}

func assertChannelEqual(t *testing.T, c1, c2 *model.ContentChannel) {
	test.AssertEqual(t, c1.Category, c2.Category, "TestGetContentChannels - Category")
	test.AssertEqual(t, c1.Logo, c2.Logo, "TestGetContentChannels - Logo")
	test.AssertEqual(t, c1.Name, c2.Name, "TestGetContentChannels - Name")
}
