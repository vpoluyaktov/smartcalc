package ui

import "fyne.io/fyne/v2"

// FixedWidthLayout is a custom layout that gives a fixed width to its children.
type FixedWidthLayout struct {
	Width float32
}

func (f *FixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(f.Width, 0)
}

func (f *FixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(f.Width, size.Height))
		o.Move(fyne.NewPos(0, 0))
	}
}
