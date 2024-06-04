package bauserrepo

import "github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"

type entityMapper struct {
}

func (m *entityMapper) dbBAUserToBAUser(baUser *BAUser) *ad.BAUser {
	return &ad.BAUser{
		ID:          baUser.ID,
		AccessToken: baUser.AccessToken,
		Name:        baUser.Name,
		IsMedia:     baUser.IsMedia,
	}
}
