package config

// Repository defines an interface for repository
type Repository interface {
	GetConfigs(configReq RequestIngredients) *[]Config
}
