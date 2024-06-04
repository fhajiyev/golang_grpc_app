package notiplus

// Config entity definition
type Config struct {
	ID                 int64
	UnitID             int64
	Title              string
	Description        string
	Icon               string
	ScheduleHourMinute string
}
