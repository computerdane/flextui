package components

import "github.com/computerdane/flextui"

type Menu struct {
	Outer *flextui.Component

	selectedIndices map[int]struct{}

	colorFunc         func(a ...any) string
	selectedColorFunc func(a ...any) string

	renderQueue map[string]struct{}
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

	return &m
}

func (m *Menu) enqueue(c *flextui.Component) {
	m.renderQueue[c.Key()] = struct{}{}
}

func (m *Menu) SetColorFunc(colorFunc func(a ...any) string) {
	m.colorFunc = colorFunc
	for i, c := range m.Outer.Children() {
		if _, exists := m.selectedIndices[i]; !exists {
			c.SetColorFunc(colorFunc)
			m.enqueue(c)
		}
	}
}

func (m *Menu) SetSelectedColorFunc(colorFunc func(a ...any) string) {
	m.selectedColorFunc = colorFunc
	for i, c := range m.Outer.Children() {
		if _, exists := m.selectedIndices[i]; exists {
			c.SetColorFunc(colorFunc)
			m.enqueue(c)
		}
	}
}

func (m *Menu) AddSelection(index int) {
	m.selectedIndices[index] = struct{}{}
	c := m.Outer.Children()[index]
	c.SetColorFunc(m.selectedColorFunc)
	m.enqueue(c)
}

func (m *Menu) RemoveSelection(index int) {
	delete(m.selectedIndices, index)
	c := m.Outer.Children()[index]
	c.SetColorFunc(m.colorFunc)
	m.enqueue(c)
}

func (m *Menu) RemoveAllSelections() {
	for i := range m.selectedIndices {
		c := m.Outer.Children()[i]
		c.SetColorFunc(m.colorFunc)
		m.enqueue(c)
	}
	m.selectedIndices = make(map[int]struct{})
}

func (m *Menu) RenderChanges() {
	for key := range m.renderQueue {
		c := flextui.GetComponentByKey(key)
		if c != nil {
			go c.Render()
		}
	}
	m.renderQueue = make(map[string]struct{})
}
