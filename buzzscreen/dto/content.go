package dto

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/cypher"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/impressiondata"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/session"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
)

// CDN related const definition
const (
	ChangeCDNRootToAkamai = false
	CDNAkamaiRoot         = "https://buzzvil.akamaized.net/buzzscreen"
	CDNCloudFrontRoot     = "https://d3oh8k1ylpijyu.cloudfront.net"
)

// CreativeType type definition
type CreativeType string

// CreativeType constants
const (
	TypeImage  CreativeType = "IMAGE"
	TypeNative CreativeType = "NATIVE"
	TypeBanner CreativeType = "BANNER"
	TypeSDK    CreativeType = "SDK"
	TypeWeb    CreativeType = "WEB"
)

type (
	// ContentBaseRequest type definition
	ContentBaseRequest struct {
		AdID            string        `form:"ad_id" query:"ad_id" validate:"required"`
		CountryReq      string        `form:"country" query:"country"`
		BirthYear       int           `form:"birth_year" query:"birth_year"`
		Birthday        string        `form:"birthday" query:"birthday"`
		Locale          string        `form:"locale" query:"locale" validate:"required"`
		TimeZone        string        `form:"timezone" query:"timezone" validate:"required"`
		Request         *http.Request `form:"-"`
		SdkVersion      int           `form:"sdk_version" query:"sdk_version" validate:"required"`
		SessionKey      string        `form:"session_key" query:"session_key" validate:"required"`
		Os              string        `form:"os" query:"os"`
		OsVersion       int           `form:"os_version" query:"os_version"`
		UnitID          int64         `form:"unit_id" query:"unit_id"`
		IsInBatteryOpts bool          `form:"is_in_battery_optimizations" query:"is_in_battery_optimizations"`
		IFV             *string       `form:"ifv" query:"ifv"`

		age       *int // 0: unknown
		yob       *int // 0: unknown
		country   string
		language  string
		localTime *time.Time
		Session   session.Session
		unit      *app.Unit
	}

	// ContentCategoriesRequest type definition
	ContentCategoriesRequest struct {
		ContentBaseRequest
	}

	// ContentCategoriesResponse type definition
	ContentCategoriesResponse struct {
		ContentCategories []*model.ContentCategory `json:"categories"`
	}

	// ESContentCampaign type definition
	ESContentCampaign struct {
		model.ContentCampaign
		Channel       *ESContentChannel   `json:"channel"`
		Clicks        int64               `json:"clicks"`
		CreativeLinks map[string][]string `json:"creative_links"`
		ImageRatio    *float64            `json:"image_ratio"`
		Impressions   int64               `json:"impressions"`
		Provider      *ESContentProvider  `json:"provider"`
		Related       *int64              `json:"related"`
		RelatedCount  *int                `json:"related_count"`
		Score         *float64            `json:"score"`
		ScoreFactors  map[string]float64  `json:"score_factors"`
		ModelArtifact string              `json:"model_artifact"`

		CreativeTypes string `json:"creative_types"`

		DetargetApp   string `gorm:"type:varchar(100)" json:"detarget_app" esNull:"__GLOB__"` //get("detarget_app")
		DetargetUnit  string `json:"detarget_unit" esNull:"__GLOB__"`
		DetargetAppID string `json:"detarget_app_id" esNull:"__GLOB__"`
		DetargetOrg   string `json:"detarget_org" esNull:"__GLOB__"`
	}

	// ESContentProvider type definition
	ESContentProvider struct {
		ID    int64 `json:"id"`
		Score int   `json:"score"`
	}

	// ESContentChannel type definition
	ESContentChannel struct {
		ID   int64  `json:"id"`
		Logo string `json:"logo"`
		Name string `json:"name"`
	}

	// ESUnitExtra type definition
	ESUnitExtra struct {
		Unit map[string]interface{} `json:"unit"`
	}

	// Common type definition
	Common struct {
		ID       int64                  `json:"id"`
		Creative map[string]interface{} `json:"creative"`

		CallToAction       *string                `json:"call_to_action,omitempty"`
		ClickTrackers      []string               `json:"click_trackers,omitempty"`
		FailTrackers       []string               `json:"fail_trackers,omitempty"`
		ImpressionTrackers []string               `json:"impression_trackers,omitempty"`
		RevenueType        *string                `json:"revenue_type,omitempty"`
		Payload            *string                `json:"payload,omitempty"`
		Events             Events                 `json:"events"`
		Extra              map[string]interface{} `json:"extra,omitempty"`
		//Deprecated
		LandingType *int `json:"landing_type,omitempty"` //Deprecated - creative 안쪽으로 이동 //DEFAULT_BROWSER = 1 (외부 브라우저), PLAYER = 2 (동영상 상품용 - 버즈스크린 LandingOverlayActivity), IN_APP = 3 (기존 slidejoy BrowserActivity)
	}

	// Events type definition
	Events []Event

	// Event type definition
	Event struct {
		Type         string            `json:"event_type"`
		TrackingURLs []string          `json:"tracking_urls"`
		Reward       *Reward           `json:"reward,omitempty"`
		Extra        map[string]string `json:"extra,omitempty"`
	}

	// Reward type definition
	Reward struct {
		Amount         int64             `json:"amount"`
		Status         string            `json:"status"`
		IssueMethod    string            `json:"issue_method"`
		StatusCheckURL string            `json:"status_check_url"`
		TTL            int64             `json:"ttl"`
		Extra          map[string]string `json:"extra,omitempty"`
	}

	// ContentArticle type definition
	ContentArticle struct {
		Common
		CleanMode     int                    `json:"clean_mode"`
		Category      *model.ContentCategory `json:"category"`
		Channel       *model.ContentChannel  `json:"channel"`
		CreatedAt     int64                  `json:"created_at"`
		Name          string                 `json:"name"`
		SourceURL     string                 `json:"source_url"`
		ScoringParams *[]float64             `json:"scoring_params,omitempty"`
	}

	// ContentImage type definition
	ContentImage struct {
		ImageURL string `json:"image_url"`
		Width    *int   `json:"width,omitempty"`
		Height   *int   `json:"height,omitempty"`
	}

	// ContentArticles type definition
	ContentArticles []*ContentArticle
)

