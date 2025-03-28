package components

import (
	"sync"

	"github.com/computerdane/flextui"
)

type ScrollableMenu struct {
	Outer *flextui.Component
	Menu  *Menu

	selectedItem int

	needsRender bool

	scrollToSelectedItem func(*flextui.Component)

	mu sync.Mutex
}

func NewScrollableMenu(items []string) *ScrollableMenu {
	sm := ScrollableMenu{Menu: NewMenu(items)}

	sm.scrollToSelectedItem = func(parent *flextui.Component) {
		if sm.Outer.IsVertical() {
			if sm.selectedItem >= sm.Menu.Outer.Scroll.Top+parent.Box().Height() {
				sm.Menu.Outer.Scroll.Top = sm.selectedItem - parent.Box().Height() + 1
				sm.Menu.Outer.UpdateLayout()
				sm.needsRender = true
			} else if sm.selectedItem < sm.Menu.Outer.Scroll.Top {
				sm.Menu.Outer.Scroll.Top = sm.selectedItem
				sm.Menu.Outer.UpdateLayout()
				sm.needsRender = true
			}
		} else {
			items := sm.Menu.Outer.Children()
			width := 0
			for i := 0; i <= sm.selectedItem; i++ {
				width += items[i].Length()
			}
			if width >= sm.Menu.Outer.Scroll.Left+parent.Box().Width() {
				sm.Menu.Outer.Scroll.Left = width - parent.Box().Width()
				sm.Menu.Outer.UpdateLayout()
				sm.needsRender = true
			} else if width-items[sm.selectedItem].Length() < sm.Menu.Outer.Scroll.Left {
				sm.Menu.Outer.Scroll.Left = width - items[sm.selectedItem].Length()
				sm.Menu.Outer.UpdateLayout()
				sm.needsRender = true
			}
		}
	}

	sm.Outer = flextui.NewComponent()
	sm.Outer.SetIsVertical(true)
	sm.Outer.AddEventListener(flextui.Event_LayoutUpdated, &sm.scrollToSelectedItem)
	sm.Outer.AddChild(sm.Menu.Outer)

	return &sm
}

func (sm *ScrollableMenu) SelectedItem() int {
	return sm.selectedItem
}

func (sm *ScrollableMenu) SetSelectedItem(i int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.Menu.RemoveSelection(sm.selectedItem)
	sm.selectedItem = i
	sm.Menu.AddSelection(sm.selectedItem)

	sm.scrollToSelectedItem(sm.Outer)
}

func (sm *ScrollableMenu) SetIsVertical(isVertical bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.Outer.SetIsVertical(isVertical)
	sm.Menu.SetIsVertical(isVertical)
	sm.Outer.SetLength(sm.Menu.Outer.Length())
}

func (sm *ScrollableMenu) RenderChanges() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.needsRender {
		go sm.Outer.Render()
		sm.needsRender = false
		return
	}

	go sm.Menu.RenderChanges()
}
