package app

import (
	"context"
)

// UseCase type definition
type UseCase interface {
	GetAppByID(ctx context.Context, appID int64) (*App, error)

	GetRewardableWelcomeRewardConfig(ctx context.Context, unitID int64, country string, unitRegisterSeconds int64) (*WelcomeRewardConfig, error)
	GetActiveWelcomeRewardConfigs(ctx context.Context, unitID int64) (WelcomeRewardConfigs, error)
	GetReferralRewardConfig(ctx context.Context, appID int64) (*ReferralRewardConfig, error)

	GetUnitByID(ctx context.Context, unitID int64) (*Unit, error)
	GetUnitByAppID(ctx context.Context, appID int64) (*Unit, error)
	GetUnitByAppIDAndType(ctx context.Context, appID int64, unitType UnitType) (*Unit, error)
}

type useCase struct {
	repo Repository
}

var _ UseCase = &useCase{}

// GetRewardableWelcomeRewardConfig returns all configs whose start time and (end time + retention days) period
// includes current time
func (u *useCase) GetRewardableWelcomeRewardConfig(ctx context.Context, unitID int64, country string, unitRegisterSeconds int64) (*WelcomeRewardConfig, error) {
	wrcs, err := u.repo.GetRewardingWelcomeRewardConfigs(ctx, unitID)
	if err != nil {
		return nil, err
	}
	if len(wrcs) == 0 {
		return nil, nil
	}
	wrcs = wrcs.FilterRewardable(unitRegisterSeconds)

	targetCountryConfig := wrcs.FilterWithCountry(country).FindOngoingWRC()
	// return the country matching config if it exists
	if targetCountryConfig != nil {
		return targetCountryConfig, nil
	}

	// if there is no matching country config, but there exists a rewardable global config
	// return the global config
	globalConfig := wrcs.FilterWithCountry(countryGlobal).FindOngoingWRC()
	if globalConfig != nil {
		return globalConfig, nil
	}

	// if country is global and there exists a uniuqe rewardable config, return the unique config
	if country == countryGlobal && len(wrcs) == 1 {
		return &wrcs[0], nil
	}

	return nil, nil
}

func (u *useCase) GetActiveWelcomeRewardConfigs(ctx context.Context, unitID int64) (WelcomeRewardConfigs, error) {
	wrcs, err := u.repo.GetRewardingWelcomeRewardConfigs(ctx, unitID)
	if err != nil {
		return nil, err
	} else if len(wrcs) == 0 {
		return nil, nil
	}

	return wrcs.FilterActive(), nil
}

// GetReferralRewardConfig func definition
func (u *useCase) GetReferralRewardConfig(ctx context.Context, appID int64) (*ReferralRewardConfig, error) {
	return u.repo.GetReferralRewardConfig(ctx, appID)
}

// GetAppByID func definition
func (u *useCase) GetAppByID(ctx context.Context, appID int64) (*App, error) {
	return u.repo.GetAppByID(ctx, appID)
}

// GetUnitByID func definition
func (u *useCase) GetUnitByID(ctx context.Context, unitID int64) (*Unit, error) {
	return u.repo.GetUnitByID(ctx, unitID)
}

// GetUnitByAppID func definition
func (u *useCase) GetUnitByAppID(ctx context.Context, appID int64) (*Unit, error) {
	return u.repo.GetUnitByAppID(ctx, appID)
}

// GetUnitByAppIDAndType func definition
func (u *useCase) GetUnitByAppIDAndType(ctx context.Context, appID int64, unitType UnitType) (*Unit, error) {
	return u.repo.GetUnitByAppIDAndType(ctx, appID, unitType)
}

// NewUseCase returns new useCase
func NewUseCase(appRepo Repository) UseCase {
	return &useCase{appRepo}
}
