package profilerequest

// Repository provides profile ID population by sending request to profilesvc.
type Repository interface {
	PopulateProfile(account Account) error
}
