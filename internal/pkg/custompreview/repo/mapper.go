package repo

import (
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview"
)

type entityMapper struct {
}

func (m *entityMapper) dbConfigToConfig(dbConfig DBConfig) *custompreview.Config {
	return &custompreview.Config{
		ID:         dbConfig.ID,
		UnitID:     dbConfig.UnitID,
		Message:    dbConfig.Message,
		LandingURL: dbConfig.LandingURL,
		Period: custompreview.Period{
			StartDate:       dbConfig.StartDate,
			EndDate:         dbConfig.EndDate,
			StartHourMinute: dbConfig.StartHourMinute,
			EndHourMinute:   dbConfig.EndHourMinute,
		},
		FrequencyLimit: custompreview.FrequencyLimit{
			DIPU: dbConfig.DIPU,
			TIPU: dbConfig.TIPU,
			DCPU: dbConfig.DCPU,
			TCPU: dbConfig.TCPU,
		},
		Icon: dbConfig.Icon,
	}
}
