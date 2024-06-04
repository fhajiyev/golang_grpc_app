package dto_test

import (
	"context"
	"testing"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"

	"github.com/Buzzvil/go-test/test"
)

func TestAd_GetName(t *testing.T) {
	ad := dto.Ad{}
	name := ad.GetName()
	test.AssertEqual(t, name, &ad.Name, "If not setted creative and seted name, just return ad.Name pointer.")

	ad.Name = "name"
	name = ad.GetName()
	test.AssertEqual(t, *name, "name", "If not setted creative, and seted name, return name.")

	ad.Creative = map[string]interface{}{
		"title": "title",
	}
	name = ad.GetName()
	test.AssertEqual(t, *name, "title", "If seted Creatives's title, return Creatives's title instead ad.Name.")

}

func TestAd_GetRewardSum(t *testing.T) {
	ad := dto.Ad{}
	ad.LandingReward = 3
	ad.ActionReward = 20
	ad.UnlockReward = 100

	sum := ad.GetRewardSum()
	test.AssertEqual(t, sum, 123, "Sum LandingReward, ActionReward, UnlockReward.")
}

func TestAds_Len(t *testing.T) {
	var ads dto.Ads = dto.Ads([]*dto.Ad{
		&dto.Ad{}, &dto.Ad{},
	})
	len := ads.Len()
	test.AssertEqual(t, len, 2, "Return dto.Ad's count.")
}

func TestAds_Swap(t *testing.T) {
	var ads dto.Ads = dto.Ads([]*dto.Ad{
		&dto.Ad{
			Name: "1",
		},
		&dto.Ad{
			Name: "2",
		},
		&dto.Ad{
			Name: "3",
		},
	})

	ads.Swap(1, 2)

	test.AssertEqual(t, ads[2].Name, "2", "Swap second, third dto.Ad.")
	test.AssertEqual(t, ads[1].Name, "3", "Swap second, third dto.Ad.")
}

func TestAdsRequest_GetTargetFill(t *testing.T) {
	req := &dto.AdsRequest{}
	targetFill := req.GetTargetFill()
	test.AssertEqual(t, targetFill, 0, "Get TargetFill(ad's len:0)")

	req.TargetFill = 2
	targetFill = req.GetTargetFill()
	test.AssertEqual(t, targetFill, 2, "Get TargetFill")
}

type mockAppUseCase struct{}

func (m *mockAppUseCase) GetAppByID(ctx context.Context, appID int64) (*app.App, error) {
	return nil, nil
}

func (m *mockAppUseCase) GetRewardableWelcomeRewardConfig(ctx context.Context, unitID int64, country string, unitRegisterSeconds int64) (*app.WelcomeRewardConfig, error) {
	return nil, nil

}
func (m *mockAppUseCase) GetActiveWelcomeRewardConfigs(ctx context.Context, unitID int64) (app.WelcomeRewardConfigs, error) {
	return nil, nil

}
func (m *mockAppUseCase) GetReferralRewardConfig(ctx context.Context, appID int64) (*app.ReferralRewardConfig, error) {
	return nil, nil
}

func (m *mockAppUseCase) GetUnitByID(ctx context.Context, unitID int64) (*app.Unit, error) {
	return &app.Unit{
		AppID: int64(1),
	}, nil
}
func (m *mockAppUseCase) GetUnitByAppID(ctx context.Context, appID int64) (*app.Unit, error) {
	return nil, nil
}
func (m *mockAppUseCase) GetUnitByAppIDAndType(ctx context.Context, appID int64, unitType app.UnitType) (*app.Unit, error) {
	return nil, nil
}
