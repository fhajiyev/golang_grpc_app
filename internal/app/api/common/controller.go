package common

import (
	"github.com/Buzzvil/buzzlib-go/core"
)

// ControllerBase type definition
type ControllerBase struct {
}

// Bind func definition
func (con *ControllerBase) Bind(c core.Context, model interface{}, callbacks ...(func(c core.Context, model interface{}) error)) error {
	if err := c.Bind(model); err != nil {
		return NewBindError(err)
	}

	if err := c.Validate(model); err != nil {
		return NewBindError(err)
	}

	for _, callback := range callbacks {
		if err := callback(c, model); err != nil {
			return NewBindError(err)
		}
	}
	return nil
}
