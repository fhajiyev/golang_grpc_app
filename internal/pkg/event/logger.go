package event

// StructuredLogger interface
type StructuredLogger interface {
	Log(m map[string]interface{})
}
