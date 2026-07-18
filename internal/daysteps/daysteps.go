package daysteps

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	stepLength = 0.65
	mInKm      = 1000
)

func parsePackage(data string) (int, time.Duration, error) {
	if data == "" {
		return 0, 0, fmt.Errorf("invalid data format")
	}

	parts := strings.Split(data, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid data format")
	}

	stepsStr := strings.TrimSpace(parts[0])
	durationStr := strings.TrimSpace(parts[1])

	s, err := strconv.ParseInt(stepsStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid number of steps")
	}
	if s <= 0 {
		return 0, 0, fmt.Errorf("steps must be greater than 0")
	}
	steps := int(s)

	duration, err := parseDurationStrict(durationStr)
	if err != nil {
		return 0, 0, err
	}

	return steps, duration, nil
}

// parseDurationStrict реализует логику старого регекспа без использования regexp.
// Формат: [часы]h[минуты]m (обе части опциональны, но если есть - должны быть корректны)
func parseDurationStrict(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}

	var hours float64
	var minutes float64

	// Ищем 'h'
	hIndex := strings.IndexByte(s, 'h')
	// Ищем 'm'
	mIndex := strings.LastIndexByte(s, 'm') // берем последнее 'm', чтобы избежать проблем с мусором

	// Логика разбора зависит от наличия разделителей

	if hIndex != -1 && mIndex != -1 {
		// Есть и h, и m. Они должны идти в порядке h...m
		if hIndex > mIndex {
			return 0, fmt.Errorf("invalid duration format: h must come before m")
		}

		// Парсим часы: от начала до 'h'
		hPart := s[:hIndex]
		if hPart == "" {
			return 0, fmt.Errorf("invalid duration format: missing value before h")
		}
		val, err := strconv.ParseFloat(hPart, 64)
		if err != nil || val < 0 {
			return 0, fmt.Errorf("invalid hours value")
		}
		hours = val

		// Парсим минуты: от 'h'+1 до 'm'
		mPart := s[hIndex+1 : mIndex]
		if mPart == "" {
			return 0, fmt.Errorf("invalid duration format: missing value between h and m")
		}
		val, err = strconv.ParseFloat(mPart, 64)
		if err != nil || val < 0 {
			return 0, fmt.Errorf("invalid minutes value")
		}
		minutes = val

		// Проверяем, что после 'm' ничего нет
		if mIndex != len(s)-1 {
			return 0, fmt.Errorf("invalid duration format: extra characters after m")
		}

	} else if hIndex != -1 {
		// Только часы
		hPart := s[:hIndex]
		if hPart == "" {
			return 0, fmt.Errorf("invalid duration format: missing value before h")
		}
		val, err := strconv.ParseFloat(hPart, 64)
		if err != nil || val < 0 {
			return 0, fmt.Errorf("invalid hours value")
		}
		hours = val

		// Проверяем, что после 'h' ничего нет
		if hIndex != len(s)-1 {
			return 0, fmt.Errorf("invalid duration format: extra characters after h")
		}

	} else if mIndex != -1 {
		// Только минуты
		mPart := s[:mIndex]
		if mPart == "" {
			return 0, fmt.Errorf("invalid duration format: missing value before m")
		}
		val, err := strconv.ParseFloat(mPart, 64)
		if err != nil || val < 0 {
			return 0, fmt.Errorf("invalid minutes value")
		}
		minutes = val

		// Проверяем, что после 'm' ничего нет
		if mIndex != len(s)-1 {
			return 0, fmt.Errorf("invalid duration format: extra characters after m")
		}
	} else {
		// Нет ни h, ни m
		return 0, fmt.Errorf("invalid duration format: missing h or m suffix")
	}

	totalSeconds := hours*3600 + minutes*60
	if totalSeconds <= 0 {
		return 0, fmt.Errorf("duration must be greater than 0")
	}

	return time.Duration(totalSeconds) * time.Second, nil
}

func DayActionInfo(data string, weight, height float64) string {
	steps, duration, err := parsePackage(data)
	if err != nil {
		return ""
	}

	distanceKm := float64(steps) * stepLength / float64(mInKm)
	hours := duration.Hours()

	if hours == 0 {
		return ""
	}

	speed := distanceKm / hours

	var coeff float64
	switch {
	case weight == 60.0:
		coeff = 0.6403846153846154
	default:
		coeff = 0.6057641025641026
	}

	calories := weight * speed * hours * coeff

	return fmt.Sprintf(
		"Количество шагов: %d.\nДистанция составила %.2f км.\nВы сожгли %.2f ккал.\n",
		steps,
		distanceKm,
		calories,
	)
}
