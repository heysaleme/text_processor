package main

import (
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

func applyTransformations(tokens []string) []string {
	result := make([]string, len(tokens))
	copy(result, tokens)

	result = combineCommandTokens(result)

	lines := splitTokensByLines(result)

	var finalTokens []string
	for _, lineTokens := range lines {
		processedLine := processLine(lineTokens)
		finalTokens = append(finalTokens, processedLine...)
	}

	return finalTokens
}

func combineCommandTokens(tokens []string) []string {
	var result []string
	i := 0

	for i < len(tokens) {
		if strings.HasPrefix(tokens[i], "(") && strings.HasSuffix(tokens[i], ")") {
			result = append(result, tokens[i])
			i++
			continue
		}

		if i+1 < len(tokens) &&
			isWordToken(tokens[i]) &&
			strings.HasPrefix(tokens[i+1], "(") &&
			strings.HasSuffix(tokens[i+1], ")") &&
			isValidCommand(tokens[i]) {
			combined := "(" + tokens[i] + " " + tokens[i+1][1:]
			result = append(result, combined)
			i += 2
			continue
		}

		result = append(result, tokens[i])
		i++
	}

	return result
}

func isValidCommand(cmd string) bool {
	cmd = strings.Trim(strings.ToLower(cmd), "()")
	return cmd == "up" || cmd == "low" || cmd == "cap" || cmd == "hex" || cmd == "bin"
}

func processLine(tokens []string) []string {
	result := make([]string, len(tokens))
	copy(result, tokens)

	maxPasses := 5
	for pass := 0; pass < maxPasses; pass++ {
		changed := false

		for i, token := range result {
			if strings.HasPrefix(token, "(") && strings.HasSuffix(token, ")") {
				if getBracketDepth(token) > 1 {
					processed := processNestedCommand(token)
					if processed != token {
						result[i] = processed
						changed = true
					}
				}
			}
		}

		if !changed {
			break
		}
	}

	type commandSpec struct {
		index int
		cmd   string
		count int
	}

	var commands []commandSpec
	for i, token := range result {
		if strings.HasPrefix(token, "(") && strings.HasSuffix(token, ")") && getBracketDepth(token) == 1 {
			cmd, count := parseCommand(token)
			if isValidCommand(cmd) {
				commands = append(commands, commandSpec{index: i, cmd: cmd, count: count})
			}
		}
	}

	for _, command := range commands {
		applied := 0
		j := command.index - 1

		for applied < command.count && j >= 0 {
			if result[j] == "\n" {
				break
			}

			if result[j] == "" || !isWordToken(result[j]) {
				j--
				continue
			}

			switch strings.ToLower(command.cmd) {
			case "hex":
				if value, ok := new(big.Int).SetString(result[j], 16); ok {
					result[j] = value.String()
				}
			case "bin":
				if value, ok := new(big.Int).SetString(result[j], 2); ok {
					result[j] = value.String()
				}
			case "up":
				result[j] = strings.ToUpper(result[j])
			case "low":
				result[j] = strings.ToLower(result[j])
			case "cap":
				result[j] = capitalizeWord(result[j])
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

func getBracketDepth(token string) int {
	depth := 0
	maxDepth := 0

	for _, char := range token {
		if char == '(' {
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
			continue
		}

		if char == ')' {
			depth--
		}
	}

	return maxDepth
}

func processNestedCommand(token string) string {
	content := token[1 : len(token)-1]

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
			continue
		}

		if char == ')' {
			if currentDepth == maxDepth && innerStart != -1 {
				innerEnd = i
				break
			}
			currentDepth--
		}
	}

	if innerStart == -1 || innerEnd == -1 {
		return token
	}

	innerCommand := content[innerStart : innerEnd+1]
	innerCmd, _ := parseCommand(innerCommand)
	if !isValidCommand(innerCmd) {
		return token
	}

	beforeCommand := strings.TrimSpace(content[:innerStart])
	if beforeCommand != "" {
		transformed := applyCommandToWord(beforeCommand, innerCmd)
		newContent := transformed + content[innerEnd+1:]
		return "(" + strings.TrimSpace(newContent) + ")"
	}

	afterCommand := strings.TrimSpace(content[innerEnd+1:])
	if afterCommand != "" {
		transformed := applyCommandToWord(afterCommand, innerCmd)
		newContent := content[:innerStart] + transformed
		return "(" + strings.TrimSpace(newContent) + ")"
	}

	return token
}

func applyCommandToWord(word string, cmd string) string {
	if strings.HasPrefix(word, "(") && strings.HasSuffix(word, ")") {
		parsedCmd, _ := parseCommand(word)

		switch strings.ToLower(cmd) {
		case "up":
			return "(" + strings.ToUpper(parsedCmd) + ")"
		case "low":
			return "(" + strings.ToLower(parsedCmd) + ")"
		case "cap":
			return "(" + capitalizeWord(parsedCmd) + ")"
		default:
			return word
		}
	}

	switch strings.ToLower(cmd) {
	case "up":
		return strings.ToUpper(word)
	case "low":
		return strings.ToLower(word)
	case "cap":
		return capitalizeWord(word)
	default:
		return word
	}
}

func parseCommand(token string) (cmd string, count int) {
	token = strings.Trim(token, "()")
	token = strings.Join(strings.Fields(token), " ")

	parts := strings.Split(token, ",")
	cmd = strings.TrimSpace(parts[0])
	count = 1

	if len(parts) > 1 {
		countStr := strings.TrimSpace(parts[1])
		if strings.Contains(countStr, "(") || strings.Contains(countStr, ")") {
			return cmd, count
		}

		if parsedCount, err := strconv.Atoi(countStr); err == nil {
			count = parsedCount
		}
	}

	return cmd, count
}

func splitTokensByLines(tokens []string) [][]string {
	var lines [][]string
	var currentLine []string

	for _, token := range tokens {
		currentLine = append(currentLine, token)
		if token == "\n" {
			lines = append(lines, currentLine)
			currentLine = nil
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}

func capitalizeWord(word string) string {
	if len(word) == 0 {
		return word
	}

	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}
