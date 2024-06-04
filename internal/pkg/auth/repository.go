package auth

// Repository handles auth data.
type Repository interface {
	CreateAuth(identifier Identifier) (string, error)
	GetAuth(token string) (*Auth, error)
}
