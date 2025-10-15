package main

import (
	"fmt"
	"os"
	"regexp"
		"math/big"

	"strconv"
	"strings"
	"unicode"

)

func main() {
	input := strings.Join(os.Args[1:], " ")
	output := processText(input)
	fmt.Println(output)
}

func processText(text string) string {
	// Токенизация
re := regexp.MustCompile(`'|[\w]+|\.\.\.|[!?]{2,}|[.,!?:;]|\(\w+(?:,\s*\d+)?\)`)
	tokens := re.FindAllString(text, -1)



	tokens = applyTransformations(tokens)
	tokens = fixArticles(tokens)
	return reconstruct(tokens)
}

func applyTransformations(tokens []string) []string {
	var result []string

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		// ищем флажки
		if strings.HasPrefix(t, "(") && strings.HasSuffix(t, ")") {
			cmd, count := parseCommand(t)

			for j := 1; j <= count && len(result)-j >= 0; j++ {
				index := len(result) - j
				word := result[index]

				switch cmd {
				case "hex":
					//big.Int вместо ParseInt (для длинных чисел)
					bigNum := new(big.Int)
					if _, ok := bigNum.SetString(word, 16); ok {
						result[index] = bigNum.String()
					}

				case "bin":
					bigNum := new(big.Int)
					if _, ok := bigNum.SetString(word, 2); ok {
						result[index] = bigNum.String()
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

	vowels := "aeiouAEIOU"

	anExceptions := map[string]bool{
		"hour": true, "honest": true, "heir": true, "honour": true, "honor": true,
	}

	aExceptions := map[string]bool{
		"university": true, "unit": true, "unicorn": true, "user": true, "european": true,
	}

	// стоп-слова
	stopwords := map[string]bool{
		"and": true, "or": true, "the": true, "a": true, "an": true, "of": true, "for": true,
	}

	firstLetter := func(next string) rune {
		for _, r := range next {
			if unicode.IsLetter(r) {
				return r
			}
		}
		return 0
	}

	for i := 0; i < len(tokens); i++ {
		word := tokens[i]

		// только при "a" или "an" 
		if (strings.EqualFold(word, "a") || strings.EqualFold(word, "an")) && i+1 < len(tokens) {
			// ищем следующий значащий токен (пропуская пунктуацию и кавычки)
			j := i + 1
			for j < len(tokens) && (isPunctuation(tokens[j]) || tokens[j] == "'" || strings.TrimSpace(tokens[j]) == "") {
				j++
			}
			if j >= len(tokens) {
				// дальше ничего — не меняем
				result = append(result, word)
				continue
			}

			next := tokens[j]
			nextLower := strings.ToLower(next)

			if stopwords[nextLower] {
				result = append(result, word)
				continue
			}

			// возьмём первую букву следующего значащего токена
			fl := firstLetter(nextLower)
			if fl == 0 {
				result = append(result, word)
				continue
			}

			// сначала проверим явные исключения по слову целиком
			if anExceptions[nextLower] {
				// должно быть "an"
				if word == "A" || word == "An" {
					result = append(result, "An")
				} else {
					result = append(result, "an")
				}
				continue
			}
			if aExceptions[nextLower] {
				// должно быть "a"
				if word == "A" || word == "An" {
					result = append(result, "A")
				} else {
					result = append(result, "a")
				}
				continue
			}

			// по первой букве: если гласная -> an, иначе a
			if strings.ContainsRune(vowels, fl) {
				if word == "A" || word == "An" {
					result = append(result, "An")
				} else {
					result = append(result, "an")
				}
			} else {
				if word == "A" || word == "An" {
					result = append(result, "A")
				} else {
					result = append(result, "a")
				}
			}
			continue
		}

		// во всех остальных случаях просто копируем токен
		result = append(result, word)
	}

	return result
}




func reconstruct(tokens []string) string {
	var sb strings.Builder
	insideQuotes := false

	for i, t := range tokens {
		// пунктуация
        if isPunctuation(t) || t == "..." {
            // убрать пробел перед знаком
            s := sb.String()
            if strings.HasSuffix(s, " ") {
                sb.Reset()
                sb.WriteString(strings.TrimRight(s, " "))
            }
            sb.WriteString(t)
            // пробел после, если следующий — слово
            if i+1 < len(tokens) {
                nxt := tokens[i+1]
                if !isPunctuation(nxt) && nxt != "..." && nxt != "'" {
                    sb.WriteString(" ")
                }
            }
            continue
        }

        // кавычки
        if t == "'" {
            if insideQuotes {
                // закрывающая кавычка — убрать пробел перед ней
                s := sb.String()
                if strings.HasSuffix(s, " ") {
                    sb.Reset()
                    sb.WriteString(strings.TrimRight(s, " "))
                }
                sb.WriteString("'")
                insideQuotes = false
            } else {
                // открывающая кавычка — пробел перед ней, если нужно
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
            // убираем пробел в начале слова после открывающей кавычки
            word = strings.TrimLeft(word, " ")
        }

        // пробел между словами (кроме сразу после открывающей кавычки)
        if sb.Len() > 0 && !strings.HasSuffix(sb.String(), " ") && (!insideQuotes || (insideQuotes && i > 0 && tokens[i-1] != "'")) {
            sb.WriteString(" ")
        }

        sb.WriteString(word)
	}

	return strings.TrimSpace(sb.String())
}


func isPunctuation(s string) bool {
	return s == "." || s == "," || s == "!" || s == "?" || s == ":" || s == ";"
}
