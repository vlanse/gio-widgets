package split

import (
	"image"
	"image/color"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type View struct {
	// Ratio keeps the current layout.
	// 0 is center, -1 completely to the left, 1 completely to the right.
	Ratio float32
	// Bar is the width for resizing the layout
	Bar unit.Dp
	// BarColor is delimiter bar color
	BarColor color.NRGBA

	drag   bool
	dragID pointer.ID
	dragX  float32
}

const defaultBarWidth = unit.Dp(10)

func (s *View) Layout(gtx layout.Context, left, right layout.Widget) layout.Dimensions {
	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarWidth)
	}

	proportion := (s.Ratio + 1) / 2
	leftSize := int(proportion*float32(gtx.Constraints.Max.X) - float32(bar))

	rightOffset := leftSize + bar
	rightSize := gtx.Constraints.Max.X - rightOffset

	// handle input
	for _, ev := range gtx.Events(s) {
		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}

		switch e.Type {
		case pointer.Press:
			if s.drag {
				break
			}
			s.dragID = e.PointerID
			s.dragX = e.Position.X

		case pointer.Drag:
			if s.dragID != e.PointerID {
				break
			}
			deltaX := e.Position.X - s.dragX
			if e.Position.X < float32(gtx.Constraints.Min.X) || e.Position.X > float32(gtx.Constraints.Max.X) {
				break // do not allow bar to be dragged beyond current viewport
			}
			s.dragX = e.Position.X

			deltaRatio := deltaX * 2 / float32(gtx.Constraints.Max.X)
			s.Ratio += deltaRatio

		case pointer.Release:
			fallthrough
		case pointer.Cancel:
			s.drag = false
		}
	}

	// register for input, draw bar itself
	func() {
		area := clip.Rect(image.Rect(leftSize, 0, rightOffset, gtx.Constraints.Max.Y)).Op()
		defer area.Push(gtx.Ops).Pop()

		pointer.InputOp{
			Tag:   s,
			Types: pointer.Press | pointer.Drag | pointer.Release,
			Grab:  s.drag,
		}.Add(gtx.Ops)

		pointer.CursorColResize.Add(gtx.Ops) // customize cursor when over bar

		paint.FillShape(gtx.Ops, s.BarColor, area)
	}()

	{
		gtx := gtx
		gtx.Constraints = layout.Constraints{Max: image.Pt(leftSize, gtx.Constraints.Max.Y)}
		left(gtx)
	}

	{
		offset := op.Offset(image.Pt(rightOffset, 0)).Push(gtx.Ops)
		defer offset.Pop()
		gtx := gtx
		gtx.Constraints = layout.Constraints{Max: image.Pt(rightSize, gtx.Constraints.Max.Y)}
		right(gtx)
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}