// String func definition
func (cas ContentArticles) String() string {
	ids := make([]int64, 0)
	for _, ca := range cas {
		ids = append(ids, ca.ID)
	}
	return fmt.Sprintf("%v", ids)
}

// GetAge func definition
func (contentReq *ContentBaseRequest) GetAge() int {
	if contentReq.age == nil {
		age := 0
		if contentReq.Birthday != "" {
			age = utils.GetAge(contentReq.Birthday)
		} else if contentReq.BirthYear > 0 {
			age = time.Now().Year() - contentReq.BirthYear - 1
		}
		contentReq.age = &age
	}
	return *contentReq.age
}

// GetYearOfBirth func definition
func (contentReq *ContentBaseRequest) GetYearOfBirth() *int {
	if contentReq.yob == nil {
		yob := 0
		if contentReq.Birthday != "" {
			birthday, _ := time.Parse("2006-01-02", contentReq.Birthday)
			if year := birthday.Year(); birthday.IsZero() == false {
				yob = year
			}
		}
		contentReq.yob = &yob
	}
	if *contentReq.yob > 0 {
		return contentReq.yob
	}
	return nil
}

// GetLanguage func definition
func (contentReq *ContentBaseRequest) GetLanguage() string {
	if contentReq.language == "" {
		contentReq.language, _ = utils.SplitLocale(contentReq.Locale)
	}
	return contentReq.language
}

