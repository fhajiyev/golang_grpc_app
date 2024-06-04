package route

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/controller"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/authmw"
)

// InitLegacyRoute registers legacy routes.
func InitLegacyRoute(router *core.Engine) {

	v1 := router.Group("/api")
	{
		registerAPIv1Routes(v1)
		controller.NewDeviceController(v1)
	}

	// Version 2
	v2 := router.Group("/api/v2")
	{
		controller.NewDeviceController(v2)
		registerContentRoutes(v2)
	}

	v3 := router.Group("/api/v3")
	{
		controller.NewDeviceController(v3)
		registerAdsRoutes(v3)
		registerContentRoutes(v3)
		registerNotificationsRoutes(v3)
	}
}

func registerAPIv1Routes(router *core.RouterGroup) {
	router.POST("/init/", controller.InitDeviceV1)
	router.POST("/init_sdk/", controller.InitSdkV1)
	router.POST("/allocation/", controller.PostAllocationV1)
	router.POST("/content_allocation/", controller.PostContentAllocationV1)

	router.GET("/content/channels/", controller.GetContentChannelsV1)
	router.GET("/content/config/", controller.GetContentConfigV1)
	router.POST("/content/config/", controller.PostContentConfigV1)

	router.POST("/reward_mock/", logReq)

	// Need to be separed as a file like notification - Howard
	router.GET("/shopping/categories/", controller.GetShoppingCategories)
}

func registerAdsRoutes(router *core.RouterGroup) {
	router.GET("/ads", controller.GetAds, authmw.AppendAuthToHeader(buzzscreen.Service.AuthUseCase))
}

func registerContentRoutes(router *core.RouterGroup) {
	contentRouter := router.Group("/content")
	{
		contentRouter.GET("/categories", controller.GetContentCategories)
		contentRouter.GET("/articles", controller.GetContentArticles)
		contentRouter.GET("/channels", controller.GetContentChannels)
		contentRouter.GET("/config/:method", controller.GetDeviceConfig)
		contentRouter.PUT("/config/:method", controller.PutDeviceConfig)
		contentRouter.POST("/scores", controller.PostContentScores)
	}
}

func registerNotificationsRoutes(router *core.RouterGroup) {
	router.GET("/notifications", controller.GetNotifications)
}

var logReq = func(c core.Context) error {
	buf, _ := ioutil.ReadAll(c.Request().Body)
	core.Logger.Infof("url: %s, body: %s", c.Request().URL, readBody(bytes.NewBuffer(buf)))
	return nil
}

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	s := buf.String()
	return s
}
