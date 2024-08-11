package term

import (
	"fmt"
	"os"
	"runtime"

	"golang.org/x/term"
)

// PrintLine prints a line to stdout if and only if stdout is a terminal.
func PrintLine() {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			if runtime.GOOS == "windows" {
				for i := 0; i < width; i++ {
					fmt.Print("-")
				}
				fmt.Println()
			} else {
				fmt.Print("\u001B[90m")
				for i := 0; i < width; i++ {
					fmt.Print("─")
				}
				fmt.Println("\u001B[0m")
			}
		}
	}
}

// PrintGreen prints green text to stdout if stdout is a terminal, otherwise plain text.
func PrintGreen(text string) {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Printf("\x1B[1;32m%s\x1B[0m", text)
	} else {
		fmt.Print(text)
	}
}

// PrintRed prints red text to stdout if stdout is a terminal, otherwise plain text.
func PrintRed(text string) {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Printf("\x1B[1;31m%s\x1B[0m", text)
	} else {
		fmt.Print(text)
	}
}

// PrintGrey prints grey text to stdout if stdout is a terminal, otherwise plain text.
func PrintGrey(text string) {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Printf("\x1B[1;90m%s\x1B[0m", text)
	} else {
		fmt.Print(text)
	}
}
