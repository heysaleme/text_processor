package main

import "strings"

func reconstruct(tokens []string) string {
	var sb strings.Builder
	insideQuotes := false

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		if token == "\n" {
			current := sb.String()
			if strings.HasSuffix(current, " ") {
				sb.Reset()
				sb.WriteString(strings.TrimRight(current, " "))
			}
			sb.WriteString("\n")
			continue
		}

		if isPunctuation(token) || token == "..." {
			current := sb.String()
			if strings.HasSuffix(current, " ") {
				sb.Reset()
				sb.WriteString(strings.TrimRight(current, " "))
			}
			sb.WriteString(token)

			if (token == "," || token == "." || token == ":" || token == ";") &&
				i+1 < len(tokens) && isWordToken(tokens[i+1]) {
				sb.WriteString(" ")
			}
			continue
		}

		if token == "'" && i > 0 && i+1 < len(tokens) {
			prevToken := tokens[i-1]
			nextToken := tokens[i+1]
			if isWordToken(prevToken) && isContractionWord(nextToken) {
				sb.WriteString("'")
				i++
				sb.WriteString(tokens[i])
				continue
			}
		}

		if token == "'" {
			if insideQuotes {
				current := sb.String()
				if strings.HasSuffix(current, " ") {
					sb.Reset()
					sb.WriteString(strings.TrimRight(current, " "))
				}
				sb.WriteString("'")
				insideQuotes = false
			} else {
				if sb.Len() > 0 && !strings.HasSuffix(sb.String(), " ") {
					sb.WriteString(" ")
				}
				sb.WriteString("'")
				insideQuotes = true
			}
			continue
		}

		current := sb.String()
		if sb.Len() > 0 &&
			!strings.HasSuffix(current, " ") &&
			!strings.HasSuffix(current, "\n") &&
			(!insideQuotes || (i > 0 && tokens[i-1] != "'")) {
			sb.WriteString(" ")
		}
		sb.WriteString(token)
	}

	return strings.TrimSpace(sb.String())
}

func isContractionWord(word string) bool {
	contractions := map[string]bool{
		"t":     true,
		"s":     true,
		"m":     true,
		"re":    true,
		"ve":    true,
		"ll":    true,
		"d":     true,
		"em":    true,
		"til":   true,
		"bout":  true,
		"cause": true,
		"round": true,
	}

	return contractions[strings.ToLower(word)]
}

func isPunctuation(token string) bool {
	if token == "..." {
		return true
	}

	for _, char := range token {
		if char != '.' && char != ',' && char != '!' && char != '?' && char != ':' && char != ';' {
			return false
		}
	}

	return len(token) > 0
}
