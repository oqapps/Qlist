package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Text struct {
	widget.BaseWidget
	Resource *canvas.Text

	DoubleTapEvent func(_ *fyne.PointEvent)
}

func (t *Text) SetText(text string) {
	t.Resource.Text = text
	t.Refresh()
}

func (t *Text) DoubleTapped(event *fyne.PointEvent) {
	if t.DoubleTapEvent != nil {
		t.DoubleTapEvent(event)
	}
}

func (t *Text) SetDoubleTapEvent(event func(_ *fyne.PointEvent)) {
	t.DoubleTapEvent = event
}

func (t *Text) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.Resource)
}

func NewText(input string) *Text {
	resource := canvas.NewText(input, theme.TextColor())
	//resource := widget.NewLabel(input)
	text := &Text{Resource: resource}
	text.ExtendBaseWidget(text)
	return text
}
