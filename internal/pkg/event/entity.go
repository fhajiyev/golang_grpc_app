package event

// Resource struct
type Resource struct {
	ID   int64
	Type string
	Name *string
}

// ResourceType is type of resource
const (
	ResourceTypeAd      string = "ad"
	ResourceTypeArticle string = "article"
)