// GetCountry func definition
func (contentReq *ContentBaseRequest) GetCountry(ctx context.Context) string {
	if contentReq.country != "" {
		return contentReq.country
	}

	if contentReq.CountryReq != "" {
		contentReq.country = contentReq.CountryReq
		return contentReq.country
	}
	unit := contentReq.GetUnit(ctx)

	if unit != nil && unit.Country != "" {
		contentReq.country = unit.Country
		return contentReq.country
	}

	_, contentReq.CountryReq = utils.SplitLocale(contentReq.Locale)
	if contentReq.Request != nil {
		loc := buzzscreen.Service.LocationUseCase.GetClientLocation(contentReq.Request, contentReq.CountryReq)
		contentReq.country = loc.Country
		return contentReq.country
	}

	contentReq.country = contentReq.CountryReq
	return contentReq.country
}

// UnpackSession unpacks session key and assign to Session
func (contentReq *ContentBaseRequest) UnpackSession() error {
	session, err := buzzscreen.Service.SessionUseCase.GetSessionFromKey(contentReq.SessionKey)
	if err != nil {
		return err
	}

	contentReq.Session = *session
	return nil
}

// GetUnit func definition
func (contentReq *ContentBaseRequest) GetUnit(ctx context.Context) *app.Unit {
	if contentReq.unit == nil {
		appUseCase := buzzscreen.Service.AppUseCase
		if contentReq.UnitID > 0 {
			contentReq.unit, _ = appUseCase.GetUnitByID(ctx, contentReq.UnitID)
		} else if contentReq.Session.AppID > 0 {
			contentReq.unit, _ = appUseCase.GetUnitByAppID(ctx, contentReq.Session.AppID)
		}
	}
	return contentReq.unit
}

// GetLocalTime func definition
func (contentReq *ContentBaseRequest) GetLocalTime(ctx context.Context) *time.Time {
	if contentReq.localTime == nil {
		if u := contentReq.GetUnit(ctx); u != nil {
			loc, err := time.LoadLocation(u.Timezone)
			localTime := time.Now()
			if err == nil {
				localTime = localTime.In(loc)
			}
			contentReq.localTime = &localTime
		}
	}
	return contentReq.localTime
}

// GetCDNImageURL func definition
func (cc *ESContentCampaign) GetCDNImageURL(creativeType string, sdkVersion int) string {
	imageURL := cc.CreativeLinks[creativeType][0]
	//noinspection ALL
	if ChangeCDNRootToAkamai && strings.Contains(imageURL, CDNCloudFrontRoot) {
		return strings.Replace(imageURL, CDNCloudFrontRoot, CDNAkamaiRoot, 1)
	}

	// If creative image name ends with `_720_1560`, it has size 720 X 1560 and has a corresponding image url with size 720 X 1230,
	// which is formed by removing `_720_1560` at the end of the file name.
	// If SDK version is lower than 3900 use 720 X 1230 image url.
	if sdkVersion < 3900 && creativeType == "A" {
		if strings.HasSuffix(imageURL, "_720_1560.jpg") {
			imageURL = strings.Replace(imageURL, "_720_1560.jpg", "_720_1230.jpg", 1)
		}

	}
	return imageURL
}

// GetDocToCreate func definition
func (cc ESContentCampaign) GetDocToCreate() ESContentCampaign {
	valueRef := reflect.ValueOf(&cc)
	setESNullValue(&valueRef)

	valueRef = reflect.ValueOf(&(cc.ContentCampaign))
	setESNullValue(&valueRef)

	return cc
}

// EscapeNull func definition
func (cc ESContentCampaign) EscapeNull() (ESContentCampaign, error) {
	valueRef := reflect.ValueOf(&cc)
	err := removeESNullValue(&valueRef)
	if err == nil {
		valueRef = reflect.ValueOf(&(cc.ContentCampaign))
		err = removeESNullValue(&valueRef)
	}

	return cc, err
}

func removeESNullValue(valueRef *reflect.Value) error {
	for i := 0; i < valueRef.Elem().NumField(); i++ {
		valueField := valueRef.Elem().Field(i)
		value := valueField
		tag := valueRef.Elem().Type().Field(i).Tag
		if esNullValue, ok := tag.Lookup("esNull"); ok && value.Interface() == esNullValue {
			switch valueField.Kind() {
			case reflect.String:
				valueField.SetString("")
			default:
				return fmt.Errorf("not supported kind - %v. You should add code below", valueField.Kind())
			}
		}
	}
	return nil
}

