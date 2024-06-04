package controller_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	uuid "github.com/satori/go.uuid"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/Buzzvil/go-test/test"
)

func TestDevice(t *testing.T) {
	adID := uuid.NewV4().String()
	response, code := execTestDevice(t, tests.HsKrAppID, fmt.Sprintf("TestUser-%d", rand.Intn(100)), adID)
	test.AssertEqual(t, response.Result["session_key"] != nil, true, "Response not nil")

	checkDeviceJSONFile(t)
	defer truncateDeviceJSONFile(t)

	userID := fmt.Sprintf("TestUser-%d", rand.Intn(100))
	response, code = execTestDevice(t, tests.HsKrAppID, userID, adID)
	test.AssertEqual(t, response.Result["session_key"] != nil, true, "Response not nil")

	newAdID := uuid.NewV4().String()
	response, code = execTestDevice(t, tests.HsKrAppID, fmt.Sprintf("TestUser-%d", rand.Intn(100)), newAdID)
	test.AssertEqual(t, response.Result["session_key"] != nil, true, "Response not nil")

	response, code = execTestDevice(t, tests.HsKrAppID, userID, newAdID)
	test.AssertEqual(t, response.Result["session_key"] != nil, true, "Response not nil")

	session, err := buzzscreen.Service.SessionUseCase.GetSessionFromKey(response.Result["session_key"].(string))
	if err != nil || session.DeviceID == 0 {
		t.Fatalf("execTestSession() - err: %v, session: %v", err, session)
	}

	_, code = execTestDevice(t, tests.DeactivatedAppID, userID, newAdID)
	core.Logger.Warnf("devicde response %+v", response)
	test.AssertEqual(t, code, 400, "return 400 for deactivated app")
}

func truncateDeviceJSONFile(t *testing.T) {
	logFile, err := os.OpenFile(env.Config.Loggers["device"].File, os.O_RDWR, 0666)
	defer logFile.Close()
	if err != nil {
		t.Fatalf("truncateDeviceJSONFile() - %s", err)
	} else {
		logFile.Truncate(0)
	}
}

func checkDeviceJSONFile(t *testing.T) {
	if jsonData, err := ioutil.ReadFile(env.Config.Loggers["device"].File); err != nil || len(jsonData) == 0 {
		t.Fatalf("checkDeviceJSONFile() - %s, %s", err, env.Config.Loggers["device"].File)
	} else {
		t.Logf("checkDeviceJSONFile() - %s", jsonData)
	}
}

func execTestDevice(t *testing.T, appID int64, userID, adID string) (*dto.CreateDeviceResponse, int) {
	params := &url.Values{
		"app_id":      {fmt.Sprintf("%d", appID)},
		"ad_id":       {adID},
		"user_id":     {userID},
		"android_id":  {"androidID_" + userID},
		"resolution":  {"1080x1776"},
		"device_name": {"Nexus One"},
		"timezone":    {"America/Los_Angeles"},
		"locale":      {"ko_KR"},
	}

	request := network.Request{
		Method:  "POST",
		Params:  params,
		URL:     ts.URL + "/api/v3/device",
		Timeout: time.Minute * 10,
	}

	// t.Log("execTestDevice() - request:", request, *params)
	var response dto.CreateDeviceResponse
	statusCode, err := request.GetResponse(&response)
	// t.Log("execTestDevice() - response:", response)
	if err != nil {
		t.Fatal(statusCode, err, response)
		return nil, 0
	}
	return &response, statusCode
}
