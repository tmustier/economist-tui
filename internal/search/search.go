package search

import "strings"

func Match(text, query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return true
	}

	text = strings.ToLower(text)
	tokens := strings.Fields(strings.ToLower(query))

	for _, token := range tokens {
		if !fuzzyContains(text, token) {
			return false
		}
	}
	return true
}

// fuzzyContains checks if all characters in query appear in text in order.
func fuzzyContains(text, query string) bool {
	if query == "" {
		return true
	}

	queryIdx := 0
	for i := 0; i < len(text) && queryIdx < len(query); i++ {
		if text[i] == query[queryIdx] {
			queryIdx++
		}
	}
	return queryIdx == len(query)
}
