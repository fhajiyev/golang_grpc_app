package repo

import (
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbnotiplus"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/notiplus"
)

type entityMapper struct {
}

func (m *entityMapper) dbConfigToConfig(dbConfig *dbnotiplus.Config) notiplus.Config {
	return notiplus.Config{
		ID:                 dbConfig.ID,
		UnitID:             dbConfig.UnitID,
		Title:              dbConfig.Title,
		Description:        dbConfig.Description,
		Icon:               dbConfig.Icon,
		ScheduleHourMinute: dbConfig.ScheduleHourMinute,
	}
}
