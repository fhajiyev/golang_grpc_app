package echo

import (
	"fmt"

	"github.com/labstack/echo"
	nr "github.com/newrelic/go-agent"
)

const (
	// NewrelicTxn defines the context key used to save newrelic transaction
	NewrelicTxn = "newrelic-txn"
)

// NewRelicWithApplication returns a NewRelic middleware with application.
// See: `NewRelic()`.
//noinspection GoUnhandledErrorResult
func NewRelicWithApplication(app nr.Application) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			transactionName := fmt.Sprintf("%s [%s]", c.Path(), c.Request().Method)
			txn := app.StartTransaction(transactionName, c.Response().Writer, c.Request())
			defer txn.End()

			c.Set(NewrelicTxn, txn)

			err := next(c)

			if err != nil {
				_ = txn.NoticeError(err)
			}

			return err
		}
	}
}
