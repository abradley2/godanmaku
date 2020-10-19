package flex

import (
	"fmt"
	"image"
	"math"

	"github.com/yohamta/godanmaku/danmaku/internal/uikit"
)

// Direction is the direction in which flex items are laid out
type Direction uint8

const (
	Row Direction = iota
	Column
)

// Justify aligns items along the main axis.
//
// https://www.w3.org/TR/css-flexbox-1/#justify-content-property
type Justify uint8

const (
	JustifyStart        Justify = iota // pack to start of line
	JustifyEnd                         // pack to end of line
	JustifyCenter                      // pack to center of line
	JustifySpaceBetween                // even spacing
	JustifySpaceAround                 // even spacing, half-size on each end
)

// FlexAlign represents align of flex children
type FlexAlign int

const (
	FlexCenter FlexAlign = iota
	FlexStart
	FlexEnd
	FlexSpaceBetween
)

// AlignItem aligns items along the cross axis.
type AlignItem uint8

const (
	AlignItemStart AlignItem = iota
	AlignItemEnd
	AlignItemCenter
)

// FlexWrap controls whether the container is single- or multi-line,
// and the direction in which the lines are laid out.
type FlexWrap uint8

const (
	NoWrap FlexWrap = iota
	Wrap
	WrapReverse
)

// AlignContent is the 'align-content' property.
// It aligns container lines when there is extra space on the cross-axis.
type AlignContent uint8

const (
	AlignContentStretch AlignContent = iota
	AlignContentStart
	AlignContentEnd
	AlignContentCenter
	AlignContentSpaceBetween
	AlignContentSpaceAround
)

// Flex is a container widget that lays out its children following the
// CSS flexbox algorithm.
type Flex struct {
	uikit.ViewEventHandlerFuncs

	Direction    Direction
	Wrap         FlexWrap
	Justify      Justify
	AlignItems   AlignItem
	AlignContent AlignContent

	width, height int
}

// NewFlex creates NewFlexContaienr
func NewFlex(width, height int) *Flex {
	f := new(Flex)

	f.Direction = Row
	f.Wrap = NoWrap
	f.Justify = JustifyCenter
	f.AlignItems = AlignItemCenter
	f.AlignContent = AlignContentCenter

	f.width = width
	f.height = height

	return f
}

func (f *Flex) OnLoad(v *uikit.View) {
	v.SetSize(f.width, f.height)
}

func (f *Flex) OnLayout(v *uikit.View) {
	var children []element
	subViews := v.GetSubView()
	for i := 0; i < len(subViews); i++ {
		c := subViews[i]
		children = append(children, element{
			flexBaseSize: float64(f.flexBaseSize(c)),
			n:            c,
		})
	}

	containerMainSize := float64(f.mainSize(v.Rect().Size()))
	// containerCrossSize := float64(f.crossSize(f.Rect().Size()))

	var lines []flexLine
	if f.Wrap == NoWrap {
		line := flexLine{child: make([]*element, len(children))}
		for i := range children {
			child := &children[i]
			line.child[i] = child
			line.mainSize += child.flexBaseSize
		}
		lines = []flexLine{line}
	} else {
		panic("not implemented")
	}

	for l := range lines {
		line := &lines[l]

		// Calculate free space
		freeSpace := float64(f.mainSize(v.Rect().Size()))
		for _, child := range line.child {
			freeSpace -= float64(f.flexBaseSize(child.n))
		}

		// Distribute free space
		for _, child := range line.child {
			child.mainSize = float64(f.flexBaseSize(child.n))
		}
	}

	// Determine cross size
	for l := range lines {
		for _, child := range lines[l].child {
			child.crossSize = float64(f.crossSize(child.n.Rect().Size()))
		}
	}

	if len(lines) == 1 {
		// Single line
		switch f.Direction {
		case Row:
			lines[0].crossSize = float64(v.Rect().Size().Y)
		case Column:
			lines[0].crossSize = float64(v.Rect().Size().X)
		}
	} else {
		panic("not implemented for multi line")
	}

	off := 0.0
	for l := range lines {
		line := &lines[l]
		line.crossOffset = off
		off += line.crossSize
	}

	// Main axis alignment
	for l := range lines {
		line := &lines[l]
		total := 0.0
		for _, child := range line.child {
			total += child.mainSize
		}
		remFree := containerMainSize - total
		off, spacing := 0.0, 0.0
		switch f.Justify {
		case JustifyStart:
		case JustifyEnd:
			off = remFree
		case JustifyCenter:
			off = remFree / 2
		case JustifySpaceBetween:
			spacing = remFree / float64(len(line.child)-1)
		case JustifySpaceAround:
			spacing = remFree / float64(len(line.child))
			off = spacing / 2
		}
		for _, child := range line.child {
			child.mainOffset = off
			off += spacing + child.mainSize
		}
	}

	// Cross axis alignment
	for l := range lines {
		line := &lines[l]
		for _, child := range line.child {
			child.crossOffset = line.crossOffset
			if child.crossSize == line.crossSize {
				continue
			}
			diff := line.crossSize - child.crossSize
			switch f.AlignItems {
			case AlignItemStart:
				// already laid out correctly
			case AlignItemEnd:
				child.crossOffset = line.crossOffset + diff
			case AlignItemCenter:
				child.crossOffset = line.crossOffset + diff/2
			}
		}
	}

	// Layout complete. Generate child Rect values.
	for l := range lines {
		line := &lines[l]
		for _, child := range line.child {
			switch f.Direction {
			case Row:
				child.n.SetRect(round(child.mainOffset),
					round(child.crossOffset),
					round(child.mainOffset+child.mainSize),
					round(child.crossOffset+child.crossSize))
			case Column:
				child.n.SetRect(round(child.crossOffset),
					round(child.mainOffset),
					round(child.crossOffset+child.crossSize),
					round(child.mainOffset+child.mainSize))
			default:
				panic(fmt.Sprint("flex: bad direction ", f.Direction))
			}
			child.n.Layout()
		}
	}
}

type element struct {
	n            *uikit.View
	flexBaseSize float64
	frozen       bool
	unclamped    float64
	mainSize     float64
	mainOffset   float64
	crossSize    float64
	crossOffset  float64
}

type flexLine struct {
	mainSize    float64
	crossSize   float64
	crossOffset float64
	child       []*element
}

func (f *Flex) mainSize(p image.Point) int {
	switch f.Direction {
	case Row:
		return p.X
	case Column:
		return p.Y
	default:
		panic(fmt.Sprint("flex: bad direction ", f.Direction))
	}
}

func (f *Flex) crossSize(p image.Point) int {
	switch f.Direction {
	case Row:
		return p.Y
	case Column:
		return p.X
	default:
		panic(fmt.Sprint("flex: bad direction ", f.Direction))
	}
}

func (f *Flex) flexBaseSize(v *uikit.View) int {
	return f.mainSize(v.Rect().Size())
}

func round(f float64) int {
	return int(math.Floor(f + .5))
}
