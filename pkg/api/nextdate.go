package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "20060102"

// NextDate вычисляет следующую дату после параметра now согласно правилу repeat.
// dstart — исходная дата в формате 20060102.
// repeat — базовые правила: "d " или "y".
// Возвращает дату в формате 20060102 или ошибку.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	// Преобразование dstart
	start, err := time.Parse(dateLayout, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid start date: %w", err)
	}

	rep := strings.TrimSpace(repeat)
	if rep == "" {
		return "", fmt.Errorf("empty repeat")
	}

	// Разбираем повторы. Сейчас реализуем базовые: d и y
	parts := strings.Fields(rep) // разделяем по пробелам: например "d 7"
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid repeat format")
	}
	rule := parts[0]

	switch rule {
	case "d":
		// d <число>
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid repeat format")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", fmt.Errorf("invalid days: %w", err)
		}
		if days <= 0 || days > 400 {
			return "", fmt.Errorf("days out of range (1..400)")
		}
		date := start
		for {
			date = date.AddDate(0, 0, days)
			if date.After(now) {
				break
			}
		}
		return date.Format(dateLayout), nil

	case "y":
		// ежегодно
		date := start
		for {
			date = date.AddDate(1, 0, 0)
			if date.After(now) {
				break
			}
		}
		return date.Format(dateLayout), nil

	default:
		// Пока не поддерживаем другие правила (w, m и т.д.)
		return "", fmt.Errorf("unsupported repeat format: %s", rule)
	}

}
