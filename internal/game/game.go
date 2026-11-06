package game

import (
	"errors"
	"time"
)

// Session owns salary, elapsed-time accrual and one collectible bonus.
type Session struct {
	Started                    time.Time
	Deadline, Ended            time.Time
	Finished                   bool
	Salary                     int64
	Annual                     bool
	Ticks, BonusCount          int64
	BonusStarted, BonusExpires time.Time
	BonusActive                bool
	monthDiv                   int64
}

func New(salary int64, annual bool, now time.Time) (*Session, error) {
	now = now.Round(0) // Wall time includes system sleep, including on macOS.
	if salary <= 0 || salary > 1_000_000_000_000 {
		return nil, errors.New("1원부터 1조 원까지 숫자로 입력해 주세요")
	}
	deadline := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
	if !now.Before(deadline) {
		deadline = deadline.AddDate(0, 0, 1)
	}
	div := int64(1)
	if annual {
		div = 12
	}
	return &Session{Started: now, Deadline: deadline, Salary: salary, Annual: annual, monthDiv: div}, nil
}
func (s *Session) Monthly() int64       { return s.Salary / s.monthDiv }
func (s *Session) Hourly() int64        { return s.Salary / (s.monthDiv * 180) }
func (s *Session) HourlyExact() float64 { return float64(s.Salary) / float64(s.monthDiv*180) }

// Keep the numerator exact until display; never truncate each five-second payment.
func (s *Session) Base() int64  { return s.Salary * s.Ticks / (s.monthDiv * 180 * 720) }
func (s *Session) Bonus() int64 { return s.Salary * s.BonusCount / (s.monthDiv * 360) }
func (s *Session) Total() int64 { return s.Base() + s.Bonus() }
func (s *Session) Advance(now time.Time) {
	now = now.Round(0)
	if s.Finished {
		return
	}
	cutoff := now
	if cutoff.After(s.Deadline) {
		cutoff = s.Deadline
	}
	ticks := int64(cutoff.Sub(s.Started) / (5 * time.Second))
	if ticks > s.Ticks {
		s.Ticks = ticks
	}
	if !now.Before(s.BonusExpires) {
		s.BonusActive = false
	}
	if !now.Before(s.Deadline) {
		s.Finished = true
		s.Ended = s.Deadline
		s.BonusActive = false
	}
}
func (s *Session) Finish(now time.Time) {
	now = now.Round(0)
	s.Advance(now)
	if s.Finished {
		return
	}
	s.Finished = true
	s.Ended = now
	s.BonusActive = false
	if s.Ended.Before(s.Started) {
		s.Ended = s.Started
	}
}
func (s *Session) SpawnBonus(now time.Time, duration time.Duration) bool {
	now = now.Round(0)
	s.Advance(now)
	if s.Finished || s.BonusActive || duration <= 0 {
		return false
	}
	s.BonusActive = true
	s.BonusStarted = now
	s.BonusExpires = now.Add(duration)
	return true
}
func (s *Session) ClaimBonus(now time.Time) bool {
	now = now.Round(0)
	s.Advance(now)
	if s.Finished || !s.BonusActive || now.Before(s.BonusStarted) {
		return false
	}
	s.BonusActive = false
	s.BonusCount++
	return true
}
