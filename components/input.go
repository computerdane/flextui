package components

import (
	"sync"

	"github.com/computerdane/flextui"
)

type Input struct {
	Outer   *flextui.Component
	content *flextui.Component

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

	cursorListener := func(c *flextui.Component) {
		input.UpdateCursorPos()
	}
	input.content.AddEventListener(flextui.Event_LayoutUpdated, &cursorListener)

	scrollListener := func(c *flextui.Component) {
		input.updateScrollPos()
	}
	input.Outer.AddEventListener(flextui.Event_LayoutUpdated, &scrollListener)

	return &input
}

func (c *Input) updateScrollPos() {
	content := *c.content.Content()
	boxWidth := c.Outer.Box().Width()
	if boxWidth > 0 && len(content) >= boxWidth {
		c.content.Scroll.Left = len(content) - boxWidth + 1
		c.content.UpdateLayout()
	} else if c.content.Scroll.Left != 0 {
		c.content.Scroll.Left = 0
		c.content.UpdateLayout()
	}
}

func (c *Input) Content() string {
	return *c.content.Content()
}

func (c *Input) SetContent(content string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.content.SetContent(content)
	c.updateScrollPos()
}

func (c *Input) SetColorFunc(colorFunc func(a ...any) string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.content.SetColorFunc(colorFunc)
}

func (c *Input) UpdateCursorPos() {
	if flextui.CursorOwner == c.Outer {
		flextui.CursorTo(c.content.Box().Top()+1, min(c.Outer.Box().Right(), c.content.Box().Left()+len(*c.content.Content())+1))
	}
}
