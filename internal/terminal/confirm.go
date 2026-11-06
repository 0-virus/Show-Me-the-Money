package terminal

import (
	"github.com/gdamore/tcell/v2"
	"strings"
	"time"
)

func (a *App) requestExit() { a.confirmExit = true; a.confirmYes = false }
func (a *App) resolveExit(yes bool, now time.Time) {
	a.confirmExit = false
	if !yes {
		return
	}
	if a.page == playing {
		a.session.Finish(now)
		a.page = result
	} else if a.page == intro {
		a.quit = true
	}
}
func (a *App) confirmationRect() rect {
	w, h := a.screen.Size()
	return rect{(w - 50) / 2, (h - 10) / 2, 50, 10}
}
func (a *App) confirmationButton(yes bool) rect {
	r := a.confirmationRect()
	x := r.x + 6
	if yes {
		x = r.x + 28
	}
	return rect{x, r.y + 4, 16, 3}
}
func (a *App) handleConfirmation(ev tcell.Event, now time.Time) {
	switch e := ev.(type) {
	case *tcell.EventResize:
		a.screen.Sync()
	case *tcell.EventKey:
		if e.Key() == tcell.KeyCtrlC {
			a.quit = true
			return
		}
		if e.Key() == tcell.KeyEscape {
			a.resolveExit(false, now)
			return
		}
		if !a.largeEnough() {
			return
		}
		switch e.Key() {
		case tcell.KeyTab, tcell.KeyBacktab, tcell.KeyLeft, tcell.KeyRight:
			a.confirmYes = !a.confirmYes
		case tcell.KeyEnter:
			a.resolveExit(a.confirmYes, now)
		case tcell.KeyRune:
			if e.Rune() == 'y' || e.Rune() == 'Y' {
				a.resolveExit(true, now)
			}
			if e.Rune() == 'n' || e.Rune() == 'N' {
				a.resolveExit(false, now)
			}
		}
	case *tcell.EventMouse:
		down := e.Buttons()&tcell.Button1 != 0
		fresh := down && !a.mouseDown
		a.mouseDown = down
		if !fresh || !a.largeEnough() {
			return
		}
		x, y := e.Position()
		for _, yes := range []bool{false, true} {
			if a.confirmationButton(yes).contains(x, y) {
				a.resolveExit(yes, now)
				return
			}
		}
	}
}
func (a *App) drawConfirmation() {
	a.screen.HideCursor()
	r := a.confirmationRect()
	a.fill(r, style(white).Background(panel))
	border := style(green).Background(panel)
	a.text(r.x, r.y, "+"+strings.Repeat("-", r.w-2)+"+", border)
	a.text(r.x, r.y+r.h-1, "+"+strings.Repeat("-", r.w-2)+"+", border)
	for y := r.y + 1; y < r.y+r.h-1; y++ {
		a.text(r.x, y, "|", border)
		a.text(r.x+r.w-1, y, "|", border)
	}
	question := "퇴근하시겠습니까?"
	if a.page == intro {
		question = "오후 반차 내시는 겁니까?"
	}
	a.center(r.y+2, question, style(white).Background(panel).Bold(true))
	a.button(a.confirmationButton(false), "아니요", !a.confirmYes)
	a.button(a.confirmationButton(true), "예", a.confirmYes)
	a.center(r.y+8, "Tab 전환 · Enter 선택 · Esc 취소", style(muted).Background(panel))
}
