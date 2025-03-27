# FlexTUI

FlexTUI is a flexible terminal user interface library inspired by CSS Flexbox. It allows you to create complex layouts with ease, using a simple and intuitive API.

## Features

- Component-Based Architecture: Create reusable UI components with properties like grow, length, and orientation.
- Responsive Layouts: Automatically adjust component sizes and positions based on terminal dimensions.
- Customizable Styles: Apply ANSI color codes for styling components and text.
- Interactive Elements: Build interactive menus and components with keyboard navigation.
- Signal Handling: Automatically handle terminal resize events and clean up on exit.

## Installation

To install FlexTUI, use `go get`:

```
go get github.com/computerdane/flextui
```

## Usage

Here's a simple example to get you started:

```go
package main

import (
	"github.com/computerdane/flextui"
	"github.com/computerdane/flextui/components"
)

func main() {
	// Create a new bordered component
	borders := components.NewBorders()
	borders.SetTitle("My App")
	borders.SetTitleIsOnBottom(false)

	// Create a new component for content
	content := flextui.NewComponent()
	content.SetContent("Hello, FlexTUI!")
	borders.Inner.AddChild(content)

	// Add the bordered component to the screen
	flextui.Screen.AddChild(borders.Outer)

	// Update layout and render
	flextui.Screen.UpdateLayout()
	flextui.Screen.Render()

	// Listen for user input here
	select {}
}
```

## Documentation

For more detailed documentation and examples, please refer to the [GoDoc](https://pkg.go.dev/github.com/computerdane/flextui).

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
