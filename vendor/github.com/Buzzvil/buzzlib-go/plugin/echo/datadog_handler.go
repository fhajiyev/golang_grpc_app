package echo

import (
	"github.com/labstack/echo"
	dd "gopkg.in/DataDog/dd-trace-go.v1/contrib/labstack/echo"
)

func DatadogTrace() echo.MiddlewareFunc {
	return dd.Middleware()
}