func setESNullValue(valueRef *reflect.Value) error {
	for i := 0; i < valueRef.Elem().NumField(); i++ {
		valueField := valueRef.Elem().Field(i)
		value := valueField
		tag := valueRef.Elem().Type().Field(i).Tag
		if esNullValue, ok := tag.Lookup("esNull"); ok && value.Interface() == "" {
			switch valueField.Kind() {
			case reflect.String:
				valueField.SetString(esNullValue)
			default:
				return fmt.Errorf("not supported kind - %v. You should add code below", valueField.Kind())
			}
		}
	}
	return nil
}

type (
	// ContentArticlesRequest type definition
	ContentArticlesRequest struct {
		ContentBaseRequest
		IDs               string `form:"ids" query:"ids"` //Bookmark 기능
		CategoryID        string `form:"category_id" query:"category_id"`
		Categories        string `form:"categories" query:"categories"`
		ChannelID         int64  `form:"channel_id" query:"channel_id"`
		EncryptedQueryKey string `form:"query_key" query:"query_key"`
		CustomTarget1     string `form:"custom_target_1" query:"custom_target_1"`
		CustomTarget2     string `form:"custom_target_2" query:"custom_target_2"`
		CustomTarget3     string `form:"custom_target_3" query:"custom_target_3"`
		FilterCategories  string `form:"filter_categories" query:"filter_categories"`
		FilterChannelIDs  string `form:"filter_channel_ids" query:"filter_channel_ids"`
		Gender            string `form:"gender" query:"gender"`
		IsDebugging       bool   `form:"is_debugging" query:"is_debugging"`
		LandingTypes      string `form:"landing_types" query:"landing_types"`
		Package           string `form:"package" query:"package"`
		//Deprecated
		PlaceType   PlaceType `form:"place_type" query:"place_type"`
		Size        int       `form:"size" query:"size"`
		TypesString string    `form:"types" query:"types" validate:"required"` //eg. {"IMAGE":["INTERSTITIAL"]} or {"NATIVE":[]}

		queryKey       *ContentQueryKey
		types          CreativeTypes
		deviceProfile  *device.Profile
		deviceActivity *device.Activity
		target         *ContentTarget
		campaignScores *map[int]int

		modelArtifact          *string
		isDebugScore           bool
		deviceCategoriesScores *map[string]float64
		deviceEntityScores     *map[string]float64
	}

	// CreativeTypes type definition
	CreativeTypes map[CreativeType][]string

	// ContentArticlesResponse type definition
	ContentArticlesResponse struct {
		ContentArticles ContentArticles `json:"articles"`
		QueryKey        *string         `json:"query_key,omitempty"`
	}

	// ContentQueryKey type definition
	ContentQueryKey struct {
		CreatedAt          int64   `json:"c"`
		Index              int     `json:"i"`
		ScoredCampaignsKey *int64  `json:"sk,omitempty"`
		CampaignIDs        []int64 `json:"cid"`
	}
)

// Marshal func definition
func (qk *ContentQueryKey) Marshal() string {
	bytes, err := json.Marshal(qk)
	if err != nil {
		panic(err)
	}
	encrypted, err := cypher.EncryptAesWithBase64([]byte(model.APIAesKey), []byte(model.APIAesIv), bytes, true)
	if err != nil {
		panic(err)
	}
	return encrypted
}

// GetCategoriesScores func definition
func (car *ContentArticlesRequest) GetCategoriesScores() *map[string]float64 {
	if car.deviceCategoriesScores == nil {
		dp := car.GetDynamoProfile()
		if dp == nil || dp.CategoriesScores == nil {
			return nil
		}
		scores := dp.CategoriesScores
		car.deviceCategoriesScores = scores
	}
	return car.deviceCategoriesScores
}

