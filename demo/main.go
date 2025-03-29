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
	defer tui.Clear()
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
	sidebarMenu1 := components.NewScrollableMenu(items)
	sidebarMenu1.Menu.SetSelectedColorFunc(color.New(color.BgYellow).Add(color.FgBlack).SprintFunc())
	sidebarMenu1.SetSelectedItem(selectedItem)
	sidebarMenu1.SetIsVertical(false)
	sidebarMenu1Wrapper := tui.NewComponent()
	sidebarMenu1Wrapper.AddChild(sidebarMenu1.Outer)
	sidebarMenu1Wrapper.SetIsVertical(true)
	sidebar.Inner.AddChild(sidebarMenu1Wrapper)

	sidebarMenu2 := components.NewScrollableMenu(items)
	sidebarMenu2.Menu.SetSelectedColorFunc(color.New(color.BgYellow).Add(color.FgBlack).SprintFunc())
	sidebarMenu2.SetSelectedItem(selectedItem)
	sidebar.Inner.AddChild(sidebarMenu2.Outer)

	mainArea := components.NewBorders()
	mainArea.Inner.SetIsVertical(true)
	mainArea.Outer.SetGrow(3)
	tui.Screen.AddChild(mainArea.Outer)

	mainContent := tui.NewComponent()
	mainArea.Inner.AddChild(mainContent)

	inputArea := tui.NewComponent()
	inputArea.SetColorFunc(color.New(color.BgRed).SprintFunc())
	mainArea.Inner.AddChild(inputArea)

	spacer := tui.NewComponent()
	spacer.SetColorFunc(color.New(color.BgBlue).SprintFunc())
	spacer.SetGrow(2)
	inputArea.AddChild(spacer)

	input := components.NewInput()
	input.SetContent("Input Box")
	tui.CursorOwner = input.Outer
	inputArea.AddChild(input.Outer)

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

	insertMode := false

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get key: %s", err)
			break
		}

		if insertMode {
			if key == keyboard.KeyEsc {
				tui.HideCursor()
				insertMode = false
				continue
			}

			if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
				content := input.Content()
				if len(content) > 0 {
					input.SetContent(content[:len(content)-1])
				}
			} else if key == keyboard.KeySpace {
				input.SetContent(input.Content() + " ")
			} else {
				input.SetContent(input.Content() + string(char))
			}

			input.UpdateCursorPos()
			go input.Outer.Render()

			continue
		}

		if char == 'q' || key == keyboard.KeyCtrlC {
			break
		}

		if char == 'i' {
			tui.ShowCursor()
			input.UpdateCursorPos()
			insertMode = true
			continue
		}

		if char == 'j' || char == 'k' {
			if char == 'j' {
				if selectedItem == len(items)-1 {
					continue
				}
				selectedItem++
			} else {
				if selectedItem == 0 {
					continue
				}
				selectedItem--
			}
			sidebarMenu1.SetSelectedItem(selectedItem)
			sidebarMenu2.SetSelectedItem(selectedItem)

			mainContent.SetContent(strings.Repeat(fmt.Sprintf("You have selected menu item %d %s\n\n", selectedItem, strings.Repeat("-", selectedItem)), selectedItem+1))

			go sidebarMenu1.RenderChanges()
			go sidebarMenu2.RenderChanges()
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
