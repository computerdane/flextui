package flextui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var Screen *Component // Parent of all components

var cursorHidden bool
var cursorRow int
var cursorCol int

func init() {
	Screen = NewComponent()
}

func HideCursor() {
	cursorHidden = true
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	cursorHidden = false
	fmt.Print("\033[?25h")
}

func CursorTo(row, col int) {
	cursorRow = row
	cursorCol = col
	fmt.Printf("\033[%d;%dH", row, col)
}

func Clear() {
	fmt.Print("\033[H\033[2J")
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
		Clear()
		os.Exit(0)
	}()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			<-resizeChan
			cancel()
			ctx, cancel = context.WithCancel(context.Background())
			go func() {
				select {
				case <-time.After(100 * time.Millisecond):
					Screen.UpdateLayout()
					Screen.Render()
				case <-ctx.Done():
				}
			}()
		}
	}()
}
