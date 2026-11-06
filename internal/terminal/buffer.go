package terminal

import "github.com/gdamore/tcell/v2"

// Compose a complete frame before touching the terminal. Clearing a real screen
// before redrawing wide text otherwise marks unchanged Korean glyphs dirty.
type bufferedScreen struct {
	tcell.Screen
	cells            tcell.CellBuffer
	cursorX, cursorY int
}

func (s *bufferedScreen) Clear() {
	w, h := s.Size()
	s.cells.Resize(w, h)
	s.cells.Fill(' ', style(white))
}
func (s *bufferedScreen) SetContent(x, y int, r rune, c []rune, st tcell.Style) {
	s.cells.SetContent(x, y, r, c, st)
}
func (s *bufferedScreen) HideCursor()         { s.cursorX = -1; s.cursorY = -1 }
func (s *bufferedScreen) ShowCursor(x, y int) { s.cursorX = x; s.cursorY = y }
func (s *bufferedScreen) Show() {
	w, h := s.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, c, st, width := s.cells.GetContent(x, y)
			oldR, oldC, oldStyle, _ := s.Screen.GetContent(x, y)
			if r != oldR || string(c) != string(oldC) || st != oldStyle {
				s.Screen.SetContent(x, y, r, c, st)
			}
			if width == 2 {
				x++
			}
		}
	}
	s.Screen.ShowCursor(s.cursorX, s.cursorY)
	s.Screen.Show()
}
