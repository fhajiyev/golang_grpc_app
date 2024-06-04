package bauserrepo

// BAUser record definition
type BAUser struct {
	ID          int64 `gorm:"primary_key"`
	AccessToken string
	Name        string
	IsMedia     bool
}

// TableName returns table name for BAUser
func (BAUser) TableName() string {
	return "buzzad_user"
}
