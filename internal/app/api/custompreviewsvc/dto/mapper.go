package dto

import "github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview"

const infinity = 1 << 28

// Mapper struct definition
type Mapper struct {
}

// ConfigToDTOConfig maps config to dto config
func (m Mapper) ConfigToDTOConfig(config custompreview.Config) *Config {
	return &Config{
		ID:         config.ID,
		UnitID:     config.UnitID,
		Message:    config.Message,
		LandingURL: config.LandingURL,
		Period: Period{
			StartDate:       config.Period.StartDate,
			EndDate:         config.Period.EndDate,
			StartHourMinute: config.Period.StartHourMinute,
			EndHourMinute:   config.Period.EndHourMinute,
		},
		FrequencyLimit: FrequencyLimit{
			DIPU: m.getInfIfNil(config.FrequencyLimit.DIPU),
			TIPU: m.getInfIfNil(config.FrequencyLimit.TIPU),
			DCPU: m.getInfIfNil(config.FrequencyLimit.DCPU),
			TCPU: m.getInfIfNil(config.FrequencyLimit.TCPU),
		},
		Icon: config.Icon,
	}
}

func (m Mapper) getInfIfNil(arg *int) int {
	if arg == nil {
		return infinity
	}
	return *arg
}
