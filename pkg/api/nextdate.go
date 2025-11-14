package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "20060102"

// NextDate вычисляет следующую дату в формате 20060102 на основе now, dstart и repeat.
// Поддерживаются базовые правила: d <число> и y. Правила w/m не поддерживаются и возвращают ошибку.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	dstart = strings.TrimSpace(dstart)
	repeat = strings.TrimSpace(repeat)

	if repeat == "" {
		return "", fmt.Errorf("empty repeat")
	}

	// разобрать dstart
	ds, err := time.Parse(dateLayout, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid dstart: %w", err)
	}

	// разбор правила
	var mode string
	var interval int

	if repeat == "y" {
		mode = "y"
	} else {
		// базовый вариант: d <число> (с пробелом или без)
		parts := []string{}
		parts = append(parts, strings.Fields(repeat)...)

		first := parts[0]
		if strings.HasPrefix(first, "d") {
			// dN или d N
			numStr := strings.TrimPrefix(first, "d")
			if numStr == "" {
				// возможно, второй элемент содержит число
				if len(parts) > 1 {
					numStr = parts[1]
				} else {
					return "", fmt.Errorf("invalid days interval")
				}
			} else if len(parts) > 1 {
				// если есть второй элемент, он может являться интервалом
				numStr = parts[1]
			}
			iv, err := strconv.Atoi(numStr)
			if err != nil {
				return "", fmt.Errorf("invalid days interval: %w", err)
			}
			if iv < 1 || iv > 400 {
				return "", fmt.Errorf("days interval out of range")
			}
			interval = iv
			mode = "d"
		} else {
			return "", fmt.Errorf("unsupported format")
		}
	}
	// вычислить следующую дату
	date := ds
	switch mode {
	case "y":
		for !date.After(now) {
			date = date.AddDate(1, 0, 0)
		}
	case "d":
		for !date.After(now) {
			date = date.AddDate(0, 0, interval)
		}
	default:
		return "", fmt.Errorf("unsupported mode")
	}
	return date.Format(dateLayout), nil
}

// nextDateHandler обрабатывает запрос /api/nextdate
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	var now time.Time
	var err error
	if nowParam == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateLayout, nowParam)
		if err != nil {
			http.Error(w, "invalid now date", http.StatusBadRequest)
			return
		}
	}

	dstart := r.FormValue("date")
	repeat := r.FormValue("repeat")
	res, err := NextDate(now, dstart, repeat)
	if err != nil {
		// вернуть текст ошибки
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(res))
}
