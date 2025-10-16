package main

import (
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run . <input.txt> <output.txt>")
		return
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	data, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Println("Error reading input file:", err)
		return
	}

	text := string(data)
	output := processText(text)

	err = os.WriteFile(outputFile, []byte(output), 0644)
	if err != nil {
		fmt.Println("Error writing output file:", err)
		return
	}

	fmt.Println("✅ Result written to", outputFile)
}

func processText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")

	re := regexp.MustCompile(`\([^)]+\)|\n|'|[\w]+|\.\.\.|[!?]{2,}|[.,!?:;]`)
	tokens := re.FindAllString(text, -1)

	tokens = fixArticles(tokens)
	tokens = applyTransformations(tokens)
	return reconstruct(tokens)
}

func isWordToken(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return true
		}
	}
	return false
}

func applyTransformations(tokens []string) []string {
	result := make([]string, len(tokens))
	copy(result, tokens)

	// Разделяем токены по строкам
	lines := splitTokensByLines(result)

	// Обрабатываем каждую строку отдельно
	var finalTokens []string
	for _, lineTokens := range lines {
		processedLine := processLine(lineTokens)
		finalTokens = append(finalTokens, processedLine...)
	}

	return finalTokens
}

// Разделяет токены по строкам
func splitTokensByLines(tokens []string) [][]string {
	var lines [][]string
	var currentLine []string

	for _, token := range tokens {
		currentLine = append(currentLine, token)
		if token == "\n" {
			lines = append(lines, currentLine)
			currentLine = []string{}
		}
	}

	// Добавляем последнюю строку, если она не пустая
	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}

func processLine(tokens []string) []string {
	result := make([]string, len(tokens))
	copy(result, tokens)

	// Собираем команды в этой строке в порядке появления
	commands := []struct {
		index int
		cmd   string
		count int
	}{}

	for i := 0; i < len(result); i++ {
		t := result[i]
		if strings.HasPrefix(t, "(") && strings.HasSuffix(t, ")") {
			cmd, count := parseCommand(t)
			commands = append(commands, struct {
				index int
				cmd   string
				count int
			}{i, cmd, count})
			// Удаляем команду из результата
			result[i] = ""
		}
	}

	// Применяем команды в порядке появления только к словам в этой строке
	for _, command := range commands {
		applied := 0
		j := command.index - 1

		for applied < command.count && j >= 0 {
			// Если дошли до начала строки - останавливаемся
			if j < 0 {
				break
			}

			// Пропускаем пустые строки (удаленные команды) и не-слова
			if result[j] == "" || !isWordToken(result[j]) {
				j--
				continue
			}

			word := result[j]

			switch strings.ToLower(command.cmd) {
			case "hex":
				bigNum := new(big.Int)
				if _, ok := bigNum.SetString(word, 16); ok {
					result[j] = bigNum.String()
				}
			case "bin":
				bigNum := new(big.Int)
				if _, ok := bigNum.SetString(word, 2); ok {
					result[j] = bigNum.String()
				}
			case "up":
				result[j] = strings.ToUpper(word)
			case "low":
				result[j] = strings.ToLower(word)
			case "cap":
				if len(word) > 0 {
					runes := []rune(word)
					runes[0] = unicode.ToUpper(runes[0])
					for k := 1; k < len(runes); k++ {
						runes[k] = unicode.ToLower(runes[k])
					}
					result[j] = string(runes)
				}
			}
			applied++
			j--
		}
	}

	// Убираем пустые строки (удаленные команды)
	var finalResult []string
	for _, token := range result {
		if token != "" {
			finalResult = append(finalResult, token)
		}
	}

	return finalResult
}

