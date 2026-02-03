//go:build windows

package terminal

import (
	"fmt"
)

// InsertInput on Windows prints the command since TIOCSTI is not available.
func InsertInput(cmd string) {
	fmt.Println(cmd)
}