// GetEntityScores func definition
func (car *ContentArticlesRequest) GetEntityScores() *map[string]float64 {
	if car.deviceEntityScores == nil {
		dp := car.GetDynamoProfile()
		if dp == nil || dp.EntityScores == nil {
			return nil
		}
		scores := dp.EntityScores
		car.deviceEntityScores = scores
	}
	return car.deviceEntityScores
}

// GetIsDebugScore func definition
func (car *ContentArticlesRequest) GetIsDebugScore() bool {
	dp := car.GetDynamoProfile()
	if dp == nil {
		return false
	}
	car.isDebugScore = dp.IsDebugScore
	return car.isDebugScore
}

// GetModelArtifact func definition
func (car *ContentArticlesRequest) GetModelArtifact(ctx context.Context) *string {
	if car.modelArtifact == nil {
		defaultModelArtifact := "v4_a"
		if car.GetUnit(ctx).OrganizationID == 148 { // Exclude ocb units
			defaultModelArtifact = "v3"
		}
		dp := car.GetDynamoProfile()
		if dp == nil || dp.ModelArtifact == nil {
			car.modelArtifact = &defaultModelArtifact
		} else {
			car.modelArtifact = dp.ModelArtifact
		}
	}
	return car.modelArtifact
}

// GetTarget func definition
func (car *ContentArticlesRequest) GetTarget(ctx context.Context) *ContentTarget {
	if car.target == nil {
		car.target = &ContentTarget{
			Age:     car.GetAge(),
			Country: car.GetCountry(ctx),
			Gender:  car.Gender,
		}
	}
	return car.target
}

// GetQueryKey func definition
func (car *ContentArticlesRequest) GetQueryKey() (*ContentQueryKey, error) {
	if car.queryKey == nil && car.EncryptedQueryKey != "" {
		car.queryKey = &ContentQueryKey{}
		decrypted, err := cypher.DecryptAesWithBase64([]byte(model.APIAesKey), []byte(model.APIAesIv), car.EncryptedQueryKey, true)
		if err != nil {
			return nil, common.NewQueryKeyError(err)
		}
		core.Logger.Debugf("ContentArticlesRequest.GetQueryKey() - decrypted: %s", decrypted)
		err = json.Unmarshal(decrypted, car.queryKey)

		if err != nil {
			return nil, common.NewQueryKeyError(err)
		}
	}
	return car.queryKey, nil
}

// GetDynamoProfile func definition
func (car *ContentArticlesRequest) GetDynamoProfile() *device.Profile {
	if car.deviceProfile == nil {
		var err error
		car.deviceProfile, err = buzzscreen.Service.DeviceUseCase.GetProfile(car.Session.DeviceID)
		if err != nil {
			core.Logger.Errorf("ContentAllocRequest - Device %d GetProfile error %v", car.Session.DeviceID, err)
		}
	}
	return car.deviceProfile
}

// GetDynamoActivity func definition
func (car *ContentArticlesRequest) GetDynamoActivity() *device.Activity {
	if car.deviceActivity == nil {
		car.deviceActivity, _ = buzzscreen.Service.DeviceUseCase.GetActivity(car.Session.DeviceID)
	}
	return car.deviceActivity
}

// GetTypes func definition
func (car *ContentArticlesRequest) GetTypes() *CreativeTypes {
	if car.types == nil && car.TypesString != "" {
		err := json.Unmarshal([]byte(car.TypesString), &car.types)
		if err != nil {
			panic(err)
		}
	}
	return &(car.types)
}

// SetImpClickPayload func definition
func (article *ContentArticle) SetImpClickPayload(ctx context.Context, campaign *model.ContentCampaign, contentReq *ContentArticlesRequest) {
	article.setImpressionURL(ctx, contentReq)
	article.setClickURL(ctx, contentReq, campaign)
	article.setPayload(ctx, contentReq, campaign)
}

