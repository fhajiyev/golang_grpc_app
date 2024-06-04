package app

import "context"

// Repository interface definition
type Repository interface {
	GetAppByID(ctx context.Context, appID int64) (*App, error)

	GetRewardingWelcomeRewardConfigs(ctx context.Context, unitID int64) (WelcomeRewardConfigs, error)
	GetReferralRewardConfig(ctx context.Context, appID int64) (*ReferralRewardConfig, error)

	GetUnitByID(ctx context.Context, unitID int64) (*Unit, error)
	GetUnitByAppID(ctx context.Context, appID int64) (*Unit, error)
	GetUnitByAppIDAndType(ctx context.Context, appID int64, unitType UnitType) (*Unit, error)
}
