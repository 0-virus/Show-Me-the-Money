package terminal

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"math/rand"
	"moneycopy/internal/game"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type page int

const (
	intro page = iota
	playing
	result
)

type rect struct{ x, y, w, h int }

func (r rect) contains(x, y int) bool { return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h }

type note struct {
	x        float64
	born     time.Time
	duration time.Duration
}
type App struct {
	screen     tcell.Screen
	now        func() time.Time
	random     *rand.Rand
	page       page
	message    string
	session    *game.Session
	nextBonus  time.Time
	amount     string
	annual     bool
	quit       bool
	mouseDown  bool
	bonusX     float64
	nextNote   time.Time
	notes      []note
	toastUntil time.Time
	toast      string
}

func New(s tcell.Screen, now func() time.Time, r *rand.Rand) *App {
	return &App{screen: &bufferedScreen{Screen: s}, now: func() time.Time { return now().Round(0) }, random: r}
}
func (a *App) start() {
	n, err := strconv.ParseInt(strings.ReplaceAll(a.amount, ",", ""), 10, 64)
	if err != nil {
		a.message = "급여를 숫자로 입력해 주세요. 예: 3,600,000"
		return
	}
	now := a.now()
	s, err := game.New(n, a.annual, now)
	if err != nil {
		a.message = err.Error()
		return
	}
	a.session = s
	a.page = playing
	a.message = ""
	a.notes = nil
	a.nextBonus = now.Add(a.bonusDelay())
	a.nextNote = now
	a.toast = "돈이 일하는 중입니다. 잠시 쉬어 가세요."
	a.toastUntil = now.Add(4 * time.Second)
}
func (a *App) bonusDelay() time.Duration { return time.Duration(180+a.random.Intn(241)) * time.Second }
func (a *App) back()                     { a.page = intro; a.session = nil; a.message = ""; a.notes = nil }
func (a *App) claim(now time.Time) {
	before := a.session.Bonus()
	if a.session.ClaimBonus(now) {
		a.toast = fmt.Sprintf("성과급 +%s원!  30분의 여유를 얻었습니다.", money(a.session.Bonus()-before))
		a.toastUntil = now.Add(3 * time.Second)
	}
}
func (a *App) Handle(ev tcell.Event) {
	now := a.now()
	// Settle before processing input, including an input arriving at exactly 18:00.
	a.Tick()
	switch e := ev.(type) {
	case *tcell.EventResize:
		a.screen.Sync()
	case *tcell.EventKey:
		if e.Key() == tcell.KeyCtrlC {
			a.quit = true
			return
		}
		switch a.page {
		case intro:
			switch e.Key() {
			case tcell.KeyEnter:
				if a.largeEnough() {
					a.start()
				}
			case tcell.KeyTab, tcell.KeyBacktab:
				a.annual = !a.annual
				a.message = ""
			case tcell.KeyCtrlU:
				a.amount = ""
				a.message = ""
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if len(a.amount) > 0 {
					runes := []rune(a.amount)
					a.amount = string(runes[:len(runes)-1])
				}
				a.message = ""
			case tcell.KeyEscape:
				a.quit = true
			case tcell.KeyRune:
				if unicode.IsPrint(e.Rune()) && len([]rune(a.amount)) < 20 {
					a.amount += string(e.Rune())
					a.message = ""
				}
			}
		case playing:
			if e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyRune && (e.Rune() == 'q' || e.Rune() == 'Q') {
				a.session.Finish(now)
				a.page = result
			}
			if a.largeEnough() && e.Key() == tcell.KeyRune && e.Rune() == ' ' {
				a.claim(now)
			}
		case result:
			if e.Key() == tcell.KeyEnter {
				a.back()
			}
			if e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyRune && e.Rune() == 'q' {
				a.quit = true
			}
		}
	case *tcell.EventMouse:
		pressed := e.Buttons()&tcell.Button1 != 0
		fresh := pressed && !a.mouseDown
		a.mouseDown = pressed
		if !fresh || !a.largeEnough() {
			return
		}
		x, y := e.Position()
		switch a.page {
		case intro:
			if a.monthRect().contains(x, y) {
				a.annual = false
			}
			if a.yearRect().contains(x, y) {
				a.annual = true
			}
			if a.startRect().contains(x, y) {
				a.start()
			}
		case playing:
			if a.stopRect().contains(x, y) {
				a.session.Finish(now)
				a.page = result
				return
			}
			if a.session.BonusActive && a.bonusRect(now).contains(x, y) {
				a.claim(now)
			}
		case result:
			if a.startRect().contains(x, y) {
				a.back()
			}
		}
	}
}
func (a *App) Tick() {
	if a.page != playing {
		return
	}
	now := a.now()
	a.session.Advance(now)
	if a.session.Finished {
		a.page = result
		return
	}
	if !now.Before(a.nextBonus) {
		if a.session.SpawnBonus(now, 10*time.Second) {
			a.bonusX = a.random.Float64()
		}
		a.nextBonus = now.Add(a.bonusDelay())
	}
	if !now.Before(a.nextNote) {
		a.notes = append(a.notes, note{a.random.Float64(), now, time.Duration(7+a.random.Intn(6)) * time.Second})
		a.nextNote = now.Add(time.Duration(400+a.random.Intn(500)) * time.Millisecond)
	}
	live := a.notes[:0]
	for _, n := range a.notes {
		if now.Sub(n.born) < n.duration {
			live = append(live, n)
		}
	}
	a.notes = live
}
func (a *App) Run() {
	events := make(chan tcell.Event, 16)
	quit := make(chan struct{})
	defer close(quit)
	go a.screen.ChannelEvents(events, quit)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	a.Draw()
	for !a.quit {
		select {
		case ev := <-events:
			if ev == nil {
				return
			}
			a.Handle(ev)
		case <-ticker.C:
			if a.page != playing {
				continue
			}
			a.Tick()
		}
		a.Draw()
	}
}
func (a *App) largeEnough() bool { w, h := a.screen.Size(); return w >= 76 && h >= 24 }
func (a *App) frame() rect {
	w, h := a.screen.Size()
	cw := min(w-4, 112)
	return rect{(w - cw) / 2, 1, cw, h - 2}
}
func (a *App) formY() int      { _, h := a.screen.Size(); return max(8, (h-20)/2+6) }
func (a *App) startRect() rect { w, _ := a.screen.Size(); return rect{w/2 - 13, a.formY() + 10, 26, 3} }
func (a *App) monthRect() rect { w, _ := a.screen.Size(); return rect{w/2 - 22, a.formY(), 20, 3} }
func (a *App) yearRect() rect  { w, _ := a.screen.Size(); return rect{w/2 + 2, a.formY(), 20, 3} }
func (a *App) stopRect() rect  { f := a.frame(); return rect{f.x + f.w - 17, 2, 14, 1} }
func (a *App) bonusRect(now time.Time) rect {
	f := a.frame()
	floor := a.floor()
	progress := now.Sub(a.session.BonusStarted).Seconds() / 10
	progress = max(0, min(progress, 1))
	return rect{f.x + 2 + int(a.bonusX*float64(f.w-26)), 13 + int(progress*float64(max(1, floor-13-5))), 22, 5}
}
func (a *App) floor() int {
	_, h := a.screen.Size()
	extra := 0
	if a.session != nil {
		extra = min(3, max(0, h-26), int(a.session.Ticks/60))
	}
	return h - 5 - extra
}
