package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	//"unicode"
)

func main() {
	input := strings.Join(os.Args[1:], " ")

	output := processText(input)
	fmt.Println(output)
}

func processText(text string) string {
	// Токенизация: разбиваем строку на слова, команды и знаки препинания
	re := regexp.MustCompile(`'\s*[^']*?\s*'|\w+|\.\.\.|[!?]{2,}|[.,!?:;]|\(\w+(?:,\s*\d+)?\)`)
	tokens := re.FindAllString(text, -1)

	// Очистка кавычек от пробелов внутри
	for i, t := range tokens {
		if strings.HasPrefix(t, "'") && strings.HasSuffix(t, "'") {
			tokens[i] = "'" + strings.TrimSpace(t[1:len(t)-1]) + "'"
		}
	}

	tokens = applyTransformations(tokens)
	tokens = fixArticles(tokens)
	return reconstruct(tokens)
}

func applyTransformations(tokens []string) []string {
	var result []string

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		// Команда?
		if strings.HasPrefix(t, "(") && strings.HasSuffix(t, ")") {
			// Разбираем (cmd) или (cmd, N)
			cmd, count := parseCommand(t)
			if count > len(result) {
				count = len(result)
			}

			for j := 1; j <= count; j++ {
				index := len(result) - j
				word := result[index]

				switch cmd {
				case "hex":
					if num, err := strconv.ParseInt(word, 16, 64); err == nil {
						result[index] = fmt.Sprintf("%d", num)
					}
				case "bin":
					if num, err := strconv.ParseInt(word, 2, 64); err == nil {
						result[index] = fmt.Sprintf("%d", num)
					}
				case "up":
					result[index] = strings.ToUpper(word)
				case "low":
					result[index] = strings.ToLower(word)
				case "cap":
					if len(word) > 0 {
						result[index] = strings.ToUpper(string(word[0])) + word[1:]
					}
				}
			}
		} else {
			result = append(result, t)
		}
	}

	return result
}

func parseCommand(token string) (cmd string, count int) {
	token = strings.Trim(token, "()")
	parts := strings.Split(token, ",")
	cmd = strings.TrimSpace(parts[0])
	count = 1
	if len(parts) > 1 {
		if c, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			count = c
		}
	}
	return
}

func fixArticles(tokens []string) []string {
	result := []string{}
	vowels := "aeiouhAEIOUH"

	for i := 0; i < len(tokens); i++ {
		if strings.ToLower(tokens[i]) == "a" && i+1 < len(tokens) {
			next := tokens[i+1]
			if len(next) > 0 && strings.ContainsRune(vowels, rune(next[0])) {
				if tokens[i] == "A" {
					result = append(result, "An")
				} else {
					result = append(result, "an")
				}
				continue
			}
		}
		result = append(result, tokens[i])
	}
	return result
}

func reconstruct(tokens []string) string {
	var sb strings.Builder

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		// Спецзнаки без пробела перед ними
		if isPunctuation(t) {
			sb.WriteString(t)
			if i+1 < len(tokens) && !isPunctuation(tokens[i+1]) {
				sb.WriteString(" ")
			}
			continue
		}

		// Точки и !?, например ...
		if t == "..." || strings.HasPrefix(t, "!!") || strings.HasPrefix(t, "!?") {
			sb.WriteString(t)
			sb.WriteString(" ")
			continue
		}

		// Апострофы
		if strings.HasPrefix(t, "'") && strings.HasSuffix(t, "'") {
			sb.WriteString(t)
			sb.WriteString(" ")
			continue
		}

		// Обычное слово
		if sb.Len() > 0 && !strings.HasSuffix(sb.String(), " ") {
			sb.WriteString(" ")
		}
		sb.WriteString(t)
	}

	return strings.TrimSpace(sb.String())
}

func isPunctuation(s string) bool {
	return s == "." || s == "," || s == "!" || s == "?" || s == ":" || s == ";"
}
