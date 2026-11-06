package terminal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
	"moneycopy/internal/game"
)

var (
	bg    = tcell.NewHexColor(0x08120f)
	panel = tcell.NewHexColor(0x10221b)
	green = tcell.NewHexColor(0x91f7b3)
	white = tcell.NewHexColor(0xe0ede4)
	muted = tcell.NewHexColor(0x779b88)
	line  = tcell.NewHexColor(0x284f3d)
	gold  = tcell.NewHexColor(0xffd66b)
	red   = tcell.NewHexColor(0xf39785)
)

func style(fg tcell.Color) tcell.Style { return tcell.StyleDefault.Foreground(fg).Background(bg) }
func money(n int64) string {
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
func decimal(n float64) string {
	s := fmt.Sprintf("%.2f", n)
	p := strings.Split(s, ".")
	i, _ := strconv.ParseInt(p[0], 10, 64)
	return money(i) + "." + p[1]
}
func duration(d time.Duration) string {
	n := int64(max(0, d.Seconds()))
	return fmt.Sprintf("%02d:%02d:%02d", n/3600, n/60%60, n%60)
}
func (a *App) text(x, y int, s string, st tcell.Style) {
	w, h := a.screen.Size()
	if y < 0 || y >= h {
		return
	}
	for _, r := range s {
		// Match tcell CellBuffer; locale-dependent ambiguous widths create gaps.
		rw := uniseg.StringWidth(string(r))
		if rw <= 0 {
			continue
		}
		if x >= 0 && x+rw <= w {
			a.screen.SetContent(x, y, r, nil, st)
		}
		x += rw
	}
}
func (a *App) center(y int, s string, st tcell.Style) {
	w, _ := a.screen.Size()
	a.text((w-uniseg.StringWidth(s))/2, y, s, st)
}
func (a *App) fill(r rect, st tcell.Style) {
	for y := r.y; y < r.y+r.h; y++ {
		for x := r.x; x < r.x+r.w; x++ {
			a.screen.SetContent(x, y, ' ', nil, st)
		}
	}
}
func (a *App) rule(x, y, w int) { a.text(x, y, strings.Repeat("─", max(0, w)), style(line)) }
func (a *App) button(r rect, label string, active bool) {
	st := style(muted).Background(panel)
	if active {
		st = style(bg).Background(green).Bold(true)
	}
	a.fill(r, st)
	a.text(r.x+(r.w-uniseg.StringWidth(label))/2, r.y+r.h/2, label, st)
}
func (a *App) Draw() {
	a.screen.SetStyle(style(white))
	a.screen.Clear()
	a.screen.HideCursor()
	if !a.largeEnough() {
		_, h := a.screen.Size()
		a.center(max(0, h/2-2), "창을 조금 넓혀 주세요", style(green))
		a.center(max(1, h/2), "최소 76열 × 24행", style(white))
		a.center(max(2, h/2+2), "게임 시계는 계속 흐릅니다", style(muted))
		a.screen.Show()
		return
	}
	f := a.frame()
	a.text(f.x+1, 2, "$ MONEY COPY", style(green).Bold(true))
	a.text(f.x+17, 2, "/ IDLE PAYDAY", style(muted))
	a.rule(f.x, 4, f.w)
	switch a.page {
	case intro:
		a.drawIntro()
	case playing:
		a.drawGame()
	case result:
		a.drawResult()
	}
	if a.confirmExit {
		a.drawConfirmation()
	}
	a.screen.Show()
}
func (a *App) drawIntro() {
	f := a.frame()
	y := a.formY()
	badge := "● OFFLINE / 01"
	a.text(f.x+f.w-uniseg.StringWidth(badge)-2, 2, badge, style(muted))
	a.center(6, "가만히 있어도 돈이 복사가 되면 얼마나 좋을까", style(green).Bold(true))
	a.button(a.monthRect(), "월급", !a.annual)
	a.button(a.yearRect(), "연봉", a.annual)
	w, _ := a.screen.Size()
	a.text(w/2-22, y+4, "급여 입력 · 원", style(muted))
	amount := a.amount
	st := style(white).Background(panel)
	if amount == "" {
		amount = "예: 3,600,000"
		st = style(muted).Background(panel)
	}
	a.fill(rect{w/2 - 22, y + 5, 44, 2}, style(white).Background(panel))
	a.text(w/2-20, y+5, amount, st.Bold(true))
	a.screen.ShowCursor(w/2-20+uniseg.StringWidth(a.amount), y+5)
	n, err := strconv.ParseInt(strings.ReplaceAll(a.amount, ",", ""), 10, 64)
	if err == nil {
		if s, e := game.New(n, a.annual, a.now()); e == nil {
			a.center(y+8, "시급 "+decimal(s.HourlyExact())+"원  /  1분마다 "+decimal(s.HourlyExact()/60)+"원", style(green))
		}
	}
	if a.message != "" {
		a.center(y+8, a.message, style(red))
	}
	a.button(a.startRect(), "돈 복사 시작 Enter", true)
	a.center(f.y+f.h-2, "Tab 월급/연봉   Enter 시작   Ctrl+U 입력 지우기   Esc 닫기", style(muted))
}
func (a *App) drawGame() {
	f := a.frame()
	s := a.session
	now := a.now()
	a.button(a.stopRect(), "게임 종료 [Esc]", false)
	a.text(f.x+2, 5, "시급 "+decimal(s.HourlyExact())+"원", style(muted))
	remain := "퇴근까지 " + duration(s.Deadline.Sub(now))
	a.text(f.x+f.w-uniseg.StringWidth(remain)-2, 5, remain, style(muted))
	col := (f.w - 4) / 3
	for i, item := range []struct {
		label string
		value int64
		color tcell.Color
	}{{"기본급", s.Base(), white}, {"성과급", s.Bonus(), gold}, {"오늘 번 돈", s.Total(), green}} {
		x := f.x + 2 + i*col
		a.text(x, 7, item.label, style(muted))
		a.text(x, 8, money(item.value)+" 원", style(item.color).Bold(true))
	}
	interval := game.PayInterval.Seconds()
	progress := max(0, min(interval, now.Sub(s.Started).Seconds()-float64(s.Ticks)*interval))
	cells := int(progress / interval * 12)
	gauge := "┃" + strings.Repeat("█", cells) + strings.Repeat(" ", 12-cells) + "┃"
	gaugeX := f.x + 2
	a.text(gaugeX, 10, gauge, style(green).Background(panel).Bold(true))
	a.text(gaugeX+uniseg.StringWidth(gauge)+2, 10, fmt.Sprintf("다음 입금 %.0fs", math.Ceil(interval-progress)), style(muted))
	a.text(f.x+f.w-24, 10, "1분치 급여 자동 적립", style(muted))
	a.drawScene(now)
	if now.Before(a.toastUntil) {
		a.center(12, a.toast, style(gold).Bold(true))
	} else if s.BonusActive {
		a.center(12, "BONUS DROP · 금색 지폐를 클릭하세요!", style(gold).Bold(true))
	}
}
func (a *App) drawResult() {
	f := a.frame()
	s := a.session
	y := a.formY()
	w, _ := a.screen.Size()
	badge := "■ SETTLED / 03"
	a.text(f.x+f.w-uniseg.StringWidth(badge)-2, 2, badge, style(muted))
	a.center(6, "신난다! 퇴근이다!", style(green).Bold(true))
	reason := "직접 종료"
	if s.Ended.Equal(s.Deadline) {
		reason = "18:00 퇴근 완료"
	}
	a.center(y, reason+"  ·  "+duration(s.Ended.Sub(s.Started)), style(muted))
	for i, item := range []struct {
		label string
		value int64
		color tcell.Color
	}{{"기본급", s.Base(), white}, {fmt.Sprintf("성과급 · %d회", s.BonusCount), s.Bonus(), gold}, {"오늘 번 돈", s.Total(), green}} {
		yy := y + 3 + i*2
		a.text(w/2-22, yy, item.label, style(item.color))
		v := money(item.value) + " 원"
		a.text(w/2+22-uniseg.StringWidth(v), yy, v, style(item.color).Bold(true))
	}
	a.rule(w/2-22, y+6, 44)
	a.button(a.startRect(), "다시 출근하기...", true)
	_, height := a.screen.Size()
	greeting := "고생하셨어요, 내일 봐요!"
	if s.Started.Weekday() == time.Friday {
		greeting = "고생하셨어요, 다음 주에 봐요!"
	}
	greetingY := min(y+14, height-2)
	if greetingY == f.y+f.h-2 {
		greetingY = height - 2
	}
	a.center(greetingY, greeting, style(white))
	a.center(f.y+f.h-2, "Enter 새로 시작   Esc 닫기", style(muted))
}

// Bill dimensions are terminal columns and rows, including the border.
func billSize(big bool) (int, int) {
	if big {
		return 11, 3
	}
	return 8, 3
}
func (a *App) bill(x, y int, big bool, tint int) {
	w, h := billSize(big)
	color := green
	if tint == 1 {
		color = muted
	}
	if big {
		color = gold
	}
	st := style(color)
	a.fill(rect{x, y, w, h}, st)
	a.text(x, y, "┏"+strings.Repeat("━", w-2)+"┓", st)
	a.text(x, y+1, "┃"+strings.Repeat(" ", w-2)+"┃", st)
	a.text(x, y+2, "┗"+strings.Repeat("━", w-2)+"┛", st)
	label := "₩"
	if big {
		label = "₩ +30m"
	}
	a.text(x+(w-uniseg.StringWidth(label))/2, y+1, label, st.Bold(big))
}
func (a *App) drawScene(now time.Time) {
	f := a.frame()
	floor := a.floor()
	noteW, noteH := billSize(false)
	_, h := a.screen.Size()
	for y := 14; y < floor; y += 4 {
		for x := f.x + 4 + (y%3)*3; x < f.x+f.w-3; x += 13 {
			a.text(x, y, "·", style(line))
		}
	}
	for _, n := range a.notes {
		p := max(0, min(1, now.Sub(n.born).Seconds()/n.duration.Seconds()))
		x := f.x + 2 + int(n.x*float64(max(0, f.w-noteW-4)))
		y := 13 + int(p*float64(max(1, floor-13-noteH)))
		a.bill(x, y, false, 0)
	}
	// Layered, deterministic stacks stay stable between frames.
	for row := h - 5; row >= floor; row -= 2 {
		for x := f.x + 1 + row%2; x+noteW <= f.x+f.w-1; x += noteW {
			a.bill(x, row, false, 1)
		}
	}
	for x := f.x + 2; x+noteW <= f.x+f.w-1; x += noteW + 2 {
		a.bill(x, floor+(x%3)/2, false, 0)
	}
	if a.session.BonusActive {
		r := a.bonusRect(now)
		a.bill(r.x, r.y, true, 0)
	}
}
