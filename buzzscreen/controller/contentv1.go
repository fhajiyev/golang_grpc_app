package controller

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/recovery"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/impressiondata"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
)

// PostContentAllocationV1 func definition
func PostContentAllocationV1(c core.Context) error {
	ctx := c.Request().Context()
	// 요청을 파싱
	var allocReq dto.ContentAllocV1Request
	if err := bindValue(c, &allocReq); err != nil {
		return err
	}

	ok, err := buzzscreen.Service.DeviceUseCase.ValidateUnitDeviceToken(allocReq.UnitDeviceToken)
	if !ok {
		core.Logger.Warnf("PostContentAllocationV1() - failed to validate unit_device_token. err: %s. req: %+v", err.Error(), allocReq)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	profile := allocReq.GetDynamoProfile()
	unit := allocReq.GetUnit(ctx)
	if profile != nil {
		service.UpdateProfileUnitRegisterSeconds(profile, unit.ID, allocReq.GetAppID(ctx), c.Path())
		service.GiveWelcomeReward(ctx, profile, allocReq.UnitDeviceToken, unit.ID, allocReq.GetCountry(ctx))
	}

	content, err := getContent(ctx, &allocReq)
	if err != nil {
		return err
	}

	err = c.JSON(http.StatusOK, dto.ContentAllocV1Response{
		Code:      200,
		Message:   "ok",
		Campaigns: content,
	})

	country := allocReq.GetCountry(ctx)
	buzzscreen.Service.Metrics.AllocationRequests.WithLabelValues(country).Inc()
	buzzscreen.Service.Metrics.AllocatedContent.WithLabelValues(country).Observe(float64(len(content)))

	core.Logger.Infof("dto.PostContentAllocationV1() - Req: %v, ContentSize: %v", allocReq, len(content))

	return err
}

func getContent(ctx context.Context, allocReq *dto.ContentAllocV1Request) ([]*dto.CampaignV1, error) {
	if allocReq.GetUnit(ctx).ContentType == app.ContentTypeNone {
		return nil, nil
	}
	esContentCampaigns, err := service.GetV1ContentCampaignsFromES(ctx, allocReq)
	if err != nil {
		return nil, err
	}
	return getCampaignsFromESContentCampaigns(ctx, esContentCampaigns, allocReq), nil
}

func getCampaignsFromESContentCampaigns(ctx context.Context, esContentList []*dto.ESContentCampaign, allocReq *dto.ContentAllocV1Request) []*dto.CampaignV1 {
	//Convert
	campaigns := make([]*dto.CampaignV1, 0)

	for _, esContent := range esContentList {
		camp := parseToCampaign(ctx, esContent, allocReq)
		if camp == nil {
			continue
		}
		clickReq := dto.ClickRequest{
			ClickURL:        camp.ClickURL,
			ClickURLClean:   camp.ClickURLClean,
			DeviceID:        allocReq.DeviceID,
			ID:              camp.ID,
			IFA:             allocReq.IFA,
			Name:            camp.Name,
			OrganizationID:  camp.OrganizationID,
			OwnerID:         camp.OwnerID,
			Type:            camp.Type,
			Unit:            allocReq.GetUnit(ctx),
			UnitDeviceToken: allocReq.UnitDeviceToken,
		}

		if ma := allocReq.GetModelArtifact(ctx); ma != nil {
			td := &trackingdata.TrackingData{
				ModelArtifact: *ma,
			}
			clickReq.TrackingData = td
		}

		camp.ClickURL = clickReq.BuildClickRedirectURL()
		if len(camp.ClickURL) > 2048 {
			core.Logger.Infof("getCampaignsFromESContentCampaigns() - %v click_url is too long (%v)\n%v", camp.ID, len(camp.ClickURL), camp.ClickURL)
			continue
		}
		camp.SetPayloadWith(ctx, allocReq)
		campaigns = append(campaigns, camp)
	}
	return campaigns
}

//es_campaign_to_response 완료
func parseToCampaign(ctx context.Context, esContent *dto.ESContentCampaign, allocReq *dto.ContentAllocV1Request) *dto.CampaignV1 {
	defer recovery.LogRecoverWith(*esContent)
	var unitExtraData map[string]interface{}
	if esContent.JSON != "{}" {
		var esExtra dto.ESUnitExtra
		if err := json.Unmarshal([]byte(esContent.JSON), &esExtra); err == nil {
			unitExtraData = esExtra.Unit
		} else {
			core.Logger.WithError(err).Errorf("parseToCampaign() - esContent.json: %v", esContent.JSON)
		}
	}

	impReq := dto.ImpressionRequest{
		ImpressionData: impressiondata.ImpressionData{
			CampaignID:      esContent.ID,
			Country:         allocReq.GetCountry(ctx),
			DeviceID:        allocReq.DeviceID,
			IFA:             allocReq.IFA,
			UnitDeviceToken: allocReq.UnitDeviceToken,
			UnitID:          allocReq.GetUnit(ctx).ID,
		},
	}

	if ma := allocReq.GetModelArtifact(ctx); ma != nil {
		td := &trackingdata.TrackingData{
			ModelArtifact: *ma,
		}
		impReq.TrackingData = td
	}

	if allocReq.Gender != "" {
		impReq.ImpressionData.Gender = &(allocReq.Gender)
	}
	yob := allocReq.GetYearOfBirth()
	if yob > 0 {
		impReq.ImpressionData.YearOfBirth = &yob
	}

	newLandingType := esContent.LandingType
	switch newLandingType {
	case model.LandingTypeYoutube: //Youtube
		newLandingType = model.LandingTypeOverlay
	case model.LandingTypeCard:
		if !app.IsHoneyscreenOrSlidejoyAppID(allocReq.GetAppID(ctx)) {
			newLandingType = model.LandingTypeBrowser
		}
	}

	baseReward := allocReq.GetUnit(ctx).BaseReward
	// OCB contents base_reward 제거. BZZPB-38
	if allocReq.GetUnit(ctx).ID == 20867641205588 {
		baseReward = 0
	}

	campaign := &dto.CampaignV1{
		ActionReward:         0,
		BaseReward:           baseReward,
		Category:             esContent.Categories,
		ChannelID:            esContent.ChannelID,
		Channel:              esContent.Channel,
		CleanMode:            getIntDefault(esContent.CleanMode, model.CleanModeForceDisabled),
		ClickURLClean:        esContent.CleanLink,
		ClickURL:             esContent.ClickURL,
		DisplayType:          esContent.DisplayType,
		EndedAt:              utils.ConvertToUnixTime(esContent.EndDate),
		FirstDisplayWeight:   int(math.Min(float64(esContent.DisplayWeight*model.DisplayWeightMultiplier), 10000000)),
		FirstDisplayPriority: 10,
		ID:                   esContent.ID,
		Image:                esContent.GetCDNImageURL(allocReq.GetCreativeType(ctx), allocReq.SdkVersion),
		ImpressionURLs:       []string{impReq.BuildImpressionURL()},
		Ipu:                  getIntPtrDefault(esContent.Ipu, 9999),
		IsAd:                 false,
		IsMedia:              esContent.OrganizationID == allocReq.GetUnit(ctx).OrganizationID,
		LandingReward:        int(esContent.LandingReward),
		LandingType:          LandingTypeResponseMapping[newLandingType],
		Name:                 esContent.Name,
		OrganizationID:       esContent.OrganizationID,
		OwnerID:              esContent.OwnerID,
		ProviderID:           esContent.ProviderID,
		SourceURL:            esContent.ClickURL,
		StartedAt:            utils.ConvertToUnixTime(esContent.StartDate),
		SupportWebp:          true,
		Timezone:             esContent.Timezone,
		TargetApp:            esContent.TargetApp,
		Tipu:                 getIntDefault(esContent.Tipu, 0),
		Type:                 model.CampaignTypeCast,
		UnitPrice:            0,
		UnlockReward:         0,
	}

	if len(unitExtraData) > 0 {
		campaign.Extra = unitExtraData
	} else {
		campaign.Extra = make(map[string]interface{})
	}

	campaign.SetPayloadWith(ctx, allocReq)

	if esContent.WeekSlot != "" {
		startSlot := utils.GetWeekdayStartsFromMonday(allocReq.GetLocalTime(ctx))*24 + allocReq.GetLocalTime(ctx).Hour()
		weekSlots := make([]string, 0)

		for i := 0; i < 12; i++ {
			weekSlots = append(weekSlots, strconv.Itoa(startSlot+i))
		}

		weekSlots = utils.Intersection(weekSlots, strings.Split(esContent.WeekSlot, ","))
		hourSlots := make([]string, 0)

		for _, weekSlot := range weekSlots {
			weekSlotInt, err := strconv.Atoi(weekSlot)
			if err == nil {
				hourSlots = append(hourSlots, strconv.Itoa(weekSlotInt%24))
			}
		}
		campaign.Slot = strings.Join(hourSlots, ",")
	} else {
		campaign.Slot = "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23"
	}

	return campaign
}

// GetContentChannelsV1 func definition
func GetContentChannelsV1(c core.Context) error {
	var contentReq dto.ContentChannelsV1Request
	if err := bindValue(c, &contentReq); err != nil {
		return err
	}

	channelQuery := model.NewContentChannelsQuery().WithIDs(contentReq.IDs)
	channelsRes := dto.ContentChannelsV1Response{
		Status:   "success",
		Channels: make(map[string]*model.ContentChannel),
	}
	for _, channel := range *service.GetContentChannels(channelQuery) {
		channelsRes.Channels[strconv.FormatInt(channel.ID, 10)] = channel
	}
	return c.JSON(http.StatusOK, &channelsRes)
}

func getIntDefault(esValue int, defaultValue int) int {
	if esValue == 0 {
		return defaultValue
	}
	return esValue
}

func getIntPtrDefault(esValue *int, defaultValue int) int {
	if esValue == nil || *esValue == 0 {
		return defaultValue
	}
	return *esValue
}

// GetContentConfigV1 func definition
func GetContentConfigV1(c core.Context) error {
	var contentReq dto.GetDeviceConfigV1Request
	if err := bindValue(c, &contentReq); err != nil {
		return err
	}

	cc, err := getDeviceContentConfig(contentReq.DeviceID)
	if err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err.Error()}
	}

	return c.JSON(http.StatusOK, &dto.DeviceConfigV1Response{
		Status:   "success",
		Category: cc.Category,
		Campaign: cc.Campaign,
		Channel:  cc.Channel,
		DeviceID: cc.DeviceID,
	})
}

// PostContentConfigV1 func definition
func PostContentConfigV1(c core.Context) error {
	var contentReq dto.PostDeviceConfigV1Request
	if err := bindValue(c, &contentReq); err != nil {
		return err
	}

	cc, err := getDeviceContentConfig(contentReq.DeviceID)
	if err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err.Error()}
	}

	switch contentReq.Type {
	case "category":
		cc.Category = contentReq.Config
		cc.Save()
	case "channel":
		cc.Channel = contentReq.Config
		cc.Save()
	case "campaign":
		cc.Campaign = contentReq.Config
		cc.Save()
	default:
		return c.JSON(http.StatusBadRequest, &(map[string]interface{}{
			"error": "Unsupported config type",
		}))
	}
	return c.JSON(http.StatusOK, &(map[string]interface{}{
		"message": "success",
	}))
}
