package flextui

import "unicode/utf8"

type content struct {
	value      *string
	updateFunc func(*Box) string
	hasUnicode bool
}

func (c *content) setValue(value *string) {
	c.hasUnicode = len(*value) != utf8.RuneCountInString(*value)
	c.value = value
}

func (c *content) displayLen() int {
	if c.hasUnicode {
		return utf8.RuneCountInString(*c.value)
	}
	return len(*c.value)
}

func (c *content) realIndex(displayIndex int) int {
	if c.hasUnicode {
		if displayIndex == 0 {
			return 0
		}
		runeCount := utf8.RuneCountInString(*c.value)
		if runeCount < displayIndex {
			return len(*c.value) + (displayIndex - runeCount)
		}
		for i := len(*c.value); i > 0; i-- {
			runeCount = utf8.RuneCountInString((*c.value)[:i])
			if runeCount == displayIndex {
				return i
			}
		}
	}
	return displayIndex
}

func (c *content) displaySubstring(a, b int) string {
	if c.hasUnicode {
		return (*c.value)[c.realIndex(a):c.realIndex(b)]
	}
	return (*c.value)[a:b]
}
