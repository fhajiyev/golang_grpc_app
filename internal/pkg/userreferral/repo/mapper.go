package repo

import "github.com/Buzzvil/buzzscreen-api/internal/pkg/userreferral"

type entityMapper struct {
}

func (m entityMapper) dbDeviceUserToDeviceUser(user DBDeviceUser) *userreferral.DeviceUser {
	return &userreferral.DeviceUser{
		ID:         user.ID,
		DeviceID:   user.DeviceID,
		Code:       user.Code,
		ReferrerID: user.ReferrerID,
		IsVerified: user.IsVerified,
	}
}
