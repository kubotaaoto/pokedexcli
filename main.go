package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Pokedex > ")
		scanned := scanner.Scan()
		if !scanned {
			fmt.Printf("scan unsuccessful")
		}
		input := scanner.Text()
		cleanedInput := cleanInput(input)
		firstWord := cleanedInput[0]

		cmd, err := lookupCommand(firstWord)
		if err != nil {
			fmt.Println(err)
			continue
		}
		args := []string{}
		if len(cleanedInput) > 1 {
			args = cleanedInput[1:]
		}

		err = cmd.callback(&cf, args...)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
