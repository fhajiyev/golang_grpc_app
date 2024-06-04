package report

// Repository interface definition
type Repository interface {
	SaveContentReport(camp Request) error
	SaveAdReport(camp Request) error
}
