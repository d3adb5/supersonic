package widgets

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	myTheme "github.com/supersonic-app/supersonic/ui/theme"
)

// SpeedLabel shows the current playback speed (e.g. "1.00x").
// It is tappable, and colored like an IconButton so it sits naturally
// alongside the transport buttons.
type SpeedLabel struct {
	widget.BaseWidget

	OnTapped func()

	speed    float64
	hovered  bool
	disabled bool

	text *canvas.Text
}

var (
	_ fyne.Tappable      = (*SpeedLabel)(nil)
	_ fyne.Disableable   = (*SpeedLabel)(nil)
	_ desktop.Hoverable  = (*SpeedLabel)(nil)
	_ desktop.Cursorable = (*SpeedLabel)(nil)
)

func NewSpeedLabel(speed float64) *SpeedLabel {
	s := &SpeedLabel{speed: speed}
	s.ExtendBaseWidget(s)
	return s
}

func formatSpeed(speed float64) string {
	return fmt.Sprintf("%.2fx", speed)
}

func (s *SpeedLabel) SetSpeed(speed float64) {
	if s.speed == speed {
		return
	}
	s.speed = speed
	s.Refresh()
}

func (s *SpeedLabel) Tapped(*fyne.PointEvent) {
	if !s.disabled && s.OnTapped != nil {
		s.OnTapped()
	}
}

func (s *SpeedLabel) Disable() {
	if !s.disabled {
		s.disabled = true
		s.Refresh()
	}
}

func (s *SpeedLabel) Enable() {
	if s.disabled {
		s.disabled = false
		s.Refresh()
	}
}

func (s *SpeedLabel) Disabled() bool {
	return s.disabled
}

func (s *SpeedLabel) MouseIn(*desktop.MouseEvent) {
	if !s.hovered {
		s.hovered = true
		s.Refresh()
	}
}

func (s *SpeedLabel) MouseOut() {
	if s.hovered {
		s.hovered = false
		s.Refresh()
	}
}

func (s *SpeedLabel) MouseMoved(*desktop.MouseEvent) {}

func (s *SpeedLabel) Cursor() desktop.Cursor {
	if s.disabled {
		return desktop.DefaultCursor
	}
	return desktop.PointerCursor
}

func (s *SpeedLabel) MinSize() fyne.Size {
	// size for the widest possible value so the layout doesn't shift as speed changes
	sz := fyne.MeasureText("0.00x", s.textSize(), s.textStyle())
	return sz.Add(fyne.NewSize(theme.Padding(), 0))
}

func (s *SpeedLabel) textSize() float32 {
	return theme.TextSize()
}

func (s *SpeedLabel) textStyle() fyne.TextStyle {
	return fyne.TextStyle{Bold: true}
}

func (s *SpeedLabel) colorName() fyne.ThemeColorName {
	switch {
	case s.disabled:
		return theme.ColorNameDisabled
	case s.hovered:
		return myTheme.ColorNameHoveredIconButton
	default:
		return myTheme.ColorNameIconButton
	}
}

func (s *SpeedLabel) Refresh() {
	if s.text == nil {
		return
	}
	s.text.Text = formatSpeed(s.speed)
	s.text.TextSize = s.textSize()
	s.text.Color = theme.Color(s.colorName())
	s.text.Refresh()
}

func (s *SpeedLabel) CreateRenderer() fyne.WidgetRenderer {
	s.text = canvas.NewText(formatSpeed(s.speed), theme.Color(s.colorName()))
	s.text.TextStyle = s.textStyle()
	s.text.TextSize = s.textSize()
	s.text.Alignment = fyne.TextAlignCenter
	return widget.NewSimpleRenderer(container.NewCenter(s.text))
}
