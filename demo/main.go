package main

import (
	"fmt"
	"os"
	"strings"

	tui "github.com/computerdane/flextui"
	"github.com/computerdane/flextui/components"
	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
)

func main() {
	keyboard.Open()
	defer keyboard.Close()

	tui.HideCursor()
	defer tui.ShowCursor()
	tui.HandleShellSignals()

	sidebar := components.NewBorders()
	sidebar.Inner.SetIsVertical(true)
	sidebar.SetTitle(" Sidebar ")
	sidebar.SetTitleIsOnBottom(true)
	sidebar.SetBorderSymbols(components.BordersSymbols_Double)
	sidebar.SetColorFunc(color.New(color.FgRed).SprintFunc())
	sidebar.SetTitleColorFunc(color.New(color.Bold).Add(color.FgBlue).SprintFunc())
	tui.Screen.AddChild(sidebar.Outer)

	items := make([]*tui.Component, 10)
	for i := range items {
		items[i] = tui.NewComponent()
		items[i].SetContent(fmt.Sprintf("Menu item %d", i))
		items[i].SetLength(1)
		sidebar.Inner.AddChild(items[i])
	}
	sidebar.Inner.AddChild(tui.NewComponent())
	selectedItem := 0
	selectedItemStyle := color.New(color.BgYellow).Add(color.FgBlack).SprintFunc()
	items[selectedItem].SetColorFunc(selectedItemStyle)

	pane := components.NewBorders()
	pane.Inner.SetIsVertical(true)
	pane.Outer.SetGrow(3)
	tui.Screen.AddChild(pane.Outer)

	tui.Screen.UpdateLayout()
	tui.Screen.Render()

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get key: %s", err)
			break
		}

		if char == 'q' || key == keyboard.KeyCtrlC {
			break
		}

		if char == 'j' || char == 'k' {
			items[selectedItem].SetColorFunc(nil)
			items[selectedItem].Render()
			if char == 'j' {
				selectedItem++
			} else {
				selectedItem--
			}
			if selectedItem < 0 {
				selectedItem = 0
			} else if selectedItem >= len(items) {
				selectedItem = len(items) - 1
			}
			items[selectedItem].SetColorFunc(selectedItemStyle)
			items[selectedItem].Render()

			pane.Inner.SetContent(strings.Repeat(fmt.Sprintf("You have selected menu item %d %s\n", selectedItem, strings.Repeat("-", selectedItem)), selectedItem+1))
			pane.Inner.Render()

			continue
		}

		if char == 'h' || char == 'l' {
			if char == 'h' {
				pane.Outer.SetGrow(pane.Outer.Grow() + 0.1)
			} else {
				pane.Outer.SetGrow(pane.Outer.Grow() - 0.1)
			}
			tui.Screen.UpdateLayout()
			tui.Screen.Render()
		}

	}
}
