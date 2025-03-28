package flextui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
)

const BLANK_CHAR = " "

const (
	Event_LayoutUpdated = iota // Triggered at the very end of UpdateLayout()
)

// A Component represents a rectangular area on the screen that can have a
// parent Component and children Components. Components are laid out according
// to simple rules inspired by CSS Flex. By default, Components lay out their
// children horizontally and space them evenly. When properties are changed on
// this Component, parent Components, or child Components, they will update
// each others properties in order to lay themselves out correctly. After
// changing any properties of a Component, the UpdateLayout() function must be
// called to apply them before the next Render().
//
// For pre-built Components that are more advanced, see the
// [github.com/computerdane/flextui/components] library.
type Component struct {
	Scroll Scroll

	box        Box
	isVertical bool
	parent     *Component
	children   []*Component

	firstChild *Component
	lastChild  *Component

	prevNeighbor *Component
	nextNeighbor *Component

	hasFlexChild bool

	content   content
	colorFunc func(a ...any) string

	grow            float64
	childrenGrowSum float64

	length            int
	childrenLengthSum int

	firstBlankRow     int
	firstBlankColumns []int

	cancelRender context.CancelFunc

	eventListeners map[int][]*func(*Component)

	mu sync.Mutex
}

func NewComponent() *Component {
	c := &Component{
		grow:           1,
		firstBlankRow:  -1,
		eventListeners: make(map[int][]*func(*Component)),
	}
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

func (c *Component) Children() []*Component {
	return c.children
}

func (c *Component) Length() int {
	return c.length
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

	// Ensure that we re-render blank content with the new color
	c.firstBlankRow = -1
	c.firstBlankColumns = nil
	c.colorFunc = colorFunc
}

// Set the Component's grow property. All Components have a default grow of 1.
// If a Component's grow is larger than others, it will take up more space
// proportional to the total grow of its neighbors.
func (c *Component) SetGrow(grow float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.parent != nil {
		c.parent.mu.Lock()
		defer c.parent.mu.Unlock()

		c.parent.childrenGrowSum += grow - c.grow
	}
	c.grow = grow
}

// Set a custom length for a Component. Overrides the grow property and
// disables flex-based layout for this Component only. Neighbor Components will
// still use their grow properties for their layouts.
func (c *Component) SetLength(length int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.parent != nil {
		c.parent.mu.Lock()
		defer c.parent.mu.Unlock()

		c.parent.childrenLengthSum += length - c.length
		if c.length == 0 && length > 0 {
			c.parent.hasFlexChild = true
			c.parent.childrenGrowSum -= c.grow
		} else if c.length > 0 && length == 0 {
			hasFlexChild := false
			for _, child := range c.parent.children {
				if child == c {
					continue
				}
				if child.length == 0 {
					hasFlexChild = true
				}
			}
			c.parent.hasFlexChild = hasFlexChild
		}
	}
	c.length = length
}

// Removes all child Components from this Component.
func (c *Component) RemoveAllChildren() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.firstChild = nil
	c.lastChild = nil
	c.hasFlexChild = false
	c.childrenGrowSum = 0
	c.childrenLengthSum = 0
	c.children = nil
}

// Adds a child Component to this Component. The order in which AddChild() is
// called will determine the order of the child Components' layout.
func (c *Component) AddChild(child *Component) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.firstChild == nil {
		c.firstChild = child
	}
	if c.lastChild != nil {
		child.prevNeighbor = c.lastChild
		c.lastChild.nextNeighbor = child
	}
	c.lastChild = child
	child.parent = c
	c.children = append(c.children, child)
	if child.length == 0 {
		c.hasFlexChild = true
		c.childrenGrowSum += child.grow
	}
	c.childrenLengthSum += child.length
}

// Attach an event listener to this component. Use the flextui.Event_*
// constants to choose an event.
func (c *Component) AddEventListener(event int, listener *func(*Component)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eventListeners[event] = append(c.eventListeners[event], listener)
}

// Remove an event listener from this component. Use the flextui.Event_*
// constants to choose an event.
func (c *Component) RemoveEventListener(event int, listener *func(*Component)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, e := range c.eventListeners[event] {
		if e == listener {
			c.eventListeners[event] = append(c.eventListeners[event][:i], c.eventListeners[event][i+1:]...)
			return
		}
	}
}

