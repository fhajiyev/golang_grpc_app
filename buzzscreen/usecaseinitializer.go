package buzzscreen

import (
	"os"

	authsvc "github.com/Buzzvil/buzzapis/go/auth"
	pbprofile "github.com/Buzzvil/buzzapis/go/profile"
	pbreward "github.com/Buzzvil/buzzapis/go/reward"
	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/jwe"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	baUserRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/ad/bauserrepo"
	adRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/ad/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	appRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/app/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
	authRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/auth/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/config"
	configRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/config/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	contentCampaignRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview"
	customPreviewRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbapp"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbcontentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscontentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	deviceRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/device/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	eventRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/event/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/impressiondata"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/location"
	locationRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/location/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/log"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/notiplus"
	notiplusRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/notiplus/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/profilerequest"
	profileRequestRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/profilerequest/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/report"
	reportRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/report/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	rewardRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/reward/repo"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/session"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/userreferral"
	userReferralRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/userreferral/repo"

	"github.com/go-redis/redis"
	"github.com/go-resty/resty"
	"google.golang.org/grpc"
)

func (bs *Buzzscreen) initAdUseCase(redisCache *rediscache.RedisCache) ad.UseCase {
	baUserRepo := baUserRepo.New(bs.DB, redisCache)
	repo := adRepo.New(redisCache, bs.BuzzAdURL)
	return ad.NewUseCase(baUserRepo, repo, log.NewStructuredLogger(log.NewFmtWrapper()))
}

func (bs *Buzzscreen) initAppUseCase(redisCache *rediscache.RedisCache) app.UseCase {
	dbapp := dbapp.NewSource(bs.DB)
	return app.NewUseCase(appRepo.New(dbapp, redisCache))
}

func (bs *Buzzscreen) initAuthUseCase() auth.UseCase {
	authsvcConn := bs.connectServiceWithGRPC(env.Config.AuthsvcURL)
	authsvcClient := authsvc.NewAuthServiceClient(authsvcConn)
	ar := authRepo.New(authsvcClient)
	return auth.NewUseCase(ar)
}

func (bs *Buzzscreen) initConfigUseCase() config.UseCase {
	cr := configRepo.New()
	return config.NewUseCase(cr)
}

func (bs *Buzzscreen) initContentCampaignUseCase(redisCache *rediscache.RedisCache) contentcampaign.UseCase {
	client := redis.NewClient(&redis.Options{
		Addr:     env.Config.StatRedis.Endpoint,
		Password: "", // no password set
		DB:       env.Config.StatRedis.DB,
	})
	ccRedis := rediscontentcampaign.NewSource(client)
	ccDB := dbcontentcampaign.NewSource(bs.DB)
	ccr := contentCampaignRepo.New(ccDB, redisCache, ccRedis)
	return contentcampaign.NewUseCase(ccr)
}

func (bs *Buzzscreen) initDeviceUseCase() device.UseCase {
	dpTable := bs.DynamoDB.Table(env.Config.DynamoTableProfile)
	dpr := deviceRepo.NewProfileRepo(&dpTable)
	daTable := bs.DynamoDB.Table(env.Config.DynamoTableActivity)
	dar := deviceRepo.NewActivityRepo(&daTable)
	dr := deviceRepo.New(dbdevice.NewSource(bs.DB))
	return device.NewUseCase(dr, dpr, dar)
}

func (bs *Buzzscreen) initEventUseCase(redisCache *rediscache.RedisCache) event.UseCase {
	rewardsvcURL := bs.getEnv("REWARDSVC_URL")
	rewardsvcConn := bs.connectServiceWithGRPC(rewardsvcURL)
	rewardsvcClient := pbreward.NewRewardServiceClient(rewardsvcConn)
	eventPayloadSecretKey := bs.getEnv("EVENT_PAYLOAD_SECRET_KEY")
	manager, err := jwe.NewManager(eventPayloadSecretKey, event.TokenExpiration)
	if err != nil {
		core.Logger.Errorf("failed to create jwe.Manager for reward payload. error:%v", err)
		os.Exit(1)
	}

	repo := eventRepo.New(rewardsvcClient, bs.BuzzScreenAPIURL, redisCache)
	return event.NewUseCase(repo, manager, log.NewStructuredLogger(log.NewFmtWrapper()))
}

func (bs *Buzzscreen) initImpressionDataUseCase() impressiondata.UseCase {
	return impressiondata.NewUseCase()
}

func (bs *Buzzscreen) initLocationUseCase() location.UseCase {
	lr := locationRepo.New(bs.GeoDB)
	return location.NewUseCase(lr)
}

func (bs *Buzzscreen) initNotiPlusUseCase() notiplus.UseCase {
	return notiplus.NewUseCase(notiplusRepo.New(bs.DB))
}

func (bs *Buzzscreen) initPayloadUseCase() payload.UseCase {
	return payload.NewUseCase()
}

func (bs *Buzzscreen) initReportUseCase() report.UseCase {
	rr := reportRepo.New(bs.DB, bs.BuzzAdURL)
	return report.NewUseCase(rr)
}

func (bs *Buzzscreen) initRewardUseCase() reward.UseCase {
	pointTable := bs.DynamoDB.Table(env.Config.DynamoTablePoint)
	rr := rewardRepo.New(&pointTable)
	return reward.NewUseCase(rr)
}

func (bs *Buzzscreen) initSessionUseCase() session.UseCase {
	return session.NewUseCase()
}

func (bs *Buzzscreen) initTrackingDataUseCase() trackingdata.UseCase {
	return trackingdata.NewUseCase()
}

func (bs *Buzzscreen) initUserReferralUseCase() userreferral.UseCase {
	verifyDeviceClient := resty.New()
	verifyDeviceClient.SetProxy(env.Config.ProxyURL)
	ur := userReferralRepo.New(bs.DB, env.Config.BuzzconInternalURL, resty.New(), verifyDeviceClient)
	return userreferral.NewUseCase(ur)
}

func (bs *Buzzscreen) initCustomPreviewUseCase() custompreview.UseCase {
	cpr := customPreviewRepo.New(bs.DB)
	return custompreview.NewUseCase(cpr)
}

func (bs *Buzzscreen) initProfileRequestUseCase() profilerequest.UseCase {
	profilesvcURL, ok := os.LookupEnv("PROFILESVC_URL")
	var prr *profileRequestRepo.Repository
	if ok {
		conn := bs.connectServiceWithGRPC(profilesvcURL)
		grpcClient := pbprofile.NewProfileServiceClient(conn)
		prr = profileRequestRepo.New(grpcClient)
	} else {
		prr = profileRequestRepo.New(nil)
	}
	return profilerequest.NewUseCase(prr)
}

func (bs *Buzzscreen) connectServiceWithGRPC(url string) *grpc.ClientConn {
	// TODO check secure option
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		core.Logger.Errorf("did not connect with remote service. url: %s, error: %v", url, err)
		os.Exit(1)
	}

	core.Logger.Infof("connectServiceWithGRPC() - connected with url: %s", url)
	return conn
}

func (bs *Buzzscreen) getEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		core.Logger.Errorf("failed to get environment variable. key: %s", key)
		os.Exit(1)
	}

	return val
}
