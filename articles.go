package main

import (
	"strings"
	"unicode"
)

func fixArticles(tokens []string) []string {
	result := make([]string, len(tokens))
	copy(result, tokens)

	for i, word := range result {
		if strings.HasPrefix(word, "(") && strings.HasSuffix(word, ")") {
			continue
		}

		if !(word == "a" || word == "A" || word == "an" || word == "An") || i+1 >= len(result) {
			continue
		}

		j := i + 1
		for j < len(result) && result[j] != "\n" &&
			(isPunctuation(result[j]) || result[j] == "'" ||
				(strings.HasPrefix(result[j], "(") && strings.HasSuffix(result[j], ")")) ||
				strings.TrimSpace(result[j]) == "") {
			j++
		}

		if j >= len(result) || result[j] == "\n" {
			continue
		}

		next := result[j]
		if isStopWord(next) {
			continue
		}

		if shouldUseAn(next) {
			if word == "a" {
				result[i] = "an"
			}
			if word == "A" {
				result[i] = "An"
			}
			continue
		}

		if word == "an" {
			result[i] = "a"
		}
		if word == "An" {
			result[i] = "A"
		}
	}

	return result
}

func isStopWord(word string) bool {
	stopwords := map[string]bool{
		"and": true,
		"or":  true,
		"the": true,
		"a":   true,
		"an":  true,
		"of":  true,
		"for": true,
	}

	return stopwords[strings.ToLower(word)]
}

func shouldUseAn(word string) bool {
	anExceptions := map[string]bool{
		"hour":   true,
		"honest": true,
		"heir":   true,
		"honour": true,
		"honor":  true,
	}
	aExceptions := map[string]bool{
		"university": true,
		"unit":       true,
		"unicorn":    true,
		"user":       true,
		"european":   true,
	}

	wordLower := strings.ToLower(word)

	if anExceptions[wordLower] {
		return true
	}
	if aExceptions[wordLower] {
		return false
	}
	if len(word) == 0 {
		return false
	}

	firstChar := unicode.ToLower(rune(word[0]))
	return strings.ContainsRune("aeiou", firstChar)
}
