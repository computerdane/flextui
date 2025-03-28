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
	defer fmt.Print("\033[H\033[2J")
	tui.HandleShellSignals()

	sidebar := components.NewBorders()
	sidebar.Inner.SetIsVertical(true)
	sidebar.SetTitle(" Sidebar ")
	sidebar.SetTitleIsOnBottom(true)
	sidebar.SetBorderSymbols(components.BordersSymbols_Double)
	sidebar.SetColorFunc(color.New(color.FgRed).SprintFunc())
	sidebar.SetTitleColorFunc(color.New(color.Bold).Add(color.FgBlue).SprintFunc())
	tui.Screen.AddChild(sidebar.Outer)

	items := make([]string, 100)
	selectedItem := 0
	for i := range items {
		items[i] = fmt.Sprintf("Menu item %d", i)
	}
	sidebarMenu := components.NewMenu(items)
	sidebarMenu.SetSelectedColorFunc(color.New(color.BgYellow).Add(color.FgBlack).SprintFunc())
	sidebarMenu.AddSelection(selectedItem)
	// sidebarMenu.SetIsVertical(false)
	sidebar.Inner.AddChild(sidebarMenu.Outer)

	mainArea := components.NewBorders()
	mainArea.Inner.SetIsVertical(true)
	mainArea.Outer.SetGrow(3)
	tui.Screen.AddChild(mainArea.Outer)

	mainContent := tui.NewComponent()
	mainArea.Inner.AddChild(mainContent)
	spacer := tui.NewComponent()
	spacer.SetColorFunc(color.New(color.BgBlue).SprintFunc())
	mainArea.Inner.AddChild(spacer)

	themesMenuArea := tui.NewComponent()
	themesMenuArea.SetLength(1)
	mainArea.Inner.AddChild(themesMenuArea)

	themesMenu := components.NewMenu([]string{" [1] Dark Theme ", " [2] Light Theme ", " [3] Epic Theme "})
	themesMenu.SetIsVertical(false)
	themesMenu.SetSelectedColorFunc(color.New(color.Bold).Add(color.FgMagenta).SprintFunc())
	themesMenu.AddSelection(0)
	themesMenuArea.AddChild(tui.NewComponent())
	themesMenuArea.AddChild(themesMenu.Outer)
	themesMenuArea.AddChild(tui.NewComponent())

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
			sidebarMenu.RemoveSelection(selectedItem)
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
			sidebarMenu.AddSelection(selectedItem)

			mainContent.SetContent(strings.Repeat(fmt.Sprintf("You have selected menu item %d %s\n\n", selectedItem, strings.Repeat("-", selectedItem)), selectedItem+1))

			go sidebarMenu.RenderChanges()
			go mainArea.Inner.Render()

			continue
		}

		if char == 'h' || char == 'l' {
			if char == 'h' {
				mainArea.Outer.SetGrow(mainArea.Outer.Grow() + 0.1)
			} else {
				mainArea.Outer.SetGrow(mainArea.Outer.Grow() - 0.1)
			}

			tui.Screen.UpdateLayout()
			go tui.Screen.Render()

			continue
		}

		if char == '1' || char == '2' || char == '3' {
			themesMenu.RemoveAllSelections()
			themesMenu.AddSelection(int(char - '1'))

			switch char {
			case '1':
				mainContent.SetColorFunc(nil)
			case '2':
				mainContent.SetColorFunc(color.New(color.BgWhite).Add(color.FgBlack).SprintFunc())
			case '3':
				mainContent.SetColorFunc(color.New(color.BgGreen).Add(color.FgBlack).SprintFunc())
			}

			go themesMenu.RenderChanges()
			go mainContent.Render()

			continue
		}

	}
}
