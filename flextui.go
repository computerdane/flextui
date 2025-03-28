package flextui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
)

var components map[uuid.UUID]*Component // Look up components quickly using their key
var Screen *Component                   // Parent of all components

func init() {
	components = make(map[uuid.UUID]*Component)
	Screen = NewComponent()
}

func GetComponentByKey(key uuid.UUID) *Component {
	return components[key]
}

func HideCursor() {
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	fmt.Print("\033[?25h")
}

// Handles SIGINT, SIGTERM, and SIGWINCH signals.
//
// - SIGINT/SIGTERM : shows the cursor and exits the current process
//
// - SIGWINCH : updates the screen layout and re-renders the whole screen
func HandleShellSignals() {
	stopChan := make(chan os.Signal, 1)
	resizeChan := make(chan os.Signal, 1)

	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(resizeChan, syscall.SIGWINCH)

	go func() {
		<-stopChan
		ShowCursor()
		os.Exit(0)
	}()

	go func() {
		for {
			<-resizeChan
			Screen.UpdateLayout()
			Screen.Render()
		}
	}()
}
