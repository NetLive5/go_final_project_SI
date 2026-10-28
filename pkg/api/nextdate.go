package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

// NextDate возвращает первую дату после now и dstart, соответствующую правилу repeat.
func NextDate(now time.Time, dstart, repeat string) (string, error) {
	start, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("неверная начальная дата %q: ожидается формат YYYYMMDD", dstart)
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", fmt.Errorf("правило повторения не должно быть пустым")
	}

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", fmt.Errorf("для повтора по дням используйте формат 'd <количество дней>'")
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval < 1 || interval > 400 {
			return "", fmt.Errorf("интервал повтора должен быть от 1 до 400 дней")
		}
		for {
			start = start.AddDate(0, 0, interval)
			if dateAfter(start, now) {
				return start.Format(dateFormat), nil
			}
		}

	case "y":
		if len(parts) != 1 {
			return "", fmt.Errorf("для ежегодного повтора используйте правило 'y'")
		}
		for {
			start = start.AddDate(1, 0, 0)
			if dateAfter(start, now) {
				return start.Format(dateFormat), nil
			}
		}

	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("для повтора по дням недели используйте формат 'w <дни недели>'")
		}
		weekdays, err := parseList(parts[1], 1, 7, false)
		if err != nil {
			return "", fmt.Errorf("неверное правило повтора по дням недели: %w", err)
		}
		for candidate := start.AddDate(0, 0, 1); ; candidate = candidate.AddDate(0, 0, 1) {
			if dateAfter(candidate, now) && weekdays[weekdayNumber(candidate.Weekday())] {
				return candidate.Format(dateFormat), nil
			}
		}

	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", fmt.Errorf("для повтора по месяцам используйте формат 'm <дни месяца> [месяцы]'")
		}
		days, err := parseList(parts[1], -2, 31, true)
		if err != nil {
			return "", fmt.Errorf("неверное правило повтора по месяцам: %w", err)
		}
		months := make(map[int]bool)
		if len(parts) == 3 {
			months, err = parseList(parts[2], 1, 12, false)
			if err != nil {
				return "", fmt.Errorf("неверное правило повтора по месяцам: %w", err)
			}
		}
		if !monthlyRuleHasDate(days, months) {
			return "", fmt.Errorf("для правила повтора не существует подходящей даты")
		}
		searchFrom := start
		if dateAfter(now, start) {
			searchFrom = now
		}
		// Григорианский календарь повторяется каждые 400 лет. Если подходящей
		// даты нет за полный цикл, это правило не сможет её породить.
		searchUntil := searchFrom.AddDate(400, 0, 0)
		for candidate := searchFrom.AddDate(0, 0, 1); candidate.Before(searchUntil); candidate = candidate.AddDate(0, 0, 1) {
			if len(months) > 0 && !months[int(candidate.Month())] {
				continue
			}
			lastDay := time.Date(candidate.Year(), candidate.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			// -1 обозначает последний день месяца, -2 — предпоследний.
			matches := days[candidate.Day()] || (days[-1] && candidate.Day() == lastDay) || (days[-2] && candidate.Day() == lastDay-1)
			if matches {
				return candidate.Format(dateFormat), nil
			}
		}
		return "", fmt.Errorf("для правила повтора не найдено подходящей даты")

	default:
		return "", fmt.Errorf("неподдерживаемое правило повтора %q", parts[0])
	}
}

func monthlyRuleHasDate(days, months map[int]bool) bool {
	for month := 1; month <= 12; month++ {
		if len(months) > 0 && !months[month] {
			continue
		}
		// Берём високосный год, чтобы учитывать возможность 29 февраля.
		lastDay := time.Date(2000, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
		for day := range days {
			if day == -1 || day == -2 || (day > 0 && day <= lastDay) {
				return true
			}
		}
	}
	return false
}

func parseList(value string, min, max int, allowLastDays bool) (map[int]bool, error) {
	items := strings.Split(value, ",")
	values := make(map[int]bool, len(items))
	for _, item := range items {
		if item == "" {
			return nil, fmt.Errorf("список содержит пустое значение")
		}
		n, err := strconv.Atoi(item)
		if err != nil || n < min || n > max || (!allowLastDays && n < 1) || (allowLastDays && n < 1 && n != -1 && n != -2) {
			return nil, fmt.Errorf("значение %q вне допустимого диапазона", item)
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
		// В правилах повторения дни недели нумеруются с понедельника (1) по воскресенье (7).
		return 7
	}
	return int(day)
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	now := time.Now()
	if value := r.FormValue("now"); value != "" {
		parsed, err := time.Parse(dateFormat, value)
		if err != nil {
			http.Error(w, "неверная текущая дата: ожидается формат YYYYMMDD", http.StatusBadRequest)
			return
		}
		now = parsed
	}

	next, err := NextDate(now, r.FormValue("date"), r.FormValue("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := w.Write([]byte(next)); err != nil {
		log.Printf("ошибка записи даты следующего выполнения: %v", err)
	}
}
