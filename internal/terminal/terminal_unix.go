//go:build !windows

package terminal

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/term"
)

// InsertInput inserts the command into the terminal's input buffer.
// This allows the user to see and edit the command before executing it.
func InsertInput(cmd string) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to just printing the command
		printCommand(cmd)

		return
	}

	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	for _, c := range cmd {
		char := byte(c)
		//nolint:errcheck // TIOCSTI may fail on some systems, we handle by falling back
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(0), syscall.TIOCSTI, uintptr(unsafe.Pointer(&char)))
	}
}

func printCommand(cmd string) {
	_, _ = os.Stdout.WriteString(cmd + "\n")
}
