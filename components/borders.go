package components

import (
	"strings"
	"sync"

	"github.com/computerdane/flextui"
)

// Define characters that the borders are made out of.
type BordersSymbols struct {
	tl string // top left corner
	tr string // top right corner
	bl string // bottom left corner
	br string // bottom right corner
	v  string // vertical bars
	h  string // horizontal bars
}

var BordersSymbols_Default = &BordersSymbols{
	tl: "┌",
	tr: "┐",
	bl: "└",
	br: "┘",
	v:  "│",
	h:  "─",
}

var BordersSymbols_Double = &BordersSymbols{
	tl: "╔",
	tr: "╗",
	bl: "╚",
	br: "╝",
	v:  "║",
	h:  "═",
}

// A Component that is surrounded by borders. Borders.Outer is the top-level
// parent component that should be added to the screen. Borders.Inner is
// the Component inside the borders that you should fill with content or
// other Components.
type Borders struct {
	Outer *flextui.Component // Outer-most parent component
	Inner *flextui.Component // Inner-most child component

	// Contains a vertical layout with the top border, the innerWrapper, and the
	// bottom border.
	midSection *flextui.Component

	// Wrapper Component around the Inner Component
	innerWrapper *flextui.Component

	top    *flextui.Component
	bottom *flextui.Component
	left   *flextui.Component
	right  *flextui.Component

	titleLeft  *flextui.Component // Border to the left of the title
	title      *flextui.Component // Component containing a title
	titleRight *flextui.Component // Border to the right of the title

	titleIsOnBottom bool

	symbols *BordersSymbols

	mu sync.Mutex
}

func NewBorders() *Borders {
	var b Borders
	b.symbols = BordersSymbols_Default

	b.Outer = flextui.NewComponent()
	b.Inner = flextui.NewComponent()
	b.midSection = flextui.NewComponent()
	b.innerWrapper = flextui.NewComponent()
	b.top = flextui.NewComponent()
	b.bottom = flextui.NewComponent()
	b.left = flextui.NewComponent()
	b.right = flextui.NewComponent()
	b.titleLeft = flextui.NewComponent()
	b.title = flextui.NewComponent()
	b.titleRight = flextui.NewComponent()

	b.midSection.SetIsVertical(true)

	// Set the length of all the border components to 1
	b.top.SetLength(1)
	b.bottom.SetLength(1)
	b.left.SetLength(1)
	b.right.SetLength(1)

	// By default, don't display the title
	b.title.SetGrow(0)

	// Horizontally lay out the left border, the midSection, and the right border
	b.Outer.AddChild(b.left)
	b.Outer.AddChild(b.midSection)
	b.Outer.AddChild(b.right)

	// Vertically lay out the top border, the innerWrapper, and the bottom border
	b.midSection.AddChild(b.top)
	b.midSection.AddChild(b.innerWrapper)
	b.midSection.AddChild(b.bottom)

	// By default, the title is drawn on the top border
	b.top.AddChild(b.titleLeft)
	b.top.AddChild(b.title)
	b.top.AddChild(b.titleRight)

	b.innerWrapper.AddChild(b.Inner)

	b.updateContentFuncs()

	return &b
}

// ContentFunc to render a horizontal bar
func (b *Borders) horizontalBorderSection(box *flextui.Box) string {
	return strings.Repeat(b.symbols.h, max(0, box.Width()))
}

// ContentFunc to render a vertical bar with corners
func (b *Borders) verticalBorderSection(box *flextui.Box, isLeft bool) string {
	middle := strings.Repeat(b.symbols.v, max(0, box.Height()-2))
	if isLeft {
		return b.symbols.tl + middle + b.symbols.bl
	} else {
		return b.symbols.tr + middle + b.symbols.br
	}
}

// Update all of the ContentFuncs for the borders depending on the current
// Borders properties.
func (b *Borders) updateContentFuncs() {
	b.titleLeft.SetContentFunc(b.horizontalBorderSection)
	b.titleRight.SetContentFunc(b.horizontalBorderSection)

	// Render a normal horizontal border where the title isn't
	if b.titleIsOnBottom {
		b.top.SetContentFunc(b.horizontalBorderSection)
		b.bottom.SetContentFunc(nil)
	} else {
		b.top.SetContentFunc(nil)
		b.bottom.SetContentFunc(b.horizontalBorderSection)
	}

	b.left.SetContentFunc(func(box *flextui.Box) string { return b.verticalBorderSection(box, true) })
	b.right.SetContentFunc(func(box *flextui.Box) string { return b.verticalBorderSection(box, false) })
}

// Set the title to be displayed on the top/bottom border.
func (b *Borders) SetTitle(title string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.title.SetContent(title)
	b.title.SetLength(len(title))
}

// Set whether the title is shown on the top border or the bottom border.
func (b *Borders) SetTitleIsOnBottom(titleIsOnBottom bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.titleIsOnBottom != titleIsOnBottom {
		// Swap the top/bottom Components and update their parent's children list
		b.top, b.bottom = b.bottom, b.top
		b.midSection.RemoveAllChildren()
		b.midSection.AddChild(b.top)
		b.midSection.AddChild(b.innerWrapper)
		b.midSection.AddChild(b.bottom)
	}
	b.titleIsOnBottom = titleIsOnBottom
	b.updateContentFuncs()
}

// Set the characters that compose the borders.
func (b *Borders) SetBorderSymbols(symbols *BordersSymbols) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.symbols = symbols
	b.updateContentFuncs()
}

// Set the ColorFunc for the borders. See [flextui.Component.SetColorFunc].
func (b *Borders) SetColorFunc(colorFunc func(a ...any) string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.top.SetColorFunc(colorFunc)
	b.bottom.SetColorFunc(colorFunc)
	b.left.SetColorFunc(colorFunc)
	b.right.SetColorFunc(colorFunc)
	b.titleLeft.SetColorFunc(colorFunc)
	b.titleRight.SetColorFunc(colorFunc)
}

// Set the ColorFunc for the title. See [flextui.Component.SetColorFunc].
func (b *Borders) SetTitleColorFunc(colorFunc func(a ...any) string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.title.SetColorFunc(colorFunc)
}
