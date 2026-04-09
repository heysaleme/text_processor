package main

import (
	"strings"
	"unicode"
)

func isWordToken(token string) bool {
	if strings.Contains(token, "'") {
		for _, char := range token {
			if char != '\'' && !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
				return false
			}
		}
		return len(token) > 0
	}

	for _, char := range token {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return false
		}
	}

	return len(token) > 0
}
