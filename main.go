package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Shout converts the input text to uppercase
func Shout(input string) string {
	return strings.ToUpper(input)
}

func main() {
	fmt.Println("Enter text to shout (press Ctrl+D to finish):")
	
	scanner := bufio.NewScanner(os.Stdin)
	
	// For single-line input
	if scanner.Scan() {
		input := scanner.Text()
		fmt.Println("\nShouting:")
		fmt.Println(Shout(input))
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}

