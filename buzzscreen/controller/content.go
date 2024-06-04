package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/recovery"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/ifa"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/asaskevich/govalidator"
)

// GetContentCategories func definition
func GetContentCategories(c core.Context) error {
	var contentReq dto.ContentCategoriesRequest
	if err := bindRequestSupport(c, &contentReq, &ContentV2CategoriesRequest{}); err != nil {
		return err
	}

	err := contentReq.UnpackSession()
	if err != nil {
		return common.NewSessionError(err)
	}

	ok, err := buzzscreen.Service.DeviceUseCase.ValidateUnitDeviceToken(contentReq.Session.UserID)
	if !ok {
		core.Logger.Warnf("GetContentCategories() - failed to validate unit_device_token. err: %s. req: %+v", err.Error(), contentReq)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	categories, err := service.GetCategories(contentReq.GetLanguage())

	if err != nil {
		core.Logger.WithError(err).WithField("user", contentReq.Locale).Errorf("GetContentCategories()")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"code": dto.CodeOk,
		"result": map[string]interface{}{
			"categories": getResponseSupport(c, *categories),
		},
	})
}

// GetContentArticles func definition
func GetContentArticles(c core.Context) error {
	ctx := c.Request().Context()
	var req dto.ContentArticlesRequest
	if err := bindRequestSupport(c, &req, &ContentV2ArticlesRequest{}); err != nil {
		return err
	}

	if ifa.ShouldReplaceIFAWithIFV(req.AdID, req.IFV) {
		req.AdID = ifa.GetDeviceIFV(*req.IFV)
	}

	err := req.UnpackSession()
	if err != nil {
		return common.NewSessionError(err)
	}

	ok, err := buzzscreen.Service.DeviceUseCase.ValidateUnitDeviceToken(req.Session.UserID)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	unit := req.GetUnit(ctx)
	if unit == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "unit can't be nil"})
	} else if !unit.IsActive {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "unit is inactive"})
	}

	profile := req.GetDynamoProfile()
	if profile != nil {
		service.UpdateProfileUnitRegisterSeconds(profile, unit.ID, req.Session.AppID, c.Path())
		service.GiveWelcomeReward(ctx, profile, req.Session.UserID, unit.ID, req.GetCountry(ctx))
	}

	categoryMap, err := service.GetCategoriesMap(req.GetLanguage())
	if err != nil {
		return err
	}

	var res dto.ContentArticlesResponse
	if req.IDs != "" {
		res.ContentArticles = getContentArticlesByIDs(ctx, &req, categoryMap)
	} else {
		var queryKey *dto.ContentQueryKey

		res.ContentArticles, queryKey, err = getContentArticlesFromES(c, &req, categoryMap)
		if err != nil {
			switch err.(type) {
			case contentcampaign.RemoteESError:
				core.Logger.WithError(err).Warnf("controller.GetContentArticles() - err: %s", err)
				return common.NewInternalServerError(err)
			default:
				core.Logger.WithError(err).Errorf("controller.GetContentArticles() - err: %s", err)
				return common.NewInternalServerError(err)
			}
		}
		if queryKey != nil {
			keyStr := queryKey.Marshal()
			res.QueryKey = &keyStr
		}
	}

	logGetContentArticles(ctx, req, res)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":   dto.CodeOk,
		"result": getResponseSupport(c, res),
	})
}

func logGetContentArticles(ctx context.Context, req dto.ContentArticlesRequest, res dto.ContentArticlesResponse) {
	var campaignIDs []int64
	for _, c := range res.ContentArticles {
		campaignIDs = append(campaignIDs, c.ID)
	}

	core.Logger.Infof("/api/v3/content/articles - Request{AppID:%d, UnitID:%d, DeviceID:%d, IFA:%s, UnitDeviceToken:%s, SDKVersion:%d, OSVersion:%d} Response{Contents:%v}",
		req.Session.AppID,
		req.GetUnit(ctx).ID,
		req.Session.DeviceID,
		req.AdID,
		req.Session.UserID,
		req.SdkVersion,
		req.OsVersion,
		campaignIDs,
	)
}

