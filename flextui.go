package flextui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
)

var components map[string]*Component // Look up components quickly using their key
var Screen *Component                // Parent of all components

func init() {
	components = make(map[string]*Component)
	Screen = NewComponent()
}

func NewComponent() *Component {
	c := &Component{
		key:  uuid.NewString(),
		grow: 1,
	}
	components[c.key] = c
	return c
}

func HideCursor() {
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	fmt.Print("\033[?25h")
}

func HandleSignals() {
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
			Screen.Update()
			Screen.Render()
		}
	}()
}