func (article *ContentArticle) setImpressionURL(ctx context.Context, contentReq *ContentArticlesRequest) {
	impReq := ImpressionRequest{
		ImpressionData: impressiondata.ImpressionData{
			UnitID:          contentReq.GetUnit(ctx).ID,
			CampaignID:      article.ID,
			IFA:             contentReq.AdID,
			DeviceID:        contentReq.Session.DeviceID,
			UnitDeviceToken: contentReq.Session.UserID,
			Country:         contentReq.GetCountry(ctx),
			YearOfBirth:     contentReq.GetYearOfBirth(),
		},
	}
	if contentReq.Gender != "" {
		impReq.ImpressionData.Gender = &(contentReq.Gender)
	}

	article.ImpressionTrackers = []string{
		impReq.BuildImpressionURL(),
	}
}

func (article *ContentArticle) setClickURL(ctx context.Context, contentReq *ContentArticlesRequest, content *model.ContentCampaign) {
	clickReq := ClickRequest{
		ClickURL:        content.ClickURL,
		ClickURLClean:   content.CleanLink,
		DeviceID:        contentReq.Session.DeviceID,
		ID:              content.ID,
		IFA:             contentReq.AdID,
		Name:            content.Name,
		OrganizationID:  content.OrganizationID,
		OwnerID:         content.OwnerID,
		Type:            model.CampaignTypeCast,
		Unit:            contentReq.GetUnit(ctx),
		UnitDeviceToken: contentReq.Session.UserID,
	}

	article.Creative["click_url"] = clickReq.BuildClickRedirectURL()
}

func (article *ContentArticle) setPayload(ctx context.Context, contentReq *ContentArticlesRequest, content *model.ContentCampaign) {
	p := &payload.Payload{
		Country:     contentReq.GetCountry(ctx),
		EndedAt:     utils.ConvertToUnixTime(content.EndDate),
		OrgID:       content.OrganizationID,
		Time:        time.Now().Unix(),
		Timezone:    content.Timezone,
		YearOfBirth: contentReq.GetYearOfBirth(),
	}

	if contentReq.Gender != "" {
		p.Gender = &contentReq.Gender
	}

	p.SetUnitID(contentReq.UnitID, contentReq.SdkVersion)

	payloadString := buzzscreen.Service.PayloadUseCase.BuildPayloadString(p)
	article.Payload = &payloadString
}

type (
	// GetContentChannelsRequest type definition
	GetContentChannelsRequest struct {
		ContentBaseRequest
		CategoryID string `form:"category_id" query:"category_id"`
		IDs        string `form:"ids" query:"ids"`
	}

	// ContentChannelsResponse type definition
	ContentChannelsResponse struct {
		Code    int     `json:"code"`
		Message *string `json:"message,omitempty"`
		Result  struct {
			Channels *model.ContentChannels `json:"channels"`
		} `json:"result"`
	}
)

type (
	// GetDeviceConfigRequest type definition
	GetDeviceConfigRequest struct {
		ContentBaseRequest
		Method string
	}

	// PutDeviceConfigRequest type definition
	PutDeviceConfigRequest struct {
		ContentBaseRequest
		Method string
		Data   string `form:"data" query:"data" validate:"required"`
	}
)

// ImpressionRequest struct definition
type ImpressionRequest struct {
	ImpressionData impressiondata.ImpressionData
	TrackingData   *trackingdata.TrackingData
}

// BuildImpressionURL func definition
func (impReq *ImpressionRequest) BuildImpressionURL() string {
	impressionDataUseCase := buzzscreen.Service.ImpressionDataUseCase
	impressionDataString := impressionDataUseCase.BuildImpressionDataString(impReq.ImpressionData)

	parameters := url.Values{
		"place":      {"__place__"},
		"position":   {"__position__"},
		"session_id": {"__session_id__"},
		"data":       {impressionDataString},
	}

	if impReq.TrackingData != nil {
		trackingDataUseCase := buzzscreen.Service.TrackingDataUseCase
		parameters.Add("tracking_data", trackingDataUseCase.BuildTrackingDataString(impReq.TrackingData))
	}

	queryString := parameters.Encode()
	return fmt.Sprint(buzzscreen.Service.BuzzScreenAPIURL, "/api/content_impression/?", queryString)
}

