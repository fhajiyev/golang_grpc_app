package dbapp

import "context"

// DBSource interface definition
type DBSource interface {
	GetAppByID(ctx context.Context, appID int64) (*App, error)
	FindRewardingWelcomeRewardConfigs(ctx context.Context, unitID int64) ([]WelcomeRewardConfig, error)
	FindReferralRewardConfig(ctx context.Context, appID int64) (*ReferralRewardConfig, error)
	GetUnit(ctx context.Context, unit *Unit) (*Unit, error)
}