func getContentArticlesFromES(c core.Context, contentReq *dto.ContentArticlesRequest, categoryMap map[string]*model.ContentCategory) (dto.ContentArticles, *dto.ContentQueryKey, error) {
	ctx := c.Request().Context()
	contentArticles := make(dto.ContentArticles, 0)
	var queryKeyFrom *dto.ContentQueryKey
	var err error
	currentTime := time.Now().Unix()
	dp := contentReq.GetDynamoProfile()
	campaignIDs := make([]int64, 0)
	if queryKeyFrom, err = contentReq.GetQueryKey(); err != nil {
		return nil, nil, err
	}
	idMap := make(map[int64]bool)

	if queryKeyFrom != nil {
		currentTime = queryKeyFrom.CreatedAt

		if dp != nil && dp.ScoredCampaignsKey != nil && *queryKeyFrom.ScoredCampaignsKey != *dp.ScoredCampaignsKey {
			core.Logger.WithField("http_request", c.Request()).Warnf("ScoredCampaigns Key is changed. deviceID: %v, queryKeyFrom.ScoredCampaignsKey: %v, dp.ScoredCampaignsKey: %v ", contentReq.Session.DeviceID, *queryKeyFrom.ScoredCampaignsKey, *dp.ScoredCampaignsKey)
		}

		//TODO: 중복이 더이상 발생하지 않으면 아래 검사로직은 제거해도 된다
		for _, id := range queryKeyFrom.CampaignIDs {
			idMap[id] = true
		}
	}

	esContentCamps, totalSize, err := service.GetContentCampaignsFromES(ctx, contentReq)
	if err != nil {
		return nil, nil, contentcampaign.RemoteESError{Err: err}
	}
	nextItemIndex := len(esContentCamps)
	if queryKeyFrom != nil {
		nextItemIndex += queryKeyFrom.Index
	}

	for _, esContent := range esContentCamps {
		if _, ok := idMap[esContent.ID]; ok {
			core.Logger.Warnf("GetContentArticles() - Duplicated. dID: %v, cID: %v, queryKeyString: %v, queryKeyValue: %+v", contentReq.Session.DeviceID, esContent.ID, contentReq.EncryptedQueryKey, queryKeyFrom)
			continue
		}
		var channel *model.ContentChannel
		if esContent.Channel != nil {
			channel = &(model.ContentChannel{
				ID:   esContent.Channel.ID,
				Logo: esContent.Channel.Logo,
				Name: esContent.Channel.Name,
			})
		} else if esContent.ChannelID != nil {
			channel = service.GetChannel(*(esContent.ChannelID))
		}

		esCreativeType := ""
		if _, ok := (*contentReq.GetTypes())["NATIVE"]; ok {
			esCreativeType = "R"
		} else if _, ok := (*contentReq.GetTypes())["IMAGE"]; ok {
			esCreativeType = "A"
		}

		imageURL := esContent.GetCDNImageURL(esCreativeType, contentReq.SdkVersion)

		if contentReq.OsVersion >= 14 && contentReq.Os != "ios" {
			imageURL = strings.Replace(imageURL, "jpeg", "webp", -1)
		}

		article := parseContentCampaignToContentArticle(&(esContent.ContentCampaign), channel, imageURL, contentReq.GetTypes(), categoryMap)

		if article == nil {
			continue
		}

		article.SetImpClickPayload(ctx, &(esContent.ContentCampaign), contentReq)

		campaignIDs = append(campaignIDs, article.ID)
		contentArticles = append(contentArticles, article)
	}

	var queryKeyTo *dto.ContentQueryKey
	if nextItemIndex < (totalSize) {
		queryKeyTo = &dto.ContentQueryKey{
			CreatedAt:   currentTime,
			Index:       nextItemIndex,
			CampaignIDs: campaignIDs,
		}
		if dp != nil && dp.ScoredCampaignsKey != nil {
			queryKeyTo.ScoredCampaignsKey = dp.ScoredCampaignsKey
		}
	}

	return contentArticles, queryKeyTo, nil
}

