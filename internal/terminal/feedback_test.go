package terminal

import (
	"github.com/gdamore/tcell/v2"
	"strings"
	"testing"
	"time"
)

func TestIntroExitRequiresConfirmation(t *testing.T) {
	a, s, _ := setup(t)
	input(a, "3600000")
	key(a, tcell.KeyEscape)
	a.Draw()
	if a.quit || !strings.Contains(dump(s), "오후 반차 내시는 겁니까?") {
		t.Fatal("missing intro confirmation")
	}
	key(a, tcell.KeyEnter)
	if a.quit || a.page != intro {
		t.Fatal("default choice must cancel")
	}
	if a.amount != "3600000" {
		t.Fatal("cancel lost salary")
	}
	key(a, tcell.KeyEscape)
	input(a, "y")
	if !a.quit {
		t.Fatal("confirm did not exit")
	}
}
func TestGameExitCancelAndConfirm(t *testing.T) {
	a, s, now := setup(t)
	input(a, "3600000")
	key(a, tcell.KeyEnter)
	key(a, tcell.KeyEscape)
	a.Draw()
	if a.page != playing || !strings.Contains(dump(s), "퇴근하시겠습니까?") {
		t.Fatal("settled before confirmation")
	}
	*now = now.Add(time.Minute)
	a.Tick()
	if a.session.Base() != 333 {
		t.Fatal("dialog paused salary")
	}
	key(a, tcell.KeyEscape)
	if a.page != playing {
		t.Fatal("escape must cancel")
	}
	key(a, tcell.KeyEscape)
	key(a, tcell.KeyTab)
	key(a, tcell.KeyEnter)
	if a.page != result || a.session.Base() != 333 {
		t.Fatal("confirm did not settle")
	}
}
func TestDeadlineClosesPendingConfirmation(t *testing.T) {
	a, s, now := setup(t)
	input(a, "3600000")
	key(a, tcell.KeyEnter)
	key(a, tcell.KeyEscape)
	*now = a.session.Deadline
	a.Tick()
	a.Draw()
	if a.page != result || strings.Contains(dump(s), "퇴근하시겠습니까?") {
		t.Fatal("dialog outlived deadline")
	}
	key(a, tcell.KeyEnter)
	if a.page != intro {
		t.Fatal("stale dialog intercepted result input")
	}
}
func TestSmallerBonusHasMatchingHitbox(t *testing.T) {
	a, _, now := setup(t)
	input(a, "3600000")
	key(a, tcell.KeyEnter)
	*now = a.nextBonus
	a.Tick()
	r := a.bonusRect(*now)
	// Clicking in the old bill's right half must no longer collect the bonus.
	a.Handle(tcell.NewEventMouse(r.x+12, r.y+1, tcell.Button1, 0))
	if a.session.Bonus() != 0 {
		t.Fatal("old oversized hitbox still active")
	}
	a.Handle(tcell.NewEventMouse(0, 0, tcell.ButtonNone, 0))
	a.Handle(tcell.NewEventMouse(r.x+5, r.y+1, tcell.Button1, 0))
	if a.session.Bonus() != 10000 {
		t.Fatal("small bill could not be collected")
	}
	if a.toast != "성과급 10,000원 지급 완료!" {
		t.Fatal(a.toast)
	}
}
func TestGaugeAndNextPaymentDoNotOverlap(t *testing.T) {
	for _, size := range []int{76, 80, 110} {
		a, s, now := setup(t)
		s.SetSize(size, 24)
		input(a, "3600000")
		key(a, tcell.KeyEnter)
		*now = now.Add(59 * time.Second)
		a.Tick()
		a.Draw()
		row := strings.Split(dump(s), "\n")[10]
		close := strings.Index(row, "]")
		label := strings.Index(row, "다음 입금")
		if close < 0 || label < close+2 || !strings.Contains(row, "1s") {
			t.Fatalf("overlap at width %d: %s", size, row)
		}
	}
}
func TestResultGreetingUsesStartWeekday(t *testing.T) {
	for _, day := range []struct {
		offset   int
		greeting string
	}{{0, "다음 주에 봐요!"}, {3, "내일 봐요!"}, {1, "내일 봐요!"}} {
		a, s, now := setup(t)
		*now = now.AddDate(0, 0, day.offset)
		input(a, "3600000")
		key(a, tcell.KeyEnter)
		*now = now.Add(time.Minute)
		key(a, tcell.KeyEscape)
		input(a, "y")
		a.Draw()
		if !strings.Contains(dump(s), "고생하셨어요, "+day.greeting) {
			t.Fatal("wrong weekday greeting", dump(s))
		}
	}
}
