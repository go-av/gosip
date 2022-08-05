package log

import "fmt"

var enablePrintMSG = false

func EnablePrintMSG(enable bool) {
	enablePrintMSG = enable
}

func PrintMSG(title string, msg string) {
	if enablePrintMSG {
		fmt.Printf("\n                    ===%s ↓ ====\n", title)
		fmt.Print(msg)
		fmt.Printf("\n                    ===%s ↑ ====\n", title)
	}
}
