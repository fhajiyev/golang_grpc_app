package controller_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/Buzzvil/go-test/test"
)

func TestGetContentArticlesFeed(t *testing.T) {
	_, removeFunc := getPatchedHttpClient()
	defer removeFunc()

	camp := createBaseContentCampaignWithID(1)
	insertContentCampaignsToESAndDB(t, camp)
	defer deleteContentCampaignsFromESAndDB(t, camp)

	newContentArticleTestCase(t, "TestGetContentArticlesFeed", func(contentCase *ContentTestCase) {
		contentCase.params.Set("types", `{"NATIVE":[]}`)
		contentCase.params.Set("unit_id", strconv.FormatInt(tests.HsKrFeedUnitID, 10))
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		for _, article := range res.Result.ContentArticles {
			compareCampaignWithArticle(tc, camp, article)
			compareCampaignWithArticleFeed(tc, camp, article)
		}
		return len(res.Result.ContentArticles) == 1
	})
	camp2 := createBaseContentCampaignWithID(2)
	insertContentCampaignsToESAndDB(t, camp2)
	defer deleteContentCampaignsFromESAndDB(t, camp2)

	newContentArticleTestCase(t, "TestGetContentArticlesFeed", func(contentCase *ContentTestCase) {
		contentCase.params.Set("types", `{"NATIVE":[]}`)
		contentCase.params.Set("unit_id", strconv.FormatInt(tests.HsKrFeedUnitID, 10))
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		return len(res.Result.ContentArticles) == 2
	})
}

func TestGetContentArticlesBookmark(t *testing.T) {
	_, removeFunc := getPatchedHttpClient()
	defer removeFunc()

	camp := createBaseContentCampaignWithID(1)
	insertContentCampaignsToESAndDB(t, camp)
	defer deleteContentCampaignsFromESAndDB(t, camp)

	newContentArticleTestCase(t, "TestGetContentArticlesBookmark", func(contentCase *ContentTestCase) {
		contentCase.params.Set("types", `{"NATIVE":[]}`)
		contentCase.params.Set("unit_id", strconv.FormatInt(tests.HsKrFeedUnitID, 10))
		contentCase.params.Set("ids", strconv.FormatInt(camp.ID, 10))
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		return len(res.Result.ContentArticles) == 1
	})
	camp2 := createBaseContentCampaignWithID(2)
	insertContentCampaignsToESAndDB(t, camp2)
	defer deleteContentCampaignsFromESAndDB(t, camp2)

	newContentArticleTestCase(t, "TestGetContentArticlesBookmark", func(contentCase *ContentTestCase) {
		contentCase.params.Set("types", `{"NATIVE":[]}`)
		contentCase.params.Set("unit_id", strconv.FormatInt(tests.HsKrFeedUnitID, 10))
		contentCase.params.Set("ids", strings.Join([]string{strconv.FormatInt(camp.ID, 10), strconv.FormatInt(camp2.ID, 10)}, ","))
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		return len(res.Result.ContentArticles) == 2
	})
}

func TestGetContentArticlesStatus(t *testing.T) {
	_, removeFunc := getPatchedHttpClient()
	defer removeFunc()

	size := 14
	camps := createBaseContentCampaigns(size)
	for i, camp := range camps {
		camp.Status = model.Status(i%7 + 1)
	}
	insertContentCampaignsToESAndDB(t, camps...)
	defer deleteContentCampaignsFromESAndDB(t, camps...)

	newContentArticleTestCase(t, "TestGetContentArticlesStatus", func(contentCase *ContentTestCase) {
		contentCase.params.Set("types", `{"NATIVE":[]}`)
		contentCase.params.Set("unit_id", strconv.FormatInt(tests.HsKrFeedUnitID, 10))
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		test.AssertEqual(tc.t, len(res.Result.ContentArticles), 12, tc.name+" - Feed")
		return true
	})

	newContentArticleTestCase(t, "TestGetContentArticlesStatus", func(contentCase *ContentTestCase) {
		contentCase.params.Set("types", `{"IMAGE":["FULLSCREEN"]}`)
		contentCase.params.Set("unit_id", strconv.FormatInt(tests.KoreaUnit.ID, 10))
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		test.AssertEqual(tc.t, len(res.Result.ContentArticles), 4, tc.name+" - Lockscreen")
		return true
	})
}

