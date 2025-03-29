package components

import (
	"sync"

	"github.com/computerdane/flextui"
)

var inputWithCursor *Input
var mu sync.Mutex

type Input struct {
	Outer   *flextui.Component
	content *flextui.Component

	hasCursor bool

	mu sync.Mutex
}

func NewInput() *Input {
	var input Input

	input.Outer = flextui.NewComponent()
	input.Outer.SetIsVertical(true)

	input.content = flextui.NewComponent()
	input.content.SetLength(1)
	input.content.SetContent("")
	input.Outer.AddChild(input.content)

	listener := func(c *flextui.Component) {
		input.UpdateCursorPos()
	}
	input.content.AddEventListener(flextui.Event_LayoutUpdated, &listener)

	return &input
}

func (c *Input) Content() string {
	return *c.content.Content()
}

func (c *Input) SetContent(content string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.content.SetContent(content)
	boxWidth := c.Outer.Box().Width()
	if boxWidth > 0 && len(content) >= boxWidth {
		c.content.Scroll.Left = len(content) - boxWidth + 1
		c.content.UpdateLayout()
	} else if c.content.Scroll.Left != 0 {
		c.content.Scroll.Left = 0
		c.content.UpdateLayout()
	}
}

func (c *Input) SetHasCursor(hasCursor bool) {
	mu.Lock()
	c.mu.Lock()
	defer mu.Unlock()
	defer c.mu.Unlock()

	if hasCursor {
		if inputWithCursor != nil {
			inputWithCursor.hasCursor = false
		}
		inputWithCursor = c
		c.hasCursor = true
	} else {
		if inputWithCursor == c {
			inputWithCursor = nil
		}
		c.hasCursor = false
	}
}

func (c *Input) UpdateCursorPos() {
	if c.hasCursor {
		flextui.CursorTo(c.content.Box().Top()+1, min(c.Outer.Box().Right(), c.content.Box().Left()+len(*c.content.Content())+1))
	}
}
