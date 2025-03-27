package flextui

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/term"
)

// A Component represents a rectangular area on the screen that can have a
// parent Component and children Components. Components are laid out according
// to simple rules inspired by CSS Flex. By default, Components lay out their
// children horizontally and space them evenly. When properties are changed
// on this Component, parent Components, or child Components, they will update
// each others properties in order to lay themselves out correctly. After
// changing any properties of a Component, the UpdateLayout() function must be
// called to apply them before the next Render().
type Component struct {
	key        string
	box        Box
	isVertical bool
	parent     *Component
	children   []*Component

	content   content
	colorFunc func(a ...any) string

	grow            float64
	childrenGrowSum float64

	length            int
	childrenLengthSum int

	mu sync.Mutex
}

func NewComponent() *Component {
	c := &Component{
		key:  uuid.NewString(),
		grow: 1,
	}
	components[c.key] = c
	return c
}

func (c *Component) Box() *Box {
	return &c.box
}

func (c *Component) Content() *string {
	return c.content.value
}

func (c *Component) Grow() float64 {
	return c.grow
}

func (c *Component) IsVertical() bool {
	return c.isVertical
}

func (c *Component) Key() string {
	return c.key
}

func (c *Component) Children() []*Component {
	return c.children
}

// Change whether child Components are laid out vertically or horizontally.
func (c *Component) SetIsVertical(isVertical bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isVertical = isVertical
}

// Set this Component's text content.
func (c *Component) SetContent(content string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.content.setValue(&content)
}

// Set the Component's content based on its Box's dimensions. Useful for
// creating responsive Components that fill their content depending on the
// width/height of the Component.
func (c *Component) SetContentFunc(updateFunc func(*Box) string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.content.updateFunc = updateFunc
}

// Set the Component's style using a function that can be called to add ANSI
// color codes before rendering the Component's content. Pairs well with the
// library github.com/fatih/color using a color's SprintFunc().
func (c *Component) SetColorFunc(colorFunc func(a ...any) string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.colorFunc = colorFunc
}

// Set the Component's grow property. All Components have a default grow of 1.
// If a Component's grow is larger than others, it will take up more space
// proportional to the total grow of its neighbors.
func (c *Component) SetGrow(grow float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.parent != nil {
		c.parent.childrenGrowSum += grow - c.grow
	}
	c.grow = grow
}

// Set a custom length for a Component. Overrides the grow property and disables
// flex-based layout for this Component only. Neighbor Components will still use
// their grow properties for their layouts.
func (c *Component) SetLength(length int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.parent != nil {
		c.parent.childrenLengthSum += length - c.length
		if c.length == 0 && length > 0 {
			c.parent.childrenGrowSum -= c.grow
		}
	}
	c.length = length
}

// Removes all child Components from this Component.
func (c *Component) RemoveAllChildren() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.childrenGrowSum = 0
	c.childrenLengthSum = 0
	c.children = nil
}

// Adds a child Component to this Component. The order in which AddChild() is
// called will determine the order of the child Components' layout.
func (c *Component) AddChild(child *Component) {
	c.mu.Lock()
	defer c.mu.Unlock()

	child.parent = c
	if c.children == nil {
		c.children = []*Component{}
	}
	c.children = append(c.children, child)
	if child.length == 0 {
		c.childrenGrowSum += child.grow
	}
	c.childrenLengthSum += child.length
}

// Updates the Box positions of this Component and all child Components.
//
// Useful for responding to layout changes triggered by screen resizing or user
// actions.
func (c *Component) UpdateLayout() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.key == Screen.key {
		// The Screen Component should always fit the terminal size
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			fmt.Println("Error getting terminal size: ", err)
			return
		}
		Screen.box.top = 0
		Screen.box.left = 0
		Screen.box.bottom = height
		Screen.box.right = width
	} else if c.parent != nil {
		// All other Components use a flex layout based on the parent's box
		nChildren := len(c.parent.children)
		width := c.parent.box.Width()
		height := c.parent.box.Height()
		if c.parent.isVertical {
			height = int(float64(height-c.parent.childrenLengthSum) / (c.parent.childrenGrowSum / c.grow))
		} else {
			width = int(float64(width-c.parent.childrenLengthSum) / (c.parent.childrenGrowSum / c.grow))
		}

		if c.parent.children[0].key == c.key {
			// The first child should have the same top/left as the parent
			c.box.top = c.parent.box.top
			c.box.left = c.parent.box.left
		} else {
			// The rest of the children should align the top/left of their box with the previous child's box
			for i, prev := range c.parent.children[:nChildren-1] {
				if c.parent.children[i+1].key == c.key {
					if c.parent.isVertical {
						c.box.top = prev.box.bottom
						c.box.left = c.parent.box.left
					} else {
						c.box.top = c.parent.box.top
						c.box.left = prev.box.right
					}
					break
				}
			}
		}
		if c.parent.children[nChildren-1].key == c.key {
			// The last child should always align its bottom/right with the parent
			c.box.bottom = c.parent.box.bottom
			c.box.right = c.parent.box.right
		} else {
			// Ensure we use the parent box's position for the cross-axis, and use the computed length if c.length is not provided
			if c.parent.isVertical {
				if c.length > 0 {
					c.box.bottom = c.box.top + c.length
				} else {
					c.box.bottom = c.box.top + height
				}
				c.box.right = c.parent.box.right
			} else {
				c.box.bottom = c.parent.box.bottom
				if c.length > 0 {
					c.box.right = c.box.left + c.length
				} else {
					c.box.right = c.box.left + width
				}
			}
		}
	}

	// Update content according to contentFunc
	if c.content.updateFunc != nil {
		value := c.content.updateFunc(&c.box)
		c.content.setValue(&value)
	}

	// Recursively update all children
	if c.children != nil {
		for _, child := range c.children {
			child.UpdateLayout()
		}
	}
}

// Render this Component's content to the screen, and render all child
// Components as well.
func (c *Component) Render() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Use a string builder so we don't flood stdout with print calls
	var builder strings.Builder

	var write func(string)
	if c.colorFunc != nil {
		write = func(s string) { builder.WriteString(c.colorFunc(s)) }
	} else {
		write = func(s string) { builder.WriteString(s) }
	}

	builder.WriteString(fmt.Sprintf("\033[%d;%dH", c.box.top, c.box.left))

	// Construct our output line-by-line
	width := c.box.Width()
	var a, b int
	for row := 0; row < c.box.bottom-c.box.top; row++ {
		// Position the cursor at the location of the current row
		builder.WriteString(fmt.Sprintf("\033[%d;%dH", c.box.top+row+1, c.box.left+1))

		// We will print a substring of c.content from a:b
		b = a + width

		// The number of spaces to print
		spaces := width

		// Print content if it exists, and print spaces where there isn't content
		if c.content.value != nil {
			contentLen := c.content.displayLen()
			if a < contentLen {
				substr := ""
				if b <= contentLen {
					substr = c.content.displaySubstring(a, b)
				} else {
					substr = c.content.displaySubstring(a, contentLen)
				}
				newlineIndex := strings.Index(substr, "\n")
				if newlineIndex != -1 {
					write(substr[:newlineIndex])
					spaces = b - a - newlineIndex
					a += newlineIndex + 1
				} else {
					write(substr)
					spaces = b - contentLen
					a = b
				}
			}
		}
		if spaces > 0 {
			write(strings.Repeat(" ", spaces))
		}
	}

	// Output to stdout
	fmt.Print(builder.String())

	// Recursively render all children
	if c.children != nil {
		for _, child := range c.children {
			child.Render()
		}
	}
}