/**
1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14
Filtered: 4, 11 (by status)
Order: 1, 3, 5, 7, 9, 13, 2, 6, 8, 10, 12, 14  (by related)
*/
func TestGetContentArticlesRelatedFiltering(t *testing.T) {
	_, removeFunc := getPatchedHttpClient()
	defer removeFunc()

	size := 14
	var idUnique int64 = 0
	camps := createBaseContentCampaigns(size)
	for i, camp := range camps {
		camp.Status = model.Status(i%7 + 1)
		if camp.ID%2 == 0 {
			camp.Related = &idUnique
		}
	}
	insertContentCampaignsToESAndDB(t, camps...)
	defer deleteContentCampaignsFromESAndDB(t, camps...)

	newContentArticleTestCase(t, "TestGetContentArticlesRelatedFiltering", func(contentCase *ContentTestCase) {
		contentCase.params.Set("types", `{"NATIVE":[]}`)
		contentCase.params.Set("unit_id", strconv.FormatInt(tests.HsKrFeedUnitID, 10))
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		return len(res.Result.ContentArticles) == 12
	})
}

func TestGetContentArticlesPagination(t *testing.T) {
	_, removeFunc := getPatchedHttpClient()
	defer removeFunc()

	camps := createBaseContentCampaigns(10)
	pageSize := 5
	insertContentCampaignsToESAndDB(t, camps...)
	defer deleteContentCampaignsFromESAndDB(t, camps...)

	newContentArticleTestCase(t, "TestGetContentArticlesPagination - Page 0", func(contentCase *ContentTestCase) {
		contentCase.params.Set("types", `{"NATIVE":[]}`)
		contentCase.params.Set("unit_id", strconv.FormatInt(tests.HsKrFeedUnitID, 10))
		contentCase.params.Set("size", strconv.Itoa(pageSize))
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		tc.t.Logf("%s - res: %+v", tc.name, *res)
		queryKey := *res.Result.QueryKey

		for i, article := range res.Result.ContentArticles {
			compareCampaignWithArticle(tc, camps[i], article)
			compareCampaignWithArticleFeed(tc, camps[i], article)
		}

		newContentArticleTestCase(t, "TestGetContentArticlesPagination - Page 1", func(contentCase *ContentTestCase) {
			contentCase.params.Set("types", `{"NATIVE":[]}`)
			contentCase.params.Set("unit_id", strconv.FormatInt(tests.HsKrFeedUnitID, 10))
			contentCase.params.Set("size", strconv.Itoa(pageSize))
			contentCase.params.Set("query_key", queryKey)
		}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
			tc.t.Logf("%s - res: %+v", tc.name, *res)

			for i, article := range res.Result.ContentArticles {
				compareCampaignWithArticle(tc, camps[i+pageSize], article)
				compareCampaignWithArticleFeed(tc, camps[i+pageSize], article)
			}
			test.AssertEqual(tc.t, len(res.Result.ContentArticles), pageSize, tc.name+" - len")
			return res.Result.QueryKey == nil
		})

		return len(res.Result.ContentArticles) == pageSize && queryKey != ""
	})
}

func TestGetContentArticlesLockScreen(t *testing.T) {
	_, removeFunc := getPatchedHttpClient()
	defer removeFunc()

	pageSize := 2
	camps := createBaseContentCampaigns(pageSize)
	insertContentCampaignsToESAndDB(t, camps...)
	defer deleteContentCampaignsFromESAndDB(t, camps...)

	newContentArticleTestCase(t, "TestGetContentArticlesLockScreen", func(contentCase *ContentTestCase) {
		contentCase.params.Set("types", `{"IMAGE":["FULLSCREEN"]}`)
		contentCase.params.Set("unit_id", strconv.FormatInt(tests.KoreaUnit.ID, 10))
		contentCase.params.Set("size", strconv.Itoa(pageSize))
		contentCase.params.Del("birthday")
		contentCase.params.Del("gender")
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		tc.t.Logf("%s - res: %v", tc.name, *res)
		for i, article := range res.Result.ContentArticles {
			compareCampaignWithArticle(tc, camps[i], article)
			compareCampaignWithArticleLockScreen(tc, camps[i], article)
		}
		return len(res.Result.ContentArticles) == pageSize
	})
}

func TestGetContentArticlesCategories(t *testing.T) {
	_, removeFunc := getPatchedHttpClient()
	defer removeFunc()

	pageSize := 5
	camps := createBaseContentCampaigns(pageSize)
	camps[0].Categories = "sports,lifestyle"
	camps[1].Categories = "sports"
	camps[2].Categories = "lifestyle"
	camps[3].Categories = "love,lifestyle"
	camps[4].Categories = "love" // exclude
	expected := 4
	insertContentCampaignsToESAndDB(t, camps...)
	defer deleteContentCampaignsFromESAndDB(t, camps...)

	newContentArticleTestCase(t, "TestGetContentArticlesCategories", func(contentCase *ContentTestCase) {
		contentCase.params.Set("categories", "lifestyle,sports")
	}).run(func(tc *ContentTestCase, res *TestContentArticlesResponse) bool {
		tc.t.Logf("%s - res: %v", tc.name, *res)
		for i, article := range res.Result.ContentArticles {
			compareCampaignWithArticle(tc, camps[i], article)
		}
		return len(res.Result.ContentArticles) == expected
	})
}

