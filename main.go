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

func isWordToken(s string) bool {
	// Слова с апострофами тоже считаем словами
	if strings.Contains(s, "'") {
		for _, r := range s {
			if r != '\'' && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
		return len(s) > 0
	}

	// Обычные слова
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return len(s) > 0
}

func applyTransformations(tokens []string) []string {
	result := make([]string, len(tokens))
	copy(result, tokens)

	// Сначала объединяем команды которые были разбиты
	result = combineCommandTokens(result)

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

// Объединяет токены типа ["Up", "(low)"] в ["(Up (low))"] но НЕ объединяет слова с командами
func combineCommandTokens(tokens []string) []string {
	var result []string
	i := 0

	for i < len(tokens) {
		// Если текущий токен - команда (уже в скобках), просто добавляем
		if strings.HasPrefix(tokens[i], "(") && strings.HasSuffix(tokens[i], ")") {
			result = append(result, tokens[i])
			i++
			continue
		}

		// Если текущий токен - слово, а следующий - команда в скобках
		// НО только если слово выглядит как команда (up, low, cap, etc)
		if i+1 < len(tokens) &&
			isWordToken(tokens[i]) &&
			strings.HasPrefix(tokens[i+1], "(") &&
			strings.HasSuffix(tokens[i+1], ")") &&
			isValidCommand(tokens[i]) { // ВАЖНО: проверяем что это команда!

			// Объединяем в одну команду: "Up" + "(low)" -> "(Up (low))"
			combined := "(" + tokens[i] + " " + tokens[i+1][1:] // убираем первую скобку у второй команды
			result = append(result, combined)
			i += 2 // Пропускаем два токена
		} else {
			result = append(result, tokens[i])
			i++
		}
	}

	fmt.Printf("🔧 COMBINED TOKENS: %v\n", result)
	return result
}
func isValidCommand(cmd string) bool {
	cmd = strings.ToLower(cmd)
	// Проверяем как команды в скобках "(up)", так и слова "up"
	cmd = strings.Trim(cmd, "()")
	cmd = strings.ToLower(cmd)
	return cmd == "up" || cmd == "low" || cmd == "cap" || cmd == "hex" || cmd == "bin"
}

func processLine(tokens []string) []string {
	result := make([]string, len(tokens))
	copy(result, tokens)

	fmt.Printf("🚀 START processLine: %v\n", tokens)

	// МНОГОПРОХОДНАЯ обработка вложенных команд - от самых глубоких к внешним
	maxPasses := 5
	for pass := 0; pass < maxPasses; pass++ {
		fmt.Printf("🔄 PASS %d\n", pass+1)
		changed := false
		for i := 0; i < len(result); i++ {
			t := result[i]
			if strings.HasPrefix(t, "(") && strings.HasSuffix(t, ")") {
				depth := getBracketDepth(t)
				if depth > 1 {
					fmt.Printf("  Processing token[%d]: %s (depth: %d)\n", i, t, depth)
					processed := processNestedCommand(t)
					if processed != t {
						result[i] = processed
						changed = true
						fmt.Printf("  CHANGED: %s -> %s\n", t, processed)
					}
				}
			}
		}
		if !changed {
			fmt.Printf("  No changes in pass %d, breaking\n", pass+1)
			break
		}
	}

	// Обрабатываем обычные команды (без вложенности)
	commands := []struct {
		index int
		cmd   string
		count int
	}{}

	for i := 0; i < len(result); i++ {
		t := result[i]
		if strings.HasPrefix(t, "(") && strings.HasSuffix(t, ")") && getBracketDepth(t) == 1 {
			cmd, count := parseCommand(t)
			if isValidCommand(cmd) {
				commands = append(commands, struct {
					index int
					cmd   string
					count int
				}{i, cmd, count})
			}
		}
	}

	// Применяем обычные команды
	for _, command := range commands {
		applied := 0
		j := command.index - 1

		for applied < command.count && j >= 0 {
			if j < 0 || result[j] == "\n" {
				break
			}

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

		result[command.index] = ""
	}

	var finalResult []string
	for _, token := range result {
		if token != "" {
			finalResult = append(finalResult, token)
		}
	}

	return finalResult
}

// Функция для определения глубины вложенности скобок
func getBracketDepth(token string) int {
	depth := 0
	maxDepth := 0
	for _, char := range token {
		if char == '(' {
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		} else if char == ')' {
			depth--
		}
	}
	return maxDepth
}

func processNestedCommand(token string) string {
	fmt.Printf("🔍 processNestedCommand INPUT: %s\n", token)

	content := token[1 : len(token)-1]

	// Ищем самую внутреннюю команду
	innerStart := -1
	innerEnd := -1
	maxDepth := 0
	currentDepth := 0

	for i, char := range content {
		if char == '(' {
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
				innerStart = i
			}
		} else if char == ')' {
			if currentDepth == maxDepth && innerStart != -1 {
				innerEnd = i
				break
			}
			currentDepth--
		}
	}

	fmt.Printf("   innerStart: %d, innerEnd: %d\n", innerStart, innerEnd)

	// Если нашли внутреннюю команду
	if innerStart != -1 && innerEnd != -1 {
		innerCommand := content[innerStart : innerEnd+1]
		innerCmd, _ := parseCommand(innerCommand)
		fmt.Printf("   Found inner command: %s -> %s\n", innerCommand, innerCmd)

		if isValidCommand(innerCmd) {
			// Ищем команду ПЕРЕД внутренней командой (внешнюю команду)
			beforeCommand := strings.TrimSpace(content[:innerStart])
			fmt.Printf("   beforeCommand: '%s'\n", beforeCommand)

			if beforeCommand != "" {
				// Применяем внутреннюю команду к внешней команде
				transformed := applyCommandToWord(beforeCommand, innerCmd)
				fmt.Printf("   transformed: '%s'\n", transformed)

				// Собираем новый контент
				newContent := transformed + content[innerEnd+1:]
				result := "(" + strings.TrimSpace(newContent) + ")"
				fmt.Printf("   OUTPUT: %s\n", result)
				return result
			}

			// Ищем команду ПОСЛЕ внутренней команды
			afterCommand := strings.TrimSpace(content[innerEnd+1:])
			fmt.Printf("   afterCommand: '%s'\n", afterCommand)

			if afterCommand != "" {
				// Применяем внутреннюю команду к команде после
				transformed := applyCommandToWord(afterCommand, innerCmd)
				fmt.Printf("   transformed: '%s'\n", transformed)

				// Собираем новый контент
				newContent := content[:innerStart] + transformed
				result := "(" + strings.TrimSpace(newContent) + ")"
				fmt.Printf("   OUTPUT: %s\n", result)
				return result
			}
		}
	}

	fmt.Printf("   NO CHANGES, OUTPUT: %s\n", token)
	return token
}

// Применяет команду к слову (уже существует)
func applyCommandToWord(word string, cmd string) string {
	fmt.Printf("   applyCommandToWord: '%s' with cmd '%s'\n", word, cmd)

	// Если word - это команда (в скобках), парсим её
	if strings.HasPrefix(word, "(") && strings.HasSuffix(word, ")") {
		cmdContent := word[1 : len(word)-1]
		parsedCmd, _ := parseCommand("(" + cmdContent + ")")
		fmt.Printf("   It's a command, parsedCmd: '%s'\n", parsedCmd)

		// Применяем команду cmd к parsedCmd
		switch strings.ToLower(cmd) {
		case "up":
			result := "(" + strings.ToUpper(parsedCmd) + ")"
			fmt.Printf("   Result: %s\n", result)
			return result
		case "low":
			result := "(" + strings.ToLower(parsedCmd) + ")"
			fmt.Printf("   Result: %s\n", result)
			return result
		case "cap":
			if len(parsedCmd) > 0 {
				runes := []rune(parsedCmd)
				runes[0] = unicode.ToUpper(runes[0])
				for k := 1; k < len(runes); k++ {
					runes[k] = unicode.ToLower(runes[k])
				}
				result := "(" + string(runes) + ")"
				fmt.Printf("   Result: %s\n", result)
				return result
			}
		}
		return word
	}

	// Обычное применение команды к слову
	switch strings.ToLower(cmd) {
	case "up":
		return strings.ToUpper(word)
	case "low":
		return strings.ToLower(word)
	case "cap":
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for k := 1; k < len(runes); k++ {
				runes[k] = unicode.ToLower(runes[k])
			}
			return string(runes)
		}
		return word
	default:
		return word
	}
}

func fixArticles(tokens []string) []string {
	result := make([]string, len(tokens))
	copy(result, tokens)

	for i := 0; i < len(result); i++ {
		word := result[i]

		// Пропускаем команды
		if strings.HasPrefix(word, "(") && strings.HasSuffix(word, ")") {
			continue
		}

		// Обрабатываем артикли ВНЕ зависимости от кавычек
		if (word == "a" || word == "A" || word == "an" || word == "An") && i+1 < len(result) {
			j := i + 1

			// Ищем следующее слово
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

			// Пропускаем стоп-слова
			if isStopWord(next) {
				continue
			}

			// Определяем правильный артикль
			if shouldUseAn(next) {
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

func isStopWord(word string) bool {
	stopwords := map[string]bool{
		"and": true, "or": true, "the": true, "a": true, "an": true, "of": true, "for": true,
	}
	return stopwords[strings.ToLower(word)]
}

func shouldUseAn(word string) bool {
	anExceptions := map[string]bool{
		"hour": true, "honest": true, "heir": true, "honour": true, "honor": true,
	}
	aExceptions := map[string]bool{
		"university": true, "unit": true, "unicorn": true, "user": true, "european": true,
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

func processText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Улучшенное регулярное выражение для вложенных скобок
	re := regexp.MustCompile(`\([^()]*(?:\([^()]*\)[^()]*)*\)|\n|'|[\w]+|\.\.\.|[!?]{2,}|[.,!?:;]`)
	tokens := re.FindAllString(text, -1)

	fmt.Printf("📝 TOKENS: %v\n", tokens)

	tokens = fixArticles(tokens)
	tokens = applyTransformations(tokens)
	return reconstruct(tokens)
}

func reconstruct(tokens []string) string {
	var sb strings.Builder
	insideQuotes := false

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		if t == "\n" {
			s := sb.String()
			if strings.HasSuffix(s, " ") {
				sb.Reset()
				sb.WriteString(strings.TrimRight(s, " "))
			}
			sb.WriteString("\n")
			continue
		}

		// пунктуация
		if isPunctuation(t) || t == "..." {
			s := sb.String()
			if strings.HasSuffix(s, " ") {
				sb.Reset()
				sb.WriteString(strings.TrimRight(s, " "))
			}
			sb.WriteString(t)

			if (t == "," || t == "." || t == ":" || t == ";") &&
				i+1 < len(tokens) && isWordToken(tokens[i+1]) {
				sb.WriteString(" ")
			}
			continue
		}

		// Обрабатываем апострофы для сокращений
		if t == "'" && i > 0 && i+1 < len(tokens) {
			prevToken := tokens[i-1]
			nextToken := tokens[i+1]

			// СПЕЦИАЛЬНЫЕ СЛУЧАИ СОКРАЩЕНИЙ
			if isWordToken(prevToken) && isContractionWord(nextToken) {
				// Это сокращение - объединяем без пробелов
				sb.WriteString("'")
				i++
				sb.WriteString(tokens[i])
				continue
			}
		}

		// кавычки
		if t == "'" {
			if insideQuotes {
				// закрывающая кавычка - убираем пробел перед ней
				s := sb.String()
				if strings.HasSuffix(s, " ") {
					sb.Reset()
					sb.WriteString(strings.TrimRight(s, " "))
				}
				sb.WriteString("'")
				insideQuotes = false
			} else {
				// открывающая кавычка - пробел перед ней
				if sb.Len() > 0 && !strings.HasSuffix(sb.String(), " ") {
					sb.WriteString(" ")
				}
				sb.WriteString("'")
				insideQuotes = true
			}
			continue
		}

		// обычное слово
		if sb.Len() > 0 && !strings.HasSuffix(sb.String(), " ") && (!insideQuotes || (insideQuotes && i > 0 && tokens[i-1] != "'")) {
			sb.WriteString(" ")
		}
		sb.WriteString(t)
	}

	return strings.TrimSpace(sb.String())
}

// Функция для определения слов-сокращений
func isContractionWord(word string) bool {
	contractions := map[string]bool{
		// Основные сокращения
		"t": true, "s": true, "m": true, "re": true, "ve": true, "ll": true, "d": true,
		// Дополнительные распространенные сокращения
		"em": true, "til": true, "bout": true, "cause": true, "round": true,
	}
	return contractions[strings.ToLower(word)]
}

func parseCommand(token string) (cmd string, count int) {
	token = strings.Trim(token, "()")
	token = strings.Join(strings.Fields(token), " ")

	parts := strings.Split(token, ",")
	cmd = strings.TrimSpace(parts[0])
	count = 1

	if len(parts) > 1 {
		countStr := strings.TrimSpace(parts[1])
		// Если параметр содержит скобки - это вложенная команда, игнорируем
		if strings.Contains(countStr, "(") || strings.Contains(countStr, ")") {
			// Для случаев типа (up, 10(bin)) - используем count = 1
			count = 1
		} else if c, err := strconv.Atoi(countStr); err == nil {
			count = c
		}
	}
	return
}

func isPunctuation(s string) bool {
	if s == "..." {
		return true
	}
	for _, char := range s {
		if char != '.' && char != ',' && char != '!' && char != '?' && char != ':' && char != ';' {
			return false
		}
	}
	return len(s) > 0
}

// Функция для разделения строк по строкам (отсутствует в вашем коде)
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

	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}
