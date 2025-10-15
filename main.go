package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

)

func main() {
	// Получаем все аргументы командной строки, начиная с первого (первый аргумент — это сам ввод)
	input := strings.Join(os.Args[1:], " ")
	output := processText(input)
	fmt.Println(output)
}

func processText(text string) string {
	// Токенизация: разбиваем строку на слова, команды и знаки препинания
re := regexp.MustCompile(`'|[\w]+|\.\.\.|[!?]{2,}|[.,!?:;]|\(\w+(?:,\s*\d+)?\)`)
	tokens := re.FindAllString(text, -1)



	tokens = applyTransformations(tokens)
	tokens = fixArticles(tokens)
	return reconstruct(tokens)
}

func applyTransformations(tokens []string) []string {
	var result []string

	// Ищем команду и применяем к нужному числу слов
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		// Если нашли команду в скобках
		if strings.HasPrefix(t, "(") && strings.HasSuffix(t, ")") {
			// Разбираем команду
			cmd, count := parseCommand(t)
			// Применяем команду только к нужным словам в результате
			// Применяем к последним словам в result
			for j := 1; j <= count && len(result)-j >= 0; j++ {
				index := len(result) - j
				word := result[index]

				// Применяем команду
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

	// базовый набор "гласных" по первой букве
	vowels := "aeiouAEIOU"

	// фонетические исключения, требующие "an" несмотря на первую букву (hour, honest...)
	anExceptions := map[string]bool{
		"hour": true, "honest": true, "heir": true, "honour": true, "honor": true,
	}

	// фонетические исключения, требующие "a" несмотря на первую гласную букву (university, user...)
	aExceptions := map[string]bool{
		"university": true, "unit": true, "unicorn": true, "user": true, "european": true,
	}

	// стоп-слова, после которых не меняем артикль
	stopwords := map[string]bool{
		"and": true, "or": true, "the": true, "a": true, "an": true, "of": true, "for": true,
	}

	// helper: вернуть первое буквенное rune в next (или 0)
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
		//lowerWord := strings.ToLower(word)

		// обрабатываем только если текущий токен — "a" или "an" (любого регистра)
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

			// если следующее — стопслово, не трогаем
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
