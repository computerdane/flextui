package components

import (
	"sync"

	"github.com/computerdane/flextui"
)

type Menu struct {
	Outer *flextui.Component

	selectedIndices map[int]struct{}

	colorFunc         func(a ...any) string
	selectedColorFunc func(a ...any) string

	renderQueue map[string]struct{}

	mu sync.Mutex
}

func NewMenu(items []string) *Menu {
	m := Menu{
		selectedIndices: make(map[int]struct{}),
		renderQueue:     make(map[string]struct{}),
	}

	m.Outer = flextui.NewComponent()
	m.Outer.SetIsVertical(true)

	for _, item := range items {
		c := flextui.NewComponent()
		c.SetContent(item)
		c.SetLength(1)
		m.Outer.AddChild(c)
	}
	m.Outer.AddChild(flextui.NewComponent()) // spacer

	m.Outer.SetLength(len(items))

	return &m
}

func (m *Menu) enqueue(c *flextui.Component) {
	m.renderQueue[c.Key()] = struct{}{}
}

func (m *Menu) items() []*flextui.Component {
	children := m.Outer.Children()
	return children[:len(children)-1]
}

func (m *Menu) SetIsVertical(isVertical bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Outer.IsVertical() != isVertical {
		items := m.items()
		if isVertical {
			for _, c := range items {
				c.SetLength(1)
			}
			m.Outer.SetLength(len(items))
		} else {
			sum := 0
			for _, c := range items {
				l := len(*c.Content())
				c.SetLength(l)
				sum += l
			}
			m.Outer.SetLength(sum)
		}
		m.Outer.SetIsVertical(isVertical)
	}
}

func (m *Menu) SetColorFunc(colorFunc func(a ...any) string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.colorFunc = colorFunc
	for i, c := range m.items() {
		if _, exists := m.selectedIndices[i]; !exists {
			c.SetColorFunc(colorFunc)
			m.enqueue(c)
		}
	}
}

func (m *Menu) SetSelectedColorFunc(colorFunc func(a ...any) string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.selectedColorFunc = colorFunc
	for i, c := range m.items() {
		if _, exists := m.selectedIndices[i]; exists {
			c.SetColorFunc(colorFunc)
			m.enqueue(c)
		}
	}
}

func (m *Menu) AddSelection(index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.selectedIndices[index] = struct{}{}
	c := m.items()[index]
	c.SetColorFunc(m.selectedColorFunc)
	m.enqueue(c)
}

func (m *Menu) RemoveSelection(index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.selectedIndices, index)
	c := m.items()[index]
	c.SetColorFunc(m.colorFunc)
	m.enqueue(c)
}

func (m *Menu) RemoveAllSelections() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.selectedIndices {
		c := m.items()[i]
		c.SetColorFunc(m.colorFunc)
		m.enqueue(c)
	}
	m.selectedIndices = make(map[int]struct{})
}

func (m *Menu) RenderChanges() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key := range m.renderQueue {
		c := flextui.GetComponentByKey(key)
		if c != nil {
			go c.Render()
		}
	}
	m.renderQueue = make(map[string]struct{})
}
