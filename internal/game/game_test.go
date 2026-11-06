package game

import (
	"testing"
	"time"
)

func morning() time.Time { return time.Date(2026, 11, 6, 9, 0, 0, 0, time.FixedZone("KST", 9*3600)) }
func TestSalaryAndMinuteAccrual(t *testing.T) {
	for _, annual := range []bool{false, true} {
		salary := int64(3600000)
		if annual {
			salary *= 12
		}
		s, err := New(salary, annual, morning())
		if err != nil {
			t.Fatal(err)
		}
		if s.Monthly() != 3600000 || s.Hourly() != 20000 {
			t.Fatalf("conversion: %+v", s)
		}
		s.Advance(morning().Add(59999 * time.Millisecond))
		if s.Base() != 0 {
			t.Fatal("paid before 60 seconds")
		}
		s.Advance(morning().Add(60 * time.Second))
		if s.Base() != 333 {
			t.Fatal(s.Base())
		}
		s.Advance(morning().Add(190 * time.Second))
		if s.Base() != 1000 {
			t.Fatal("missed elapsed ticks", s.Base())
		}
		s.Advance(morning().Add(70 * time.Second))
		if s.Base() != 1000 {
			t.Fatal("clock reversal removed money")
		}
	}
}
func TestFractionalAccrualDoesNotTruncateEachTick(t *testing.T) {
	s, _ := New(1000, false, morning())
	s.Advance(morning().Add(3600 * time.Second))
	if s.Base() != 5 {
		t.Fatalf("want floor(1000/180)=5, got %d", s.Base())
	}
	a, _ := New(1001, true, morning())
	a.Advance(morning().Add(9 * time.Hour))
	if a.Base() != 4 {
		t.Fatal("annual conversion rounded early", a.Base())
	}
}
func TestInvalidSalary(t *testing.T) {
	for _, n := range []int64{-1, 0, 1000000000001} {
		if _, err := New(n, false, morning()); err == nil {
			t.Fatalf("accepted %d", n)
		}
	}
	s, err := New(1000000000000, false, morning())
	if err != nil {
		t.Fatal(err)
	}
	s.Advance(s.Deadline)
	if s.Base() != 50000000000 {
		t.Fatal("large amount overflow", s.Base())
	}
}
func TestDeadlineAndLateResume(t *testing.T) {
	start := time.Date(2026, 11, 6, 17, 58, 48, 0, time.UTC)
	s, _ := New(3600000, false, start)
	s.Advance(start.Add(time.Hour))
	if !s.Finished || s.Base() != 333 || !s.Ended.Equal(s.Deadline) {
		t.Fatalf("deadline not capped: %+v", s)
	}
	s.Advance(start.Add(2 * time.Hour))
	if s.Base() != 333 {
		t.Fatal("paid after finish")
	}
	for _, hour := range []int{18, 22} {
		late := time.Date(2026, 11, 6, hour, 0, 0, 0, time.UTC)
		n, _ := New(1, false, late)
		if n.Deadline.Day() != 7 || n.Deadline.Hour() != 18 {
			t.Fatal(n.Deadline)
		}
	}
}
func TestBonusOnlyOnceAndBeforeLanding(t *testing.T) {
	s, _ := New(43200000, true, morning())
	if s.ClaimBonus(morning()) {
		t.Fatal("claimed nonexistent bonus")
	}
	if !s.SpawnBonus(morning(), 10*time.Second) {
		t.Fatal("spawn failed")
	}
	if !s.ClaimBonus(morning().Add(9*time.Second)) || s.Bonus() != 10000 {
		t.Fatal("bonus not paid", s.Bonus())
	}
	if s.ClaimBonus(morning().Add(9 * time.Second)) {
		t.Fatal("double claim")
	}
	s.SpawnBonus(morning().Add(10*time.Second), 10*time.Second)
	if s.ClaimBonus(morning().Add(20 * time.Second)) {
		t.Fatal("claimed landed bonus")
	}
	s.Finish(morning().Add(21 * time.Second))
	if s.Base() != 0 || s.Total() != 10000 {
		t.Fatal("wrong result", s.Total())
	}
	if s.SpawnBonus(morning().Add(22*time.Second), time.Second) {
		t.Fatal("spawned after end")
	}
}
func TestFinishFreezesAndBonusCannotCrossDeadline(t *testing.T) {
	s, _ := New(3600000, false, morning())
	s.Finish(morning().Add(65 * time.Second))
	s.Finish(morning().Add(140 * time.Second))
	if s.Base() != 333 || !s.Ended.Equal(morning().Add(65*time.Second)) {
		t.Fatal("finish not idempotent")
	}
	n, _ := New(3600000, false, morning())
	n.SpawnBonus(n.Deadline.Add(-time.Second), 10*time.Second)
	if n.ClaimBonus(n.Deadline) || !n.Finished {
		t.Fatal("claimed after cutoff")
	}
}

func TestSessionUsesWallClockAcrossSleep(t *testing.T) {
	now := time.Now()
	morning := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	now = now.Add(morning.Sub(now))
	s, _ := New(3600000, false, now)
	if s.Started != now.Round(0) {
		t.Fatal("session retained suspend-sensitive monotonic clock")
	}
	s.SpawnBonus(now, 10*time.Second)
	if s.BonusExpires != now.Add(10*time.Second).Round(0) {
		t.Fatal("bonus retained monotonic clock")
	}
	s.Advance(now.Round(0).Add(time.Minute))
	if s.BonusActive || s.Base() != 333 {
		t.Fatal("wall clock resume did not expire bonus and credit time")
	}
}
