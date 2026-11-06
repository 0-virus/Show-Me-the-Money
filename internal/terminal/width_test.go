package terminal

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Korean Windows code page 949 makes go-runewidth treat box/block glyphs
// as wide, while tcell and Windows Terminal display them in one cell.
func TestConnectedArtInKoreanLocale(t *testing.T) {
	previous := runewidth.DefaultCondition
	runewidth.DefaultCondition = &runewidth.Condition{EastAsianWidth: true, StrictEmojiNeutral: true}
	t.Cleanup(func() { runewidth.DefaultCondition = previous })
	a, s, now := setup(t)
	a.screen.Clear()
	a.bill(5, 5, false, 0)
	a.screen.Show()
	for x := 5; x < 13; x++ {
		r, _, _, width := s.GetContent(x, 5)
		if r == ' ' || width != 1 {
			t.Errorf("bill border gap at column %d: %q (width %d)", x, r, width)
		}
	}
	a.text(20, 5, "한글", style(white))
	a.screen.Show()
	for i, want := range []rune("한글") {
		r, _, _, width := s.GetContent(20+i*2, 5)
		if r != want || width != 2 {
			t.Errorf("Korean text width changed: %q width %d", r, width)
		}
	}
	input(a, "3600000")
	key(a, tcell.KeyEnter)
	*now = now.Add(30 * time.Second)
	a.Tick()
	a.Draw()
	start := a.frame().x + 3
	for x := start; x < start+6; x++ {
		r, _, _, width := s.GetContent(x, 10)
		if r != '█' || width != 1 {
			t.Errorf("gauge gap at column %d: %q (width %d)", x, r, width)
		}
	}
}
