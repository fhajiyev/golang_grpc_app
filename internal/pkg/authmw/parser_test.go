package authmw_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
)

type testValidator struct {
	validator *validator.Validate
}

func (tv *testValidator) Validate(i interface{}) error {
	return tv.validator.Struct(i)
}

func newApplicationFormContext(values url.Values) echo.Context {
	engine := core.NewEngine(nil)
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(values.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	return engine.NewContext(req, rec)
}
