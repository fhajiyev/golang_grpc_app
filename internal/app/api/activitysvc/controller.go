package activitysvc

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/activitysvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/location"
	uuid "github.com/satori/go.uuid"
)

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, deviceUseCase device.UseCase, locUseCase location.UseCase) Controller {
	con := Controller{
		locUseCase:    locUseCase,
		deviceUseCase: deviceUseCase,
	}
	e.POST("api/activity/", con.PostActivity)
	return con
}

// Controller type definition
type Controller struct {
	*common.ControllerBase
	locUseCase    location.UseCase
	deviceUseCase device.UseCase
}

// PostActivity log device's activity that will be sent to redshift later
func (con *Controller) PostActivity(c core.Context) error {
	var activityReq dto.ActivityRequest
	if err := con.Bind(c, &activityReq); err != nil {
		return err
	}
	core.Logger.Debugf("PostActivityV1 - activityReq: %v", activityReq)

	if activityReq.UnitDeviceToken != nil {
		ok, err := con.deviceUseCase.ValidateUnitDeviceToken(*activityReq.UnitDeviceToken)
		if !ok {
			core.Logger.Infof("PostActivity() - failed to validate unit_device_token. set unit_device_token to nil. err: %s req: %+v", err.Error(), activityReq)
			activityReq.UnitDeviceToken = nil
		}
	}

	// get ip and country
	ip := net.ParseIP(c.RealIP())
	if ip = ip.To4(); ip != nil {
		activityReq.IP = ip.String()
		if country, err := con.locUseCase.GetCountryFromIP(ip); err == nil {
			activityReq.IPCountry = country
		}
	}
	jsonData, err := con.getActivityJSONStr(&activityReq)
	if err != nil {
		return common.NewBindError(err)
	}
	logObj := map[string]interface{}{
		"log_type":  "Activity",
		"json_data": jsonData,
		"message":   "general",
	}

	// Log to general logger
	core.Loggers["general"].WithFields(logObj).Info("Log")

	return c.JSON(http.StatusOK, dto.ActivityResponse{
		Gudid:      activityReq.Gudid,
		Period:     600,
		TypeFilter: "all",
	})
}

func (con *Controller) getActivityJSONStr(actReq *dto.ActivityRequest) (string, error) {
	if actReq.Gudid == "" {
		actReq.Gudid = uuid.NewV4().String()
	}
	// try casting unit_id
	_, err := strconv.ParseInt(actReq.UnitID, 10, 64)
	if err != nil {
		return "", err
	}

	// Set bi_event parameters
	switch actReq.UnitID {
	case "416524360013206":
		actReq.AppID = "11"
	case "210342277740215":
		actReq.AppID = "10"
	default:
		actReq.AppID = actReq.UnitID
	}

	// calcuate offset
	ts, err := strconv.ParseInt(actReq.DeviceTimestamp, 10, 64)
	if err != nil {
		return "", err
	}
	clientTime := time.Unix(ts, 0)
	serverTime := time.Now()
	offset := serverTime.Sub(clientTime)
	if offset.Minutes() > 10 || offset.Minutes() < -10 {
		// Adjust if offset is more/less than 10min
		actReq.ServerTime = serverTime.Unix()
		actReq.ClientTime = clientTime.Unix()
		// Adjust createdAt/timestamp inside activity
		for _, a := range actReq.Activities {
			act := map[string]interface{}(*a)
			createdAt := time.Unix(int64(act["timestamp"].(float64)), 0)
			createdAt = createdAt.Add(offset)
			act["client_created_at"] = act["timestamp"].(float64) // Keep raw
			act["timestamp"] = createdAt.Unix()
		}
	}

	actJSON, err := json.Marshal(actReq)
	if err != nil {
		return "", err
	}
	return string(actJSON), nil
}