func compareCampaignWithArticle(tc *ContentTestCase, camp *dto.ESContentCampaign, article *dto.ContentArticle) {
	test.AssertEqual(tc.t, article.ID, camp.ID, tc.name+" - id")
	test.AssertEqual(tc.t, article.Creative["click_url"] != nil, true, tc.name+" - click_url")
	test.AssertEqual(tc.t, article.CleanMode, camp.CleanMode, tc.name+" - cleanMode")
	test.AssertEqual(tc.t, article.SourceURL, camp.ClickURL, tc.name+" - SourceURL")
	test.AssertEqual(tc.t, article.Creative["title"], camp.Title, tc.name+" - title")
	test.AssertEqual(tc.t, article.Creative["description"], camp.Description, tc.name+" - description")
	test.AssertEqual(tc.t, article.CreatedAt, utils.ConvertToUnixTime(camp.PublishedAt), tc.name+" - CreatedAt")

	test.AssertEqual(tc.t, int(article.Creative["landing_type"].(float64))-int(camp.LandingType) == 0, true, tc.name+" - LandingType")
}

func compareCampaignWithArticleFeed(tc *ContentTestCase, camp *dto.ESContentCampaign, article *dto.ContentArticle) {
	jsonMap := make(map[string]interface{})
	json.Unmarshal([]byte(camp.JSON), &jsonMap)

	test.AssertEqual(tc.t, article.Creative["type"], "NATIVE", tc.name+" - type")
	test.AssertEqual(tc.t, article.Creative["image_url"], camp.CreativeLinks["R"][0], tc.name+" - ImageURL")
	test.AssertEqual(tc.t, article.Channel.ID, *camp.ChannelID, tc.name+" - Channel ID")
	test.AssertEqual(tc.t, article.Channel.Name, tests.ContentChannelMap[*camp.ChannelID].Name, tc.name+" - Channel Name")
	test.AssertEqual(tc.t, article.Creative["icon_url"].(string), tests.ContentChannelMap[*camp.ChannelID].Logo, tc.name+" - Channel Logo")

	if len(jsonMap) > 0 {
		test.AssertEqual(tc.t, article.Creative["width"].(float64)-jsonMap["imgW"].(float64) == 0, true, tc.name+" - width")
		test.AssertEqual(tc.t, article.Creative["height"].(float64)-jsonMap["imgH"].(float64) == 0, true, tc.name+" - height")
	}
}

func compareCampaignWithArticleLockScreen(tc *ContentTestCase, camp *dto.ESContentCampaign, article *dto.ContentArticle) {
	test.AssertEqual(tc.t, article.Creative["type"], "IMAGE", tc.name+" - type")
	test.AssertEqual(tc.t, article.Creative["width"].(float64)-720 == 0, true, tc.name+" - width")
	test.AssertEqual(tc.t, article.Creative["height"].(float64)-1230 == 0, true, tc.name+" - height")
	test.AssertEqual(tc.t, article.Creative["size_type"], "FULLSCREEN", tc.name+" - sizeType")
	test.AssertEqual(tc.t, article.Creative["image_url"], camp.CreativeLinks["A"][0], tc.name+" - ImageURL")
}

type (
	// ContentTestCase type definition
	ContentTestCase struct {
		t      *testing.T
		name   string
		params *url.Values
	}
)

func newContentArticleTestCase(t *testing.T, name string, builder func(*ContentTestCase)) *ContentTestCase {
	requestParams, _ := buildV3BaseTestRequest(t)
	cac := &ContentTestCase{
		name:   name,
		params: requestParams,
		t:      t,
	}

	if builder != nil {
		builder(cac)
	}

	return cac
}

type (
	// TestContentArticlesResponse type definition
	TestContentArticlesResponse struct {
		Result dto.ContentArticlesResponse `json:"result"`
		Code   int                         `json:"code"`
	}
)

func (tc *ContentTestCase) run(equalFunc func(tc *ContentTestCase, res *TestContentArticlesResponse) bool) {
	var articlesResponse TestContentArticlesResponse
	t := tc.t
	statusCode, err := (&network.Request{
		Method: "GET",
		Params: tc.params,
		URL:    ts.URL + "/api/v3/content/articles",
	}).GetResponse(&articlesResponse)

	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if err != nil || statusCode != 200 {
		t.Fatal(statusCode, err, articlesResponse)
	} else {
		test.AssertEqual(t, equalFunc(tc, &articlesResponse), true, fmt.Sprintf("ContentTestCase - %v", tc.name))
	}
}
