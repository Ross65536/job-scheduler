package internal

func MapSubmap(m map[string]interface{}, fields ...string) map[string]interface{} {
	mapCopy := make(map[string]interface{})

	for _, field := range fields {
		if v, ok := m[field]; ok {
			mapCopy[field] = v
		}
	}

	return mapCopy
}
