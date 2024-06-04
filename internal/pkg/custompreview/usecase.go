package custompreview

import (
	"fmt"
	"strconv"
	"time"
)

// NewUseCase returns UseCase intercase
func NewUseCase(repo Repository) UseCase {
	// time is UTC based
	return &useCase{repo: repo}
}

// UseCase interface definition
type UseCase interface {
	GetConfigByUnitID(unitID int64, timezone string, targetTime time.Time) (*Config, error)
}

type useCase struct {
	repo Repository
}

func (u *useCase) GetConfigByUnitID(unitID int64, timezone string, targetTime time.Time) (*Config, error) {
	// timezone string is only supported for constant in zip file https://www.iana.org/time-zones

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timezone to location %v", err)
	}

	// adjust time for searching db
	searchTargetTime := u.replaceTimezone(targetTime.In(location), time.UTC)

	config, err := u.repo.GetConfigByUnitID(unitID, true, searchTargetTime)
	if err != nil {
		return nil, err
	} else if config == nil {
		return nil, nil
	}

	// adjust time for client's timezone
	config.StartDate = u.replaceTimezone(config.StartDate, location)
	config.EndDate = u.replaceTimezone(config.EndDate, location)
	config.StartHourMinute, err = u.removeTimezoneToHourMinute(config.StartHourMinute, location, targetTime)
	if err != nil {
		return nil, err
	}
	config.EndHourMinute, err = u.removeTimezoneToHourMinute(config.EndHourMinute, location, targetTime)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// replaceTimezone returns time with preserved year, month,hour, minute, second, nanosecond value but change it to utc
func (u *useCase) replaceTimezone(targetTime time.Time, location *time.Location) time.Time {
	return time.Date(targetTime.Year(), targetTime.Month(), targetTime.Day(), targetTime.Hour(), targetTime.Minute(), targetTime.Second(), targetTime.Nanosecond(), location).In(time.UTC)
}

func (u *useCase) removeTimezoneToHourMinute(hourMinute string, location *time.Location, targetTime time.Time) (string, error) {
	// targetTime is used to extract offset second. The reason why we need it is that according to target time, daylight saving time can be applied then it can affect offset second
	// ex) 2020.03.01 : daylight saving time is not applied -> America/New_York -05:00
	// ex) 2020.05.01 : daylight saving time is applied -> America/New_York -04:00

	// hourMinute format is as below and it is checked when putting data to DB
	// ex) 00:00, 17:02, 23:58
	_, offsetSecond := targetTime.In(location).Zone()
	offsetHour := offsetSecond / 3600
	inputHour := hourMinute[0:2]
	inputMinute := hourMinute[3:5]

	hourInt, err := strconv.Atoi(inputHour)
	if err != nil {
		return "", err
	}
	hourInt = (hourInt - offsetHour + 24) % 24

	hourStr := strconv.Itoa(hourInt)
	if hourInt/10 == 0 {
		hourStr = "0" + hourStr
	}
	return hourStr + ":" + inputMinute, nil
}
