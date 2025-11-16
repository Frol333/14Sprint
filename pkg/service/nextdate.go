package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "20060102"

func NextDate(current time.Time, repeat string) (time.Time, error) {
	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return current, fmt.Errorf("repeat should not be empty")
	}

	switch parts[0] {

	// ---------- DAILY ----------
	case "d":
		if len(parts) < 2 {
			return current, fmt.Errorf("daily repeat should contain interval")
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil {
			return current, err
		}
		if n <= 0 || n > 400 {
			return current, fmt.Errorf("daily repeat interval should be between 1 and 400")
		}
		return current.AddDate(0, 0, n), nil

	// ---------- YEARLY ----------
	case "y":
		return current.AddDate(1, 0, 0), nil

	// ---------- MONTHLY WITH DAY LIST ----------
	case "m":
		if len(parts) < 2 {
			return current, fmt.Errorf("monthly repeat should contain day list")
		}
		// parts[1] = "1,2"
		daysStr := parts[1]
		dayList := parseIntList(daysStr)

		// Если указан список месяцев
		monthList := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		if len(parts) >= 3 {
			monthList = parseIntList(parts[2])
		}

		// Ищем следующую подходящую дату
		return nextMonthlyDate(current, dayList, monthList), nil
	}

	return current, fmt.Errorf("unknown repeat: %s", repeat)
}

func parseIntList(s string) []int {
	parts := strings.Split(s, ",")
	res := make([]int, 0, len(parts))

	for _, p := range parts {
		if v, err := strconv.Atoi(p); err == nil {
			res = append(res, v)
		}
	}
	return res
}

func nextMonthlyDate(current time.Time, days []int, months []int) time.Time {
	sort.Ints(days)
	sort.Ints(months)

	y, m, _ := current.Date()
	loc := current.Location()

	for {
		// Проверяем, подходит ли месяц
		if contains(months, int(m)) {
			// Проверяем дни
			for _, d := range days {
				// Если день >= текущего — это кандидат
				candidate := time.Date(y, m, d, current.Hour(), current.Minute(), current.Second(), 0, loc)
				if candidate.After(current) {
					return candidate
				}
			}
		}

		// Переходим к следующему месяцу
		t := time.Date(y, m, 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0)
		y, m = t.Year(), t.Month()
	}
}

func contains(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