// Updates the Box positions of this Component and all child Components.
//
// Useful for responding to layout changes triggered by screen resizing or user
// actions.
func (c *Component) UpdateLayout() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c == Screen {
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
			height = max(0, int(float64(height-c.parent.childrenLengthSum)/(c.parent.childrenGrowSum/c.grow)))
		} else {
			width = max(0, int(float64(width-c.parent.childrenLengthSum)/(c.parent.childrenGrowSum/c.grow)))
		}

		if c.parent.firstChild == c {
			// The first child should have the same top/left as the parent
			c.box.top = c.parent.box.top
			c.box.left = c.parent.box.left
		} else {
			// The rest of the children should align the top/left of their box with the previous child's box
			if c.parent.isVertical {
				c.box.top = c.prevNeighbor.box.bottom
				c.box.left = c.parent.box.left
			} else {
				c.box.top = c.parent.box.top
				c.box.left = c.prevNeighbor.box.right
			}
		}
		if c.parent.lastChild == c {
			// The last child will, by default, align its bottom/right with the parent
			c.box.bottom = c.parent.box.bottom
			c.box.right = c.parent.box.right

			// If the last child doesn't have a neighbor with a
			// flex layout, and if it also has a fixed length,
			// don't align it with the parent
			if !c.parent.hasFlexChild && c.length != 0 {
				if c.parent.isVertical {
					c.box.bottom = c.box.top + c.length
				} else {
					c.box.right = c.box.left + c.length
				}
			}

			// If the last child 1) has a neighbor with a flex
			// layout, 2) has a fixed length, and 3) is not the
			// only child, snap it to the end of the parent, and
			// update its neighbors to align with itself.
			if c.parent.hasFlexChild && c.length != 0 && nChildren != 1 {
				if c.parent.isVertical {
					c.box.top = c.box.bottom - c.length
				} else {
					c.box.left = c.box.right - c.length
				}
				// Align previous children with this Component's new position
				first := c.prevNeighbor
				second := c
				for first != nil && second != nil {
					if c.parent.isVertical {
						if first.box.bottom == second.box.top {
							break
						}
						first.box.bottom = second.box.top
						if first.length != 0 {
							first.box.top = first.box.bottom - first.length
						}
					} else {
						if first.box.right == second.box.left {
							break
						}
						first.box.right = second.box.left
						if first.length != 0 {
							first.box.left = first.box.right - first.length
						}
					}
					first = first.prevNeighbor
					second = second.prevNeighbor
				}
			}
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

		// Apply scrolling
		c.box.top -= c.Scroll.Top
		c.box.left -= c.Scroll.Left
		c.box.right -= c.Scroll.Right
		c.box.bottom -= c.Scroll.Bottom
	}

	// Update content according to contentFunc
	if c.content.updateFunc != nil {
		value := c.content.updateFunc(&c.box)
		c.content.setValue(&value)
	}

	// Reset the blank areas since the layout may have changed
	c.firstBlankRow = -1
	c.firstBlankColumns = nil

	// Recursively update all children
	for _, child := range c.children {
		child.UpdateLayout()
	}

	for _, handler := range c.eventListeners[Event_LayoutUpdated] {
		(*handler)(c)
	}
}

func (c *Component) blankLine(width int) string {
	blankLine := strings.Repeat(BLANK_CHAR, max(0, width))
	return blankLine
}

func (c *Component) clipWithBounds(result *string, bounds *Box) {
	clipA := max(0, min(len(*result)-1, bounds.left-c.box.left))
	clipB := max(clipA, len(*result)-max(0, c.box.right-bounds.right))
	*result = (*result)[clipA:clipB]
}

