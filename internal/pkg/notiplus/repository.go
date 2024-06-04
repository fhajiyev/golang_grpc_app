package notiplus

// Repository interface definition
type Repository interface {
	GetConfigsByUnitID(unitID int64) ([]Config, error)
}