func getContentArticlesByIDs(ctx context.Context, contentReq *dto.ContentArticlesRequest, categoryMap map[string]*model.ContentCategory) dto.ContentArticles {
	contentArticles := make(dto.ContentArticles, 0)
	contentCamps := service.GetContentCampaignsFromDB(strings.Split(contentReq.IDs, ","))
	channelMap := service.GetChannelMap(contentCamps)
	for _, content := range contentCamps {
		imageURL := content.Image
		if contentReq.OsVersion >= 14 && contentReq.Os != "ios" {
			imageURL = strings.Replace(imageURL, "jpeg", "webp", -1)
		}
		if article := parseContentCampaignToContentArticle(content, channelMap[*content.ChannelID], imageURL, contentReq.GetTypes(), categoryMap); article != nil {
			article.SetImpClickPayload(ctx, content, contentReq)
			contentArticles = append(contentArticles, article)
		} else {
			core.Logger.WithField("http_request", contentReq.Request).Warnf("GetContentArticles() - article is nil. cID: %v", content.ID)
			continue
		}
	}
	return contentArticles
}

func parseContentCampaignToContentArticle(cc *model.ContentCampaign, channel *model.ContentChannel, imageURL string, types *dto.CreativeTypes, categoryMap map[string]*model.ContentCategory) *dto.ContentArticle {
	defer recovery.LimitedLogRecoverWith(*cc, 100)
	publishedAtUnix := utils.ConvertToUnixTime(cc.PublishedAt)

	article := &dto.ContentArticle{
		Common: dto.Common{
			ID: cc.ID,
			//LandingType: cc.LandingType,
			Creative: map[string]interface{}{
				"click_url":    "", // 외부에서 생성
				"description":  cc.Description,
				"image_url":    imageURL,
				"title":        cc.Title,
				"landing_type": cc.LandingType,
				"filterable":   true,
			},
		},
		Channel:   channel,
		Category:  categoryMap[cc.Categories],
		CleanMode: cc.CleanMode,
		CreatedAt: publishedAtUnix,
		Name:      cc.Name,
		SourceURL: cc.ClickURL,
	}
	if _, ok := (*types)[dto.TypeNative]; ok {
		article.Creative["type"] = dto.TypeNative
		extraData := *cc.GetExtraData()
		for key, value := range extraData {
			switch key {
			case "imgW":
				article.Creative["width"] = value
			case "imgH":
				article.Creative["height"] = value
			case "videoId", "videoSource":
				article.Creative[govalidator.CamelCaseToUnderscore(key)] = value
			}
		}
		if article.Channel != nil {
			article.Creative["icon_url"] = article.Channel.Logo
		}
	} else if _, ok := (*types)[dto.TypeImage]; ok {
		article.Creative["type"] = dto.TypeImage
		article.Creative["width"] = 720
		article.Creative["height"] = 1230
		article.Creative["size_type"] = "FULLSCREEN"
	}

	var unitExtraData map[string]interface{}
	if cc.JSON != "{}" {
		var esExtra dto.ESUnitExtra
		if err := json.Unmarshal([]byte(cc.JSON), &esExtra); err == nil {
			unitExtraData = esExtra.Unit
		} else {
			core.Logger.WithError(err).Errorf("parseContentCampaignToContentArticle() - ContentCampaign.json: %v", cc.JSON)
		}
	}

	if len(unitExtraData) > 0 {
		article.Extra = unitExtraData
	}

	return article
}

