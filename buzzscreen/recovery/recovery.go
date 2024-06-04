package recovery

import (
	"fmt"

	"runtime/debug"

	"math/rand"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
)

// LogRecoverWith func definition
func LogRecoverWith(obj interface{}) {
	if r := recover(); r != nil {
		core.Logger.Errorf("LogRecoverWith - object: %+v\n%s", obj, debug.Stack())
	}
}

// LimitedLogRecoverWith func definition
func LimitedLogRecoverWith(obj interface{}, chance int) {
	if r := recover(); r != nil {
		rand.Seed(time.Now().UnixNano())
		switch rand.Int() % chance {
		case 0:
			core.Logger.Errorf("LimitedLogRecoverWith - err: %v\nobject: %+v\n%s", r, obj, debug.Stack())
		default:
			core.Logger.Warnf("LimitedLogRecoverWith - err: %v\nobject: %+v\n%s", r, obj, debug.Stack())
		}
	}
}

// MutedRecovery func definition
func MutedRecovery(c core.Context) error {
	if r := recover(); r != nil {
		logger := core.Logger.WithField("http_request", c.Request())
		switch x := r.(type) {
		case error:
			logger = logger.WithError(x)
		}

		logger.Warn(fmt.Sprint(r))
	}
	return nil
}