// Render this Component's content to the screen, and render all child
// Components as well.
func (c *Component) Render() {
	if c.cancelRender != nil {
		c.cancelRender()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelRender = cancel

	// Recursively render all children
	startRow := 0
	for i, child := range c.children {
		select {
		case <-ctx.Done():
			return
		default:
		}
		child.Render()
		if i == len(c.children)-1 {
			startRow = child.box.bottom - c.children[0].box.top
		}
	}

	var builder strings.Builder

	width := c.box.Width()
	height := c.box.Height()
	firstBlankRow := -1                      // Will become the new value of c.firstBlankRow
	firstBlankColumns := make([]int, height) // Will become the new value of c.firstBlankColumns
	var a, b int                             // The start and end index of the content substring we are rendering
	var contentLen int
	var blankLine string

	// We want to handle things more efficiently if the content is blank
	isBlank := c.content.value == nil || *c.content.value == ""

	// If the screen is already clear, skip rendering
	if isBlank && c.firstBlankRow == 0 {
		return
	}

	if !isBlank {
		contentLen = c.content.displayLen()
	}

	// Loop through all parents to find bounding box so we can detect overflow
	bounds := c.box
	parent := c.parent
	for parent != nil {
		bounds.top = min(parent.box.bottom, max(parent.box.top, bounds.top))
		bounds.left = min(parent.box.right, max(parent.box.left, bounds.left))
		bounds.bottom = min(parent.box.bottom, max(parent.box.top, bounds.bottom))
		bounds.right = min(parent.box.right, max(parent.box.left, bounds.right))
		parent = parent.parent
	}

	// Construct our output line-by-line
	for row := startRow; row < height; row++ {
		select {
		case <-ctx.Done():
			return
		default:
			// Uncomment to view rendering order and test race conditions
			// time.Sleep(5 * time.Millisecond)
		}

		top := c.box.top + row

		// Check if we are out of bounds
		if top < bounds.top || top >= bounds.bottom || top+1 > bounds.bottom || c.box.left >= bounds.right {
			continue
		}

		// Position the cursor at the location of the current row
		builder.WriteString(fmt.Sprintf("\033[%d;%dH", top+1, max(bounds.left, c.box.left)+1))

		// Simple handler for blank content
		if isBlank {
			if blankLine == "" {
				blankLine = c.blankLine(width)
				c.clipWithBounds(&blankLine, &bounds)
				if c.colorFunc != nil {
					blankLine = c.colorFunc(blankLine)
				}
			}
			builder.WriteString(blankLine)
			continue
		}

		// We will print a substring of c.content from a:b
		b = a + width

		// Print content if it exists, and print blank space where there isn't content
		var result string
		if a < contentLen {
			var nBlanks int // The number of blank characters to append to the result

			// Get the section of content that should be rendered on this line
			substr := c.content.displaySubstring(a, min(contentLen, b))

			// Handle newlines by splitting the substr and rendering the remainder on the next iteration
			newlineIndex := strings.Index(substr, "\n")
			if newlineIndex != -1 {
				result = substr[:newlineIndex]
				nBlanks = b - a - newlineIndex
				a += newlineIndex + 1
			} else {
				result = substr
				nBlanks = b - contentLen
				a = b
			}

			// If we know where we already rendered blank space, update nBlanks accoringly
			if c.firstBlankColumns != nil && row < len(c.firstBlankColumns) {
				if c.firstBlankColumns[row] <= len(result) {
					nBlanks = 0
				} else {
					nBlanks = c.firstBlankColumns[row] - len(result)
				}
			}
			firstBlankColumns[row] = len(result)

			// Clear the remainder of the current line
			if nBlanks > 0 {
				result += strings.Repeat(BLANK_CHAR, nBlanks)
			}
		} else {
			// If we are done rendering content, save the first blank row
			if firstBlankRow == -1 {
				firstBlankRow = row
			}
			// If we are beyond the first blank row from the previous render, skip the rest of this render
			if c.firstBlankRow != -1 && c.firstBlankRow <= row {
				goto Output
			}

			// Clear the current line
			if blankLine == "" {
				blankLine = c.blankLine(width)
			}
			result = blankLine
		}

		c.clipWithBounds(&result, &bounds)
		if c.colorFunc != nil {
			result = c.colorFunc(result)
		}

		builder.WriteString(result)
	}

Output:
	// Update the locations of blank space from this render
	if isBlank {
		c.firstBlankRow = 0
	} else {
		c.firstBlankRow = firstBlankRow
	}
	c.firstBlankColumns = firstBlankColumns

	// Output to stdout
	fmt.Print(builder.String())
}
