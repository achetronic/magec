package trigger

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// cronSchedule represents a parsed standard cron expression (min hour dom month dow).
type cronSchedule struct {
	minutes  [60]bool
	hours    [24]bool
	days     [31]bool
	months   [12]bool
	weekdays [7]bool
}

// parseCron parses a standard 5-field cron expression.
func parseCron(expr string) (*cronSchedule, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("expected 5 fields, got %d", len(fields))
	}

	s := &cronSchedule{}

	if err := parseField(fields[0], s.minutes[:], 0, 59); err != nil {
		return nil, fmt.Errorf("minute: %w", err)
	}
	if err := parseField(fields[1], s.hours[:], 0, 23); err != nil {
		return nil, fmt.Errorf("hour: %w", err)
	}
	if err := parseField(fields[2], s.days[:], 1, 31); err != nil {
		return nil, fmt.Errorf("day: %w", err)
	}
	if err := parseField(fields[3], s.months[:], 1, 12); err != nil {
		return nil, fmt.Errorf("month: %w", err)
	}
	if err := parseField(fields[4], s.weekdays[:], 0, 6); err != nil {
		return nil, fmt.Errorf("weekday: %w", err)
	}

	return s, nil
}

func parseField(field string, bits []bool, min, max int) error {
	for _, part := range strings.Split(field, ",") {
		step := 1
		if idx := strings.Index(part, "/"); idx >= 0 {
			var err error
			step, err = strconv.Atoi(part[idx+1:])
			if err != nil || step <= 0 {
				return fmt.Errorf("invalid step %q", part[idx+1:])
			}
			part = part[:idx]
		}

		var lo, hi int
		if part == "*" {
			lo, hi = min, max
		} else if idx := strings.Index(part, "-"); idx >= 0 {
			var err error
			lo, err = strconv.Atoi(part[:idx])
			if err != nil {
				return fmt.Errorf("invalid range start %q", part[:idx])
			}
			hi, err = strconv.Atoi(part[idx+1:])
			if err != nil {
				return fmt.Errorf("invalid range end %q", part[idx+1:])
			}
		} else {
			val, err := strconv.Atoi(part)
			if err != nil {
				return fmt.Errorf("invalid value %q", part)
			}
			lo, hi = val, val
		}

		if lo < min || hi > max || lo > hi {
			return fmt.Errorf("out of range: %d-%d (allowed %d-%d)", lo, hi, min, max)
		}

		offset := min
		for i := lo; i <= hi; i += step {
			bits[i-offset] = true
		}
	}
	return nil
}

// Next returns the next time after t that matches the schedule.
func (s *cronSchedule) Next(t time.Time) time.Time {
	t = t.Add(time.Minute).Truncate(time.Minute)

	for i := 0; i < 366*24*60; i++ {
		if s.matches(t) {
			return t
		}
		t = t.Add(time.Minute)
	}

	return t
}

func (s *cronSchedule) matches(t time.Time) bool {
	return s.minutes[t.Minute()] &&
		s.hours[t.Hour()] &&
		s.days[t.Day()-1] &&
		s.months[t.Month()-1] &&
		s.weekdays[t.Weekday()]
}
