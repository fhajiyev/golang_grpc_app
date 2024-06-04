package repo

import (
	"regexp"
	"strings"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/config"
)

// Repository struct definition
type Repository struct {
}

// GetConfigs returns config entities based on the request parameters.
func (r *Repository) GetConfigs(configReq config.RequestIngredients) *[]config.Config {
	// TODO: 이것은 hard-coded된 값으로, 추후 config service가 생길시, 옮겨야함.
	configs := []config.Config{
		config.Config{
			Key:   "BATTERY_OPTIMIZATION_GUIDE_PAGE_URL",
			Value: "https://ad-client.buzzvil.com/optimization/battery/",
		},
		config.Config{
			Key:   "YOUTUBE_KEY",
			Value: "AIzaSyDfXQ2ilCqLsIluMoQu55qOtZfK4L6e5uM",
		},
		config.Config{
			Key:   "V3_TEST_ENABLED",
			Value: "false",
		},
		config.Config{
			Key:   "V3_TEST_THRESHOLD",
			Value: "7330025",
		},
		config.Config{
			Key:   "STU_OPT_OUT_VIEW_COUNT_THRESHOLD",
			Value: "3",
		},
		config.Config{
			Key:   "OVERLAY_PERMISSION_GUIDE_FORCED",
			Value: "false",
		},
	}

	configStuOptEnabled := config.Config{
		Key:   "STU_OPT_OUT_ENABLED",
		Value: "false",
	}

	switch configReq.UnitID {
	case 210342277740215: // SJ
		configStuOptEnabled.Value = "true"
	case 419318955785795: // BS-Sample
		configStuOptEnabled.Value = "true"
	case 100000043, 100000045, 100000050: // HS-KR/JP/TW
		configStuOptEnabled.Value = "true"
	case 415722463832660: // CLiP
		configStuOptEnabled.Value = "true"
	case 289984773704012: // UPlus
		configStuOptEnabled.Value = "true"
	case 479287856771881: // LiivMate
		configStuOptEnabled.Value = "true"
	case 391284973618433: // 하나멤버스
		configStuOptEnabled.Value = "true"
	case 509710633829287: // WeMakePrice
		configStuOptEnabled.Value = "true"
	case 291112716543436: // H.Point
		configStuOptEnabled.Value = "true"
	case 339396701870506: // [JP] GetMoney
		configStuOptEnabled.Value = "true"
	case 370467839098518: // [JP] JRE
		configStuOptEnabled.Value = "true"
	case 495496451416174: // [JP] Ponta
		configStuOptEnabled.Value = "true"
	case 34236817485821: // [JP] LINE
		configStuOptEnabled.Value = "true"
	case 557935196158088: // [JP] Influence
		configStuOptEnabled.Value = "true"
	}

	configs = append(configs, configStuOptEnabled)

	configOverlayPermissionGuidePeriod := config.Config{
		Key:   "OVERLAY_PERMISSION_GUIDE_PERIOD_IN_SEC",
		Value: "21600", // 6 hours
	}

	switch configReq.UnitID {
	case 320907354574352: // 서울도시가스
		configOverlayPermissionGuidePeriod.Value = "43200" // 12 hours
	case 20867641205588: // OCB
		configOverlayPermissionGuidePeriod.Value = "31536000" // 365 days - TS-4222
	case 289984773704012: // UPlus
		configOverlayPermissionGuidePeriod.Value = "86400" // 24 hours - TS-4458
	}
	configs = append(configs, configOverlayPermissionGuidePeriod)

	configBatteryOptGuideEnabled := config.Config{
		Key:   "BATTERY_OPTIMIZATION_GUIDE_ENABLED",
		Value: "false",
	}
	//BS-2033 20867641205588, 238793895727606
	//if LG and SAMSUNG that aren't OCB apps, return true. ELSE return false.
	if (configReq.UnitID != 20867641205588 && configReq.UnitID != 238793895727606) && r.checkManufacturer(configReq.Manufacturer) {
		configBatteryOptGuideEnabled.Value = "true"
	}

	configs = append(configs, configBatteryOptGuideEnabled)
	return &configs
}

func (r *Repository) checkManufacturer(manufacturer string) bool {
	//SAMSUNG - ^s[a-z]{5}g
	//^LG - ^LG
	regex, err := regexp.Compile("^(lg)|^s[a-z]{5}g")
	if err != nil {
		return false
	}
	return regex.MatchString(strings.ToLower(manufacturer))
}

// New returns Repository struct
func New() *Repository {
	return &Repository{}
}
