package terminal

import (
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Optional snapshots contain actual renderer cells, not a separate mockup.
func saveSnapshot(t *testing.T, s tcell.Screen, name string) {
	t.Helper()
	dir := os.Getenv("MONEYCOPY_SNAPSHOTS")
	if dir == "" {
		return
	}
	type cell struct {
		X, Y   int
		Text   string
		FG, BG int32
	}
	w, h := s.Size()
	cells := []cell{}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, _, st, width := s.GetContent(x, y)
			if st == tcell.StyleDefault {
				st = style(white)
			}
			fg, bg, _ := st.Decompose()
			cells = append(cells, cell{x, y, string(r), fg.Hex(), bg.Hex()})
			if width == 2 {
				x++
			}
		}
	}
	data, err := json.Marshal(struct {
		Width, Height int
		Cells         []cell
	}{w, h, cells})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, name+".json"), data, 0644); err != nil {
		t.Fatal(err)
	}
}
func TestRendererSnapshots(t *testing.T) {
	a, s, now := setup(t)
	input(a, "3600000")
	a.Draw()
	saveSnapshot(t, s, "intro")
	key(a, tcell.KeyEnter)
	for i := 0; i < 40; i++ {
		*now = now.Add(300 * time.Millisecond)
		a.Tick()
	}
	a.session.SpawnBonus(*now, 10*time.Second)
	a.bonusX = .57
	*now = now.Add(3 * time.Second)
	a.Tick()
	a.Draw()
	saveSnapshot(t, s, "game")
	input(a, " ")
	key(a, tcell.KeyEscape)
	a.Draw()
	saveSnapshot(t, s, "result")
	s.SetSize(76, 28)
	a.Draw()
	saveSnapshot(t, s, "small-result")
}
