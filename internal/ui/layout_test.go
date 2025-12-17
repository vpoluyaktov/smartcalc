package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

func TestFixedWidthLayoutMinSize(t *testing.T) {
	layout := &FixedWidthLayout{Width: 100}

	// MinSize should return the fixed width with 0 height
	size := layout.MinSize(nil)

	if size.Width != 100 {
		t.Errorf("MinSize().Width = %v, want 100", size.Width)
	}
	if size.Height != 0 {
		t.Errorf("MinSize().Height = %v, want 0", size.Height)
	}
}

func TestFixedWidthLayoutLayout(t *testing.T) {
	layout := &FixedWidthLayout{Width: 50}

	// Create a test object
	rect := canvas.NewRectangle(nil)
	objects := []fyne.CanvasObject{rect}

	// Layout with a container size
	containerSize := fyne.NewSize(200, 300)
	layout.Layout(objects, containerSize)

	// Object should be resized to fixed width and container height
	if rect.Size().Width != 50 {
		t.Errorf("object width = %v, want 50", rect.Size().Width)
	}
	if rect.Size().Height != 300 {
		t.Errorf("object height = %v, want 300", rect.Size().Height)
	}

	// Object should be positioned at origin
	if rect.Position().X != 0 {
		t.Errorf("object X = %v, want 0", rect.Position().X)
	}
	if rect.Position().Y != 0 {
		t.Errorf("object Y = %v, want 0", rect.Position().Y)
	}
}

func TestFixedWidthLayoutMultipleObjects(t *testing.T) {
	layout := &FixedWidthLayout{Width: 75}

	rect1 := canvas.NewRectangle(nil)
	rect2 := canvas.NewRectangle(nil)
	objects := []fyne.CanvasObject{rect1, rect2}

	containerSize := fyne.NewSize(200, 150)
	layout.Layout(objects, containerSize)

	// Both objects should have the same fixed width
	if rect1.Size().Width != 75 {
		t.Errorf("rect1 width = %v, want 75", rect1.Size().Width)
	}
	if rect2.Size().Width != 75 {
		t.Errorf("rect2 width = %v, want 75", rect2.Size().Width)
	}
}