// ClickRequest struct definition
type ClickRequest struct {
	ClickURL        string
	ClickURLClean   string
	DeviceID        int64
	ID              int64
	IFA             string
	Name            string
	OrganizationID  int64
	OwnerID         int64
	Type            string
	Unit            *app.Unit
	UnitDeviceToken string
	TrackingData    *trackingdata.TrackingData

	TrackingURL  *string
	UseRewardAPI bool
}

// BuildClickRedirectURL func definition
func (clickReq *ClickRequest) BuildClickRedirectURL() string {
	if clickReq.ClickURL == "" {
		return ""
	}

	parameters := url.Values{
		"app_id":                   {strconv.FormatInt(clickReq.Unit.AppID, 10)},
		"base_reward":              {"__base_reward__"},
		"campaign_id":              {strconv.FormatInt(clickReq.ID, 10)},
		"campaign_type":            {clickReq.Type},
		"campaign_name":            {clickReq.Name},
		"campaign_payload":         {"__campaign_payload__"},
		"check":                    {"__check__"},
		"client_unit_device_token": {"__unit_device_token__"}, // 클라이언트에서 실시간으로 생성하는 unit_device_token을 받기위해 추가함
		"device_id":                {strconv.FormatInt(clickReq.DeviceID, 10)},
		"ifa":                      {clickReq.IFA},
		"unit_device_token":        {clickReq.UnitDeviceToken},
		"unit_id":                  {strconv.FormatInt(clickReq.Unit.ID, 10)},
		"place":                    {"__place__"},
		"position":                 {"__position__"},
		"redirect_url":             {strings.Replace(clickReq.ClickURL, "{ifa}", clickReq.IFA, -1)},
		"redirect_url_clean":       {clickReq.ClickURLClean},
		"reward":                   {"__reward__"},
		"session_id":               {"__session_id__"},
		"slot":                     {"__slot__"},
		"use_clean_mode":           {"__use_clean_mode__"},
	}

	if clickReq.TrackingURL != nil {
		parameters.Add("tracking_url", *clickReq.TrackingURL)
		parameters.Add("use_reward_api", strconv.FormatBool(clickReq.UseRewardAPI))
	}

	if clickReq.TrackingData != nil {
		trackingDataUseCase := buzzscreen.Service.TrackingDataUseCase
		parameters.Add("tracking_data", trackingDataUseCase.BuildTrackingDataString(clickReq.TrackingData))
	}

	if clickReq.OwnerID != 0 {
		parameters.Add("campaign_owner_id", strconv.FormatInt(clickReq.OwnerID, 10))
	} else {
		parameters.Add("campaign_owner_id", "")
	}

	if clickReq.OrganizationID == clickReq.Unit.OrganizationID {
		parameters.Add("campaign_is_media", strconv.Itoa(1))
	} else {
		parameters.Add("campaign_is_media", strconv.Itoa(0))
	}

	queryString := parameters.Encode()
	return fmt.Sprint(buzzscreen.Service.BuzzScreenAPIURL, "/api/click_redirect/?", queryString)
}

// SetUseRewardAPI set UseRewardAPI field
func (clickReq *ClickRequest) SetUseRewardAPI(sdkVersion int, canHandleEventsInfo bool, minimumStayDuration int) {
	if 10000 <= sdkVersion { // benefit
		if canHandleEventsInfo { // client call track event url directly
			clickReq.UseRewardAPI = false
		} else { // client requests to give reward via rewards API
			clickReq.UseRewardAPI = true
		}
	} else if 1800 <= sdkVersion && sdkVersion < 10000 { // lockscreen supporting landing reward (landing_page_duration)
		if minimumStayDuration > 0 { // client requests to give reward via rewards API after staying 'landing_page_duration' in the page
			clickReq.UseRewardAPI = true
		} else { // client requests to give reward via rewards API after 1s
			clickReq.UseRewardAPI = false
		}
	} else { // old lockscreen that does not support landing reward
		clickReq.UseRewardAPI = false
	}
}
