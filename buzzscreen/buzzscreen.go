package buzzscreen

import (
	"net/http"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/mq"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/activitysvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/appsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/clickredirectsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/configsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/contentimpressionsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/custompreviewsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/eventsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/installedappsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/monitorsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/notiplussvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/policysvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/reportsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/rewardsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/unlocksvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/userreferralsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/impressiondata"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/location"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/session"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
	"github.com/go-redis/redis"
	"github.com/guregu/dynamo"
	"github.com/jinzhu/gorm"
	"github.com/oschwald/geoip2-golang"
	echopprof "github.com/sevennt/echo-pprof"
	"gopkg.in/olivere/elastic.v5"
)

// Buzzscreen microservice
type Buzzscreen struct {
	DB       *gorm.DB
	DynamoDB *dynamo.DB
	ES       *elastic.Client
	GeoDB    *geoip2.Reader
	Metrics  *env.Metrics
	Redis    *redis.Client

	BuzzAdURL        string
	BuzzScreenAPIURL string

	// Deprecated - DDD 적용 이후 삭제 해야함
	AdUseCase             ad.UseCase
	AppUseCase            app.UseCase
	AuthUseCase           auth.UseCase
	DeviceUseCase         device.UseCase
	EventUseCase          event.UseCase
	ImpressionDataUseCase impressiondata.UseCase
	LocationUseCase       location.UseCase
	PayloadUseCase        payload.UseCase
	RewardUseCase         reward.UseCase
	SessionUseCase        session.UseCase
	TrackingDataUseCase   trackingdata.UseCase
}

// Service is buzzscreen service instance
var Service *Buzzscreen

// Init will do service initialization
func (bs *Buzzscreen) Init() (err error) {
	env.LoadServerConfig()
	if err := bs.waitForConn(); err != nil { // https://buzzvil.atlassian.net/browse/BS-2802
		return err
	}

	bs.DB, err = env.GetDatabase()
	if err != nil {
		return err
	}

	bs.DynamoDB = env.GetDynamoDB()
	bs.ES = env.GetElasticsearch()
	bs.Metrics = env.NewMetrics()
	bs.Redis = env.InitRedis()

	if !(env.IsLocal() || env.IsTest()) {
		bs.GeoDB, err = geoip2.Open("/go/GeoLite2-Country.mmdb") // TODO use relative path
	}
	return
}

func (bs *Buzzscreen) waitForConn() (err error) {
	timeLimit := time.Now().Add(time.Minute)
	for {
		_, err = http.Get(env.Config.ElasticSearch.Host + "/_cat/health")
		if err == nil || time.Now().After(timeLimit) {
			break
		}
		time.Sleep(time.Second * 2)
		core.Logger.Warnf("Buzzscreen.waitForConn() - retry... host: %s", env.Config.ElasticSearch.Host)
	}
	return
}

// Clean will do cleaning task before server exits
func (bs *Buzzscreen) Clean() (err error) {
	if bs.DB != nil {
		bs.DB.Close()
	}
	if bs.GeoDB != nil {
		bs.GeoDB.Close()
	}
	return
}

// Health will check the health of the service.
func (bs *Buzzscreen) Health() (err error) {
	return // There is another route for checking health - /live, /ready
}

// RegisterRoute adds a routing to the driver
func (bs *Buzzscreen) RegisterRoute(driver *core.Engine) {
	echopprof.Wrap(driver)

	bs.BuzzAdURL = bs.getEnv("BUZZAD_URL")
	bs.BuzzScreenAPIURL = bs.getEnv("BUZZSCREEN_API_URL")
	amqpURL := bs.getEnv("AMQP_URL")
	publisher := mq.CreatePublisher(amqpURL, core.Logger)

	redisCache := rediscache.NewSource(bs.Redis)

	adUC := bs.initAdUseCase(redisCache)
	appUC := bs.initAppUseCase(redisCache)
	authUC := bs.initAuthUseCase()
	configUC := bs.initConfigUseCase()
	contentCampaignUC := bs.initContentCampaignUseCase(redisCache)
	deviceUC := bs.initDeviceUseCase()
	eventUC := bs.initEventUseCase(redisCache)
	impressionDataUC := bs.initImpressionDataUseCase()
	locationUC := bs.initLocationUseCase()
	notiplusUC := bs.initNotiPlusUseCase()
	payloadUC := bs.initPayloadUseCase()
	reportUC := bs.initReportUseCase()
	rewardUC := bs.initRewardUseCase()
	sessionUC := bs.initSessionUseCase()
	trackingDataUC := bs.initTrackingDataUseCase()
	userReferralUC := bs.initUserReferralUseCase()
	customPreviewUC := bs.initCustomPreviewUseCase()
	profileRequestUC := bs.initProfileRequestUseCase()

	activitysvc.NewController(driver, deviceUC, locationUC)
	notiplussvc.NewController(driver, notiplusUC)
	appsvc.NewController(driver, appUC)
	clickredirectsvc.NewController(driver, rewardUC, appUC, contentCampaignUC, payloadUC, trackingDataUC, deviceUC, eventUC, profileRequestUC, bs.BuzzAdURL)
	configsvc.NewController(driver, configUC)
	contentimpressionsvc.NewController(driver, trackingDataUC, impressionDataUC, contentCampaignUC, deviceUC)
	eventsvc.NewController(driver, appUC, authUC, deviceUC, eventUC, contentCampaignUC, adUC, publisher)
	installedappsvc.NewController(driver, deviceUC, bs.BuzzAdURL)
	monitorsvc.NewController(driver)
	policysvc.NewController(driver, appUC, locationUC)
	reportsvc.NewController(driver, reportUC, appUC)
	rewardsvc.NewController(driver, appUC, eventUC, rewardUC)
	unlocksvc.NewController(driver, rewardUC, appUC, payloadUC)
	userreferralsvc.NewController(driver, userReferralUC, deviceUC, appUC)
	custompreviewsvc.NewController(driver, customPreviewUC, nil)

	bs.AdUseCase = adUC
	bs.AppUseCase = appUC
	bs.AuthUseCase = authUC
	bs.DeviceUseCase = deviceUC
	bs.EventUseCase = eventUC
	bs.ImpressionDataUseCase = impressionDataUC
	bs.LocationUseCase = locationUC
	bs.PayloadUseCase = payloadUC
	bs.RewardUseCase = rewardUC
	bs.SessionUseCase = sessionUC
	bs.TrackingDataUseCase = trackingDataUC
}

// New returns a buzzscreen service instance
func New() core.Service {
	Service = &Buzzscreen{}
	return core.Service(Service)
}
