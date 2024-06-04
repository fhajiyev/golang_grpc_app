package monitorsvc

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const numLiveGoroutineCount = 2500

// Controller type definition
type Controller struct {
}

// GetMonitorGoroutines returns the number of goroutine
func (con *Controller) GetMonitorGoroutines(c core.Context) error {
	return pprof.Lookup("goroutine").WriteTo(c.Response(), 1)
}

// GetMonitorMemory returns memory usage
func (con *Controller) GetMonitorMemory(c core.Context) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"Alloc":      fmt.Sprintf("%vmb", bToMb(m.Alloc)),
		"TotalAlloc": fmt.Sprintf("%vmb", bToMb(m.TotalAlloc)),
		"Sys":        fmt.Sprintf("%vmb", bToMb(m.Sys)),
		"NumGC":      fmt.Sprintf("%v", m.NumGC),
	})
}

// Live returns error if the count of goroutines is higher than numLiveGoroutineCount.
func (con *Controller) Live(c core.Context) error {
	if numGoroutine := runtime.NumGoroutine(); numGoroutine > numLiveGoroutineCount {
		return &core.HttpError{
			Code:    http.StatusServiceUnavailable,
			Message: fmt.Sprintf("Too many goroutines: %d", numGoroutine),
		}
	}
	return nil
}

// Ready returns false when checking network connectivity fails or the count of goroutines is higher than numReadyGoroutineCount.
func (con *Controller) Ready(c core.Context) error {
	return con.checkESConnectivity()
}

func (con *Controller) checkESConnectivity() error {
	var res ESHealthRes
	req := &network.Request{
		Method:  http.MethodGet,
		URL:     env.Config.ElasticSearch.Host + "/_cat/health?format=JSON",
		Timeout: time.Second,
	}
	code, err := req.GetResponse(&res)
	if err != nil {
		return err
	}
	if code != http.StatusOK || len(res) == 0 {
		return fmt.Errorf("elasticsearch is not found")
	}
	if res[0].Status != ESStatusOK {
		return fmt.Errorf("elasticsearch is not ready. status - %s", res[0].Status)
	}
	return nil
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine) Controller {
	con := Controller{}
	// Deprecated
	e.GET("/go/common/ping", con.Live)
	// Deprecated
	e.GET("/common/ping/", con.Live)

	e.GET("/live", con.Live)
	e.GET("/ready", con.Ready)

	e.GET("/monitor/goroutine", con.GetMonitorGoroutines)
	e.GET("/monitor/memory", con.GetMonitorMemory)

	promHandler := promhttp.Handler()
	promHandleFunc := func(c core.Context) error {
		promHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	}
	e.GET("/monitor/metrics", promHandleFunc)
	return con
}
