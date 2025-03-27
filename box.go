package flextui

import "fmt"

type Box struct {
	top    int
	left   int
	right  int
	bottom int
}

func (b *Box) Top() int {
	return b.top
}

func (b *Box) Left() int {
	return b.left
}

func (b *Box) Bottom() int {
	return b.bottom
}

func (b *Box) Right() int {
	return b.right
}

func (b *Box) Width() int {
	return b.right - b.left
}

func (b *Box) Height() int {
	return b.bottom - b.top
}

func (b *Box) ToString() string {
	return fmt.Sprintf("[T: %d, L: %d, B: %d, R: %d]", b.top, b.left, b.bottom, b.right)
}
