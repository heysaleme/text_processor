package main

import (
	"regexp"
	"strings"
)

var tokenPattern = regexp.MustCompile(`\([^()]*(?:\([^()]*\)[^()]*)*\)|\n|'|[\w]+|\.\.\.|[!?]{2,}|[.,!?:;]`)

func processText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")

	tokens := tokenPattern.FindAllString(text, -1)
	tokens = fixArticles(tokens)
	tokens = applyTransformations(tokens)

	return reconstruct(tokens)
}
