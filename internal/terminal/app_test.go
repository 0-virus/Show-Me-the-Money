package terminal

import (
	"github.com/gdamore/tcell/v2"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func setup(t *testing.T) (*App, tcell.SimulationScreen, *time.Time) {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(110, 36)
	t.Cleanup(s.Fini)
	now := time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)
	return New(s, func() time.Time { return now }, rand.New(rand.NewSource(1))), s, &now
}
func input(a *App, text string) {
	for _, r := range text {
		a.Handle(tcell.NewEventKey(tcell.KeyRune, r, 0))
	}
}
func key(a *App, k tcell.Key) { a.Handle(tcell.NewEventKey(k, 0, 0)) }
func dump(s tcell.Screen) string {
	var b strings.Builder
	w, h := s.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, _, _, width := s.GetContent(x, y)
			b.WriteRune(r)
			if width == 2 {
				x++
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}
func TestInputPlayResultAndRestart(t *testing.T) {
	a, s, now := setup(t)
	a.Draw()
	if !strings.Contains(dump(s), "가만히 있어도") {
		t.Fatal("missing title")
	}
	key(a, tcell.KeyEnter)
	if a.page != intro || a.message == "" {
		t.Fatal("accepted empty input")
	}
	input(a, "3,600,000")
	key(a, tcell.KeyEnter)
	if a.page != playing {
		t.Fatal("did not start", a.message)
	}
	*now = now.Add(10 * time.Second)
	a.Tick()
	a.Draw()
	if a.session.Base() != 55 || !strings.Contains(dump(s), "55") {
		t.Fatal("income not rendered")
	}
	key(a, tcell.KeyEscape)
	a.Draw()
	if a.page != result || !strings.Contains(dump(s), "기본급") {
		t.Fatal("missing result")
	}
	key(a, tcell.KeyEnter)
	if a.page != intro || a.session != nil {
		t.Fatal("restart leaked session")
	}
}
func TestAnnualSelectionAndInvalidInput(t *testing.T) {
	a, _, _ := setup(t)
	input(a, "0")
	key(a, tcell.KeyEnter)
	if a.page != intro {
		t.Fatal("accepted zero")
	}
	key(a, tcell.KeyCtrlU)
	key(a, tcell.KeyTab)
	input(a, "43200000")
	key(a, tcell.KeyEnter)
	if a.session == nil || a.session.Hourly() != 20000 {
		t.Fatal("annual conversion")
	}
}
func TestBonusMouseAndLanding(t *testing.T) {
	a, s, now := setup(t)
	input(a, "3600000")
	key(a, tcell.KeyEnter)
	*now = a.nextBonus
	a.Tick()
	a.Draw()
	if !a.session.BonusActive {
		t.Fatal("bonus did not spawn")
	}
	rect := a.bonusRect(*now)
	a.Handle(tcell.NewEventMouse(rect.x+2, rect.y+2, tcell.Button1, 0))
	if a.session.Bonus() != 10000 {
		t.Fatal("click missed visible bonus")
	}
	a.Handle(tcell.NewEventMouse(rect.x+2, rect.y+2, tcell.Button1, 0))
	if a.session.Bonus() != 10000 {
		t.Fatal("duplicate bonus")
	}
	*now = a.nextBonus
	a.Tick()
	*now = now.Add(10 * time.Second)
	a.Tick()
	a.Draw()
	key(a, tcell.KeyRune)
	if a.session.Bonus() != 10000 {
		t.Fatal("landing paid money")
	}
	_ = s
}
func TestDeadlineResizeAndMouseButtons(t *testing.T) {
	a, s, now := setup(t)
	input(a, "3600000")
	a.Draw()
	r := a.startRect()
	a.Handle(tcell.NewEventMouse(r.x+1, r.y, tcell.Button1, 0))
	if a.page != playing {
		t.Fatal("start button missed")
	}
	s.SetSize(40, 12)
	a.Draw()
	if !strings.Contains(dump(s), "창") {
		t.Fatal("missing size guidance")
	}
	*now = a.session.Deadline
	a.Tick()
	if a.page != result {
		t.Fatal("deadline did not settle while small")
	}
	s.SetSize(90, 30)
	a.Draw()
	if !strings.Contains(dump(s), "18:00") {
		t.Fatal("cutoff not shown")
	}
	a.Handle(tcell.NewEventMouse(0, 0, tcell.ButtonNone, 0))
	r = a.startRect()
	a.Handle(tcell.NewEventMouse(r.x+1, r.y, tcell.Button1, 0))
	if a.page != intro {
		t.Fatal("back button missed")
	}
}

func TestDefaultTerminalSizeCanStart(t *testing.T) {
	a, s, _ := setup(t)
	s.SetSize(80, 24)
	input(a, "3600000")
	a.Draw()
	key(a, tcell.KeyEnter)
	if a.page != playing {
		t.Fatal("default 80x24 terminal cannot play")
	}
}
