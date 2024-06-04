package cookiehandler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/clickredirectsvc/cookiehandler"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/clickredirectsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/stretchr/testify/suite"
)

const (
	// expireOffset const definition
	expireOffset = 365 * 24 * time.Hour
	// timeFormat const definition
	timeFormat = "Mon, 02 Jan 2006 15:00:00 GMT"
	// cookieVersion const definition
	cookieVersion = "3"
)

func TestCookieHandlerSuite(t *testing.T) {
	suite.Run(t, new(CookieHandlerTestSuite))
}

type CookieHandlerTestSuite struct {
	suite.Suite
	cookieHandler cookiehandler.CookieHandler
	engine        *core.Engine
}

func (ts *CookieHandlerTestSuite) SetupSuite() {
	tests.GetTestServer(nil)
	ts.engine = core.NewEngine(nil)
}

func (ts *CookieHandlerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (ctx core.Context, rec *httptest.ResponseRecorder) {
	rec = httptest.NewRecorder()
	ctx = ts.engine.NewContext(httpRequest, rec)
	return
}

func (ts *CookieHandlerTestSuite) Test_SetCookieAndChecksum_NotUpdated() {
	networkReq := ts.buildNetworkRequest(nil)

	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "cookie_id",
		Value: "sample_cookie_id_value",
	})
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "checksum",
		Value: "e7da932274da2c857dece05711a6cc58",
	})
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "cookie_version",
		Value: cookieVersion,
	})

	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())
	ts.cookieHandler = cookiehandler.CookieHandler{Ctx: ctx}

	updated, cookieID := ts.cookieHandler.SetCookieIDAndChecksum("sample_ifa_value")
	ts.False(updated)
	ts.Equal(cookieID, "sample_cookie_id_value")
	ts.Equal(ctx.Response().Header().Get("Set-Cookie"), "")
}

func (ts *CookieHandlerTestSuite) Test_SetCookieAndChecksum_Updated_NothingSet() {
	networkReq := ts.buildNetworkRequest(nil)
	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())
	ts.cookieHandler = cookiehandler.CookieHandler{Ctx: ctx}

	updated, cookieID := ts.cookieHandler.SetCookieIDAndChecksum("sample_ifa_value")
	ts.True(updated)
	ts.Equal(cookieID, "6197673a-2900-5277-bd18-0d78cd83f831")
}

func (ts *CookieHandlerTestSuite) Test_SetCookieAndChecksum_Updated_NoChecksum() {
	networkReq := ts.buildNetworkRequest(nil)
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "cookie_id",
		Value: "sample_cookie_id_value",
	})
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "cookie_version",
		Value: cookieVersion,
	})

	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())
	expiration := time.Now().Add(expireOffset).UTC().Format(timeFormat)
	ts.cookieHandler = cookiehandler.CookieHandler{Ctx: ctx}

	updated, cookieID := ts.cookieHandler.SetCookieIDAndChecksum("sample_ifa_value")
	ts.True(updated)
	ts.Equal(cookieID, "sample_cookie_id_value")
	ts.Equal(ctx.Response().Header().Get("Set-Cookie"), "checksum=e7da932274da2c857dece05711a6cc58;Domain=buzzvil.com;Expires="+expiration+";Path=/;SameSite=None;Secure")
}

func (ts *CookieHandlerTestSuite) Test_SetCookieAndChecksum_Updated_NoCookieID() {
	networkReq := ts.buildNetworkRequest(nil)
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "checksum",
		Value: "0bf877c58c778af9ddd5676a46068cd6",
	})
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "cookie_version",
		Value: cookieVersion,
	})

	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())
	expiration := time.Now().Add(expireOffset).UTC().Format(timeFormat)
	ts.cookieHandler = cookiehandler.CookieHandler{Ctx: ctx}

	updated, cookieID := ts.cookieHandler.SetCookieIDAndChecksum("sample_ifa_value")
	ts.True(updated)
	ts.Equal(cookieID, "6197673a-2900-5277-bd18-0d78cd83f831")
	ts.Equal(ctx.Response().Header().Get("Set-Cookie"), "cookie_id=6197673a-2900-5277-bd18-0d78cd83f831;Domain=buzzvil.com;Expires="+expiration+";Path=/;SameSite=None;Secure")
}

func (ts *CookieHandlerTestSuite) Test_SetCookieAndChecksum_Updated_OtherIFA() {
	networkReq := ts.buildNetworkRequest(nil)
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "cookie_id",
		Value: "sample_cookie_id_value",
	})
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "checksum",
		Value: "e7da932274da2c857dece05711a6cc58",
	})
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "cookie_version",
		Value: cookieVersion,
	})

	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())
	expiration := time.Now().Add(expireOffset).UTC().Format(timeFormat)
	ts.cookieHandler = cookiehandler.CookieHandler{Ctx: ctx}

	updated, cookieID := ts.cookieHandler.SetCookieIDAndChecksum("sample_ifa_value_other")
	ts.True(updated)
	ts.Equal(cookieID, "sample_cookie_id_value")
	ts.Equal(ctx.Response().Header().Get("Set-Cookie"), "checksum=d5e954968fd530add9af1f1f7d9368cf;Domain=buzzvil.com;Expires="+expiration+";Path=/;SameSite=None;Secure")
}

func (ts *CookieHandlerTestSuite) Test_SetCookieAndChecksum_Updated_NoVersion() {
	networkReq := ts.buildNetworkRequest(nil)
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "cookie_id",
		Value: "sample_cookie_id_value",
	})
	networkReq.GetHTTPRequest().AddCookie(&http.Cookie{
		Name:  "checksum",
		Value: "e7da932274da2c857dece05711a6cc58",
	})

	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())
	expiration := time.Now().Add(expireOffset).UTC().Format(timeFormat)
	ts.cookieHandler = cookiehandler.CookieHandler{Ctx: ctx}

	updated, cookieID := ts.cookieHandler.SetCookieIDAndChecksum("sample_ifa_value")
	ts.True(updated)
	ts.Equal(cookieID, "sample_cookie_id_value")
	ts.Equal(ctx.Response().Header().Get("Set-Cookie"), "cookie_version="+cookieVersion+";Domain=buzzvil.com;Expires="+expiration+";Path=/;SameSite=None;Secure")
}

func (ts *CookieHandlerTestSuite) buildNetworkRequest(req *dto.GetClickRedirectRequest) *network.Request {
	networkReq := &network.Request{
		Method: http.MethodGet,
		Header: &http.Header{
			"User-Agent": {"Mozilla/5.0 (Linux; Android 4.2.1; en-us; Nexus 5 Build/JOP40D) AppleWebKit/535.19 (KHTML, like Gecko; googleweblight) Chrome/38.0.1025.166 Mobile Safari/535.19"},
		},
	}

	return networkReq.Build()
}