// GetContentChannels func definition
func GetContentChannels(c core.Context) error {
	ctx := c.Request().Context()
	var contentReq dto.GetContentChannelsRequest
	if err := bindRequestSupport(c, &contentReq, &ContentV2ArticlesRequest{}); err != nil {
		return err
	}

	err := contentReq.UnpackSession()
	if err != nil {
		return common.NewSessionError(err)
	}

	ok, err := buzzscreen.Service.DeviceUseCase.ValidateUnitDeviceToken(contentReq.Session.UserID)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	country := contentReq.GetCountry(ctx)

	channelsResponse := dto.ContentChannelsResponse{}
	var channelQuery *model.ContentChannelsQuery
	if contentReq.IDs != "" {
		channelQuery = model.NewContentChannelsQuery().WithIDs(contentReq.IDs)
	} else {
		channelQuery = model.NewContentChannelsQuery().WithCountryAndCategoryID(service.GetSupportedCountry(country), contentReq.CategoryID)
	}
	channelsResponse.Result.Channels = service.GetContentChannels(channelQuery)
	return c.JSON(http.StatusOK, getResponseSupport(c, channelsResponse))
}

type (
	// ContentV2BaseRequest type definition
	ContentV2BaseRequest struct {
		AdID       string        `form:"adId" query:"adId" validate:"required"`
		CountryReq string        `form:"country" query:"country"`
		Birthday   string        `form:"birthday" query:"birthday"`
		Locale     string        `form:"locale" query:"locale" validate:"required"`
		TimeZone   string        `form:"timeZone" query:"timeZone" validate:"required"`
		Request    *http.Request `form:"-"`
		SdkVersion int           `form:"sdkVersion" query:"sdkVersion" validate:"required"`
		SessionKey string        `form:"sessionKey" query:"sessionKey" validate:"required"`
		Os         string        `form:"os" query:"os"`
		OsVersion  int           `form:"osVersion" query:"osVersion" validate:"required"`
		UnitID     int64         `form:"unitId" query:"unitId"`
	}
	// ContentV2CategoriesRequest type definition
	ContentV2CategoriesRequest struct {
		ContentV2BaseRequest
	}
	// ContentV2ArticlesRequest type definition
	ContentV2ArticlesRequest struct {
		ContentV2BaseRequest
		IDs               string        `form:"ids" query:"ids"` //Bookmark 기능
		CategoryID        string        `form:"categoryId" query:"categoryId"`
		Categories        string        `form:"categories" query:"categories"`
		ChannelID         int64         `form:"channelId" query:"channelId"`
		EncryptedQueryKey string        `form:"queryKey" query:"queryKey"`
		CustomTarget1     string        `form:"customTarget1" query:"customTarget1"`
		CustomTarget2     string        `form:"customTarget2" query:"customTarget2"`
		CustomTarget3     string        `form:"customTarget3" query:"customTarget3"`
		FilterCategories  string        `form:"filterCategories" query:"filterCategories"`
		FilterChannelIDs  string        `form:"filterChannelIDs" query:"filterChannelIDs"`
		Gender            string        `form:"gender" query:"gender"`
		IsDebugging       bool          `form:"isDebugging" query:"isDebugging"`
		LandingTypes      string        `form:"landingTypes" query:"landingTypes"`
		Package           string        `form:"package" query:"package"`
		PlaceType         dto.PlaceType `form:"placeType" query:"placeType"`
		Size              int           `form:"size" query:"size"`
		TypesString       string        `form:"types" query:"types" validate:"required"` //eg. {"IMAGE":["INTERSTITIAL"]} or {"NATIVE":[]}
	}
	// GetContentV2ChannelsRequest type definition
	GetContentV2ChannelsRequest struct {
		ContentV2BaseRequest
		CategoryID string `form:"categoryId" query:"categoryId"`
		IDs        string `form:"ids" query:"ids"`
	}
	// GetDeviceV2ConfigRequest type definition
	GetDeviceV2ConfigRequest struct {
		ContentV2BaseRequest
		Method string
	}
	// PutDeviceV2ConfigRequest type definition
	PutDeviceV2ConfigRequest struct {
		ContentV2BaseRequest
		Method string
		Data   string `form:"data" query:"data" validate:"required"`
	}
)
