package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]

	var arr []string

	for i := 0; i < len(args); i++ {
		if args[i] == " " && word != " " {
			arr = append(arr, word)
			continue
		} else  {
			word += args[i]
		}
	}

	fmt.Println(arr)
}