func fixArticles(tokens []string) []string {
	result := make([]string, len(tokens))
	copy(result, tokens)

	vowels := "aeiouAEIOU"

	anExceptions := map[string]bool{
		"hour": true, "honest": true, "heir": true, "honour": true, "honor": true,
	}

	aExceptions := map[string]bool{
		"university": true, "unit": true, "unicorn": true, "user": true, "european": true,
	}

	firstLetter := func(next string) rune {
		for _, r := range next {
			if unicode.IsLetter(r) {
				return r
			}
		}
		return 0
	}

	for i := 0; i < len(result); i++ {
		word := result[i]

		// Пропускаем команды
		if strings.HasPrefix(word, "(") && strings.HasSuffix(word, ")") {
			continue
		}

		// Обрабатываем все варианты артиклей
		if (word == "a" || word == "A" || word == "an" || word == "An") && i+1 < len(result) {
			j := i + 1

			// Ищем следующее слово в той же строке (останавливаемся на \n)
			for j < len(result) && result[j] != "\n" &&
				(isPunctuation(result[j]) || result[j] == "'" ||
					(strings.HasPrefix(result[j], "(") && strings.HasSuffix(result[j], ")")) ||
					strings.TrimSpace(result[j]) == "") {
				j++
			}

			// Если дошли до конца или нашли перевод строки - не меняем артикль
			if j >= len(result) || result[j] == "\n" {
				continue
			}

			next := result[j]

			// Если следующее слово тоже "a" или "an" - не меняем текущий артикль
			if next == "a" || next == "A" || next == "an" || next == "An" {
				continue
			}

			nextLower := strings.ToLower(next)

			fl := firstLetter(next)
			if fl == 0 {
				continue
			}

			// Проверяем исключения
			if anExceptions[nextLower] {
				if word == "a" || word == "A" {
					if word == "A" {
						result[i] = "An"
					} else {
						result[i] = "an"
					}
				}
				continue
			}
			if aExceptions[nextLower] {
				if word == "an" || word == "An" {
					if word == "An" {
						result[i] = "A"
					} else {
						result[i] = "a"
					}
				}
				continue
			}

			// По первой букве - используем переменную vowels
			if strings.ContainsRune(vowels, fl) {
				if word == "a" || word == "A" {
					if word == "A" {
						result[i] = "An"
					} else {
						result[i] = "an"
					}
				}
			} else {
				if word == "an" || word == "An" {
					if word == "An" {
						result[i] = "A"
					} else {
						result[i] = "a"
					}
				}
			}
		}
	}

	return result
}

func reconstruct(tokens []string) string {
	var sb strings.Builder
	insideQuotes := false
	sentenceStart := true

	for i, t := range tokens {
		if t == "\n" {
			s := sb.String()
			if strings.HasSuffix(s, " ") {
				sb.Reset()
				sb.WriteString(strings.TrimRight(s, " "))
			}
			sb.WriteString("\n")
			sentenceStart = true
			continue
		}

		// пунктуация
		if isPunctuation(t) || t == "..." {
			// убрать пробел перед знаком
			s := sb.String()
			if strings.HasSuffix(s, " ") {
				sb.Reset()
				sb.WriteString(strings.TrimRight(s, " "))
			}
			sb.WriteString(t)

			// После пунктуации ставим пробел, если следующий токен - слово
			if i+1 < len(tokens) && isWordToken(tokens[i+1]) {
				sb.WriteString(" ")
			}
			sentenceStart = (t == "." || t == "!" || t == "?" || t == "...")
			continue
		}

		// кавычки
		if t == "'" {
			if insideQuotes {
				s := sb.String()
				if strings.HasSuffix(s, " ") {
					sb.Reset()
					sb.WriteString(strings.TrimRight(s, " "))
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

		// обычное слово
		word := t
		if insideQuotes {
			word = strings.TrimLeft(word, " ")
		}

		// Если это начало предложения и слово - строчный артикль, делаем первую букву заглавной
		if sentenceStart && (word == "a" || word == "an") {
			if word == "a" {
				word = "A"
			} else if word == "an" {
				word = "An"
			}
		}

		// Добавляем пробел между словами, но не если предыдущий токен был пунктуацией (кроме некоторых случаев)
		if sb.Len() > 0 && !strings.HasSuffix(sb.String(), " ") {
			// Проверяем предыдущий токен
			if i > 0 {
				prevToken := tokens[i-1]
				// Если предыдущий токен - пунктуация, которая требует пробела перед словом
				if isPunctuation(prevToken) && prevToken != ":" && prevToken != ";" {
					sb.WriteString(" ")
				} else if !isPunctuation(prevToken) && prevToken != "'" && prevToken != "\n" {
					// Если предыдущий токен - слово, ставим пробел
					sb.WriteString(" ")
				}
			} else if !insideQuotes || (insideQuotes && i > 0 && tokens[i-1] != "'") {
				sb.WriteString(" ")
			}
		}

		sb.WriteString(word)
		sentenceStart = false
	}

	return strings.TrimSpace(sb.String())
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

func isPunctuation(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Если строка состоит из нескольких одинаковых знаков пунктуации
	if len(s) > 1 {
		firstChar := rune(s[0])
		for _, char := range s {
			if char != firstChar {
				return false
			}
		}
		return isSinglePunctuation(string(firstChar))
	}
	return isSinglePunctuation(s)
}

func isSinglePunctuation(s string) bool {
	return s == "." || s == "," || s == "!" || s == "?" || s == ":" || s == ";"
}
