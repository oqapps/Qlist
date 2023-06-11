package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Entry struct  {
	widget.Entry
	DoubleTapEvent func(_ *fyne.PointEvent)
}



func (t *Entry) DoubleTapped (event *fyne.PointEvent)  {
	if t.DoubleTapEvent != nil {
		t.DoubleTapEvent(event)
	}
}

func (t *Entry) SetDoubleTapEvent(event func(_ *fyne.PointEvent)) {
	t.DoubleTapEvent = event
}


func NewEntry(input string) *Entry {
	entry := Entry{}
	entry.SetText(input)
	return &entry
}