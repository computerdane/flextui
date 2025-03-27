package components

import (
	"strings"

	"github.com/computerdane/flextui"
)

type BordersSymbols struct {
	tl string
	tr string
	bl string
	br string
	v  string
	h  string
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

type Borders struct {
	Outer *flextui.Component
	Inner *flextui.Component

	midSection   *flextui.Component
	innerWrapper *flextui.Component

	top    *flextui.Component
	bottom *flextui.Component
	left   *flextui.Component
	right  *flextui.Component

	titleLeft  *flextui.Component
	title      *flextui.Component
	titleRight *flextui.Component

	titleIsOnBottom bool

	symbols *BordersSymbols
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

	b.top.SetLength(1)
	b.bottom.SetLength(1)
	b.left.SetLength(1)
	b.right.SetLength(1)
	b.title.SetGrow(0)

	b.Outer.AddChild(b.left)
	b.Outer.AddChild(b.midSection)
	b.Outer.AddChild(b.right)

	b.midSection.AddChild(b.top)
	b.midSection.AddChild(b.innerWrapper)
	b.midSection.AddChild(b.bottom)

	b.top.AddChild(b.titleLeft)
	b.top.AddChild(b.title)
	b.top.AddChild(b.titleRight)

	b.innerWrapper.AddChild(b.Inner)

	b.updateContentFuncs()

	return &b
}

func (b *Borders) horizontalBorderSection(box *flextui.Box) string {
	return strings.Repeat(b.symbols.h, max(0, box.Width()))
}

func (b *Borders) verticalBorderSection(box *flextui.Box, isLeft bool) string {
	middle := strings.Repeat(b.symbols.v, max(box.Height()-2))
	if isLeft {
		return b.symbols.tl + middle + b.symbols.bl
	} else {
		return b.symbols.tr + middle + b.symbols.br
	}
}

func (b *Borders) updateContentFuncs() {
	b.titleLeft.SetContentFunc(b.horizontalBorderSection)
	b.titleRight.SetContentFunc(b.horizontalBorderSection)
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

func (b *Borders) SetTitle(title string) {
	b.title.SetContent(title)
	b.title.SetLength(len(title))
}

func (b *Borders) SetTitleIsOnBottom(titleIsOnBottom bool) {
	if b.titleIsOnBottom != titleIsOnBottom {
		b.top, b.bottom = b.bottom, b.top
		b.midSection.RemoveAllChildren()
		b.midSection.AddChild(b.top)
		b.midSection.AddChild(b.innerWrapper)
		b.midSection.AddChild(b.bottom)
	}
	b.titleIsOnBottom = titleIsOnBottom
	b.updateContentFuncs()
}

func (b *Borders) SetSymbols(symbols *BordersSymbols) {
	b.symbols = symbols
	b.updateContentFuncs()
}

func (b *Borders) SetColorFunc(colorFunc func(a ...any) string) {
	b.top.SetColorFunc(colorFunc)
	b.bottom.SetColorFunc(colorFunc)
	b.left.SetColorFunc(colorFunc)
	b.right.SetColorFunc(colorFunc)
	b.titleLeft.SetColorFunc(colorFunc)
	b.titleRight.SetColorFunc(colorFunc)
}

func (b *Borders) SetTitleColorFunc(colorFunc func(a ...any) string) {
	b.title.SetColorFunc(colorFunc)
}
