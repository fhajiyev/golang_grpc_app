package reward

// Repository provides storing RequestIngredients into DB
type Repository interface {
	Save(req RequestIngredients) (int, error)
	GetImpressionPoints(deviceID int64, maxPeriod Period) []Point
}
