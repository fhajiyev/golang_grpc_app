package utils

// TransformMapFromCamelToUnderScore transformed camel map into underscore map
func TransformMapFromCamelToUnderScore(camelMap map[string]interface{}) map[string]interface{} {
	underScoreMap := map[string]interface{}{}
	for k, v := range camelMap {
		underScoreMap[CamelToUnderscore(k)] = v
	}

	return underScoreMap
}
