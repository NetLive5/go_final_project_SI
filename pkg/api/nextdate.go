package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func NextDate(now time.Time, dstart, repeat string) (string, error) {
	start, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid start date %q: %w", dstart, err)
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", fmt.Errorf("repeat rule must not be empty")
	}

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", fmt.Errorf("daily rule must be 'd <days>'")
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval < 1 || interval > 400 {
			return "", fmt.Errorf("daily interval must be between 1 and 400")
		}
		for {
			start = start.AddDate(0, 0, interval)
			if dateAfter(start, now) {
				return start.Format(dateFormat), nil
			}
		}

	case "y":
		if len(parts) != 1 {
			return "", fmt.Errorf("yearly rule must be 'y'")
		}
		for {
			start = start.AddDate(1, 0, 0)
			if dateAfter(start, now) {
				return start.Format(dateFormat), nil
			}
		}

	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("weekly rule must be 'w <weekdays>'")
		}
		weekdays, err := parseList(parts[1], 1, 7, false)
		if err != nil {
			return "", fmt.Errorf("invalid weekly rule: %w", err)
		}
		for candidate := start.AddDate(0, 0, 1); ; candidate = candidate.AddDate(0, 0, 1) {
			if dateAfter(candidate, now) && weekdays[weekdayNumber(candidate.Weekday())] {
				return candidate.Format(dateFormat), nil
			}
		}

	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", fmt.Errorf("monthly rule must be 'm <days> [months]'")
		}
		days, err := parseList(parts[1], -2, 31, true)
		if err != nil {
			return "", fmt.Errorf("invalid monthly rule: %w", err)
		}
		months := make(map[int]bool)
		if len(parts) == 3 {
			months, err = parseList(parts[2], 1, 12, false)
			if err != nil {
				return "", fmt.Errorf("invalid monthly rule: %w", err)
			}
		}
		for candidate := start.AddDate(0, 0, 1); ; candidate = candidate.AddDate(0, 0, 1) {
			if !dateAfter(candidate, now) {
				continue
			}
			if len(months) > 0 && !months[int(candidate.Month())] {
				continue
			}
			lastDay := time.Date(candidate.Year(), candidate.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			matches := days[candidate.Day()] || (days[-1] && candidate.Day() == lastDay) || (days[-2] && candidate.Day() == lastDay-1)
			if matches {
				return candidate.Format(dateFormat), nil
			}
		}

	default:
		return "", fmt.Errorf("unsupported repeat rule %q", parts[0])
	}
}

func parseList(value string, min, max int, allowLastDays bool) (map[int]bool, error) {
	items := strings.Split(value, ",")
	values := make(map[int]bool, len(items))
	for _, item := range items {
		if item == "" {
			return nil, fmt.Errorf("list contains an empty value")
		}
		n, err := strconv.Atoi(item)
		if err != nil || n < min || n > max || (!allowLastDays && n < 1) || (allowLastDays && n < 1 && n != -1 && n != -2) {
			return nil, fmt.Errorf("value %q is out of range", item)
		}
		values[n] = true
	}
	return values, nil
}

func dateAfter(date, now time.Time) bool {
	return date.Format(dateFormat) > now.Format(dateFormat)
}

func weekdayNumber(day time.Weekday) int {
	if day == time.Sunday {
		return 7
	}
	return int(day)
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	now := time.Now()
	if value := r.FormValue("now"); value != "" {
		parsed, err := time.Parse(dateFormat, value)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid now date: %v", err), http.StatusBadRequest)
			return
		}
		now = parsed
	}

	next, err := NextDate(now, r.FormValue("date"), r.FormValue("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, _ = w.Write([]byte(next))
}
