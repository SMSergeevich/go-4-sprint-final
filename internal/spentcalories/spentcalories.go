package spentcalories

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidFormat   = errors.New("invalid data format")
	ErrUnknownActivity = errors.New("unknown activity type")
	ErrInvalidSteps    = errors.New("invalid number of steps")
	ErrInvalidDuration = errors.New("invalid duration")
	ErrInvalidHeight   = errors.New("invalid height")
)

const (
	lenStep                    = 0.65
	mInKm                      = 1000
	stepLengthCoefficient      = 0.45
	walkingCaloriesCoefficient = 0.5
	runningCaloriesCoefficient = 1.0
)

// parseTraining разбирает строку вида "1000,Ходьба,1h30m"
func parseTraining(data string) (int, string, time.Duration, error) {
	if data == "" {
		return 0, "", 0, ErrInvalidFormat
	}

	parts := strings.Split(data, ",")
	if len(parts) != 3 {
		return 0, "", 0, ErrInvalidFormat
	}

	// TrimSpace нужен, если тесты допускают пробелы вокруг значений.
	stepsStr := strings.TrimSpace(parts[0])
	activity := strings.TrimSpace(parts[1])
	durationStr := strings.TrimSpace(parts[2])

	s, err := strconv.ParseInt(stepsStr, 10, 64)
	if err != nil {
		return 0, "", 0, ErrInvalidSteps
	}
	if s <= 0 {
		return 0, "", 0, ErrInvalidSteps
	}
	steps := int(s)

	duration, err := parseDuration(durationStr)
	if err != nil {
		return 0, "", 0, ErrInvalidDuration
	}
	if duration <= 0 {
		return 0, "", 0, ErrInvalidDuration
	}

	return steps, activity, duration, nil
}

// parseDuration парсит длительность без regexp. Поддерживает форматы:
// "1h", "30m", "1h30m". Строго проверяет порядок и отсутствие мусора.
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, ErrInvalidDuration
	}

	var hours, minutes float64
	n := len(s)

	hIndex := strings.IndexByte(s, 'h')

	if hIndex != -1 {
		// Есть часы. Перед 'h' должно быть число.
		hPart := s[:hIndex]
		if hPart == "" {
			return 0, ErrInvalidDuration
		}
		val, err := strconv.ParseFloat(hPart, 64)
		if err != nil || val < 0 {
			return 0, ErrInvalidDuration
		}
		hours = val

		restAfterH := s[hIndex+1:]
		if restAfterH == "" {
			// Формат "1h" - валиден
		} else {
			// После 'h' должны идти минуты (формат "30m")
			mIndexInRest := strings.LastIndexByte(restAfterH, 'm')
			if mIndexInRest == -1 {
				return 0, ErrInvalidDuration // Есть символы после h, но нет m
			}

			// После 'm' ничего быть не должно
			if mIndexInRest != len(restAfterH)-1 {
				return 0, ErrInvalidDuration
			}

			mPart := restAfterH[:mIndexInRest]
			if mPart == "" {
				return 0, ErrInvalidDuration // Формат "1hm"
			}

			val, err = strconv.ParseFloat(mPart, 64)
			if err != nil || val < 0 {
				return 0, ErrInvalidDuration
			}
			minutes = val
		}
	} else {
		// Часов нет. Должны быть только минуты.
		mIndex := strings.LastIndexByte(s, 'm')
		if mIndex == -1 {
			return 0, ErrInvalidDuration // Нет ни h, ни m
		}

		// После 'm' ничего быть не должно
		if mIndex != n-1 {
			return 0, ErrInvalidDuration
		}

		mPart := s[:mIndex]
		if mPart == "" {
			return 0, ErrInvalidDuration // Просто "m"
		}

		val, err := strconv.ParseFloat(mPart, 64)
		if err != nil || val < 0 {
			return 0, ErrInvalidDuration
		}
		minutes = val
	}

	totalSeconds := hours*3600 + minutes*60
	if totalSeconds <= 0 {
		return 0, ErrInvalidDuration
	}

	return time.Duration(totalSeconds) * time.Second, nil
}

func distance(steps int, height float64) float64 {
	if steps <= 0 {
		return 0
	}

	var stepLen float64
	if height > 0 {
		stepLen = height * stepLengthCoefficient
	} else {
		// Fallback, если рост невалиден или не передан
		stepLen = lenStep
	}

	return float64(steps) * stepLen / float64(mInKm)
}

func meanSpeed(steps int, height float64, duration time.Duration) float64 {
	if steps <= 0 || duration <= 0 {
		return 0
	}

	distKm := distance(steps, height)
	hours := duration.Hours()
	if hours == 0 {
		return 0
	}

	return distKm / hours
}

func WalkingSpentCalories(steps int, weight float64, height float64, duration time.Duration) (float64, error) {
	if steps <= 0 || weight <= 0 || duration <= 0 {
		return 0, ErrInvalidSteps
	}
	// Рост обязателен для этой формулы согласно логике
	if height <= 0 {
		return 0, ErrInvalidHeight
	}

	speed := meanSpeed(steps, height, duration)
	calories := weight * speed * duration.Hours() * walkingCaloriesCoefficient
	return calories, nil
}

func RunningSpentCalories(steps int, weight float64, height float64, duration time.Duration) (float64, error) {
	if steps <= 0 || weight <= 0 || duration <= 0 {
		return 0, ErrInvalidSteps
	}
	if height <= 0 {
		return 0, ErrInvalidHeight
	}

	speed := meanSpeed(steps, height, duration)
	calories := weight * speed * duration.Hours() * runningCaloriesCoefficient
	return calories, nil
}

func TrainingInfo(data string, weight, height float64) (string, error) {
	steps, activityType, duration, err := parseTraining(data)
	if err != nil {
		return "", err
	}

	// Для TrainingInfo рост 0 допустим: будет использован шаг по умолчанию (lenStep)
	dist := distance(steps, height)
	speed := meanSpeed(steps, height, duration)

	var calories float64
	switch activityType {
	case "Ходьба":
		calories, err = WalkingSpentCalories(steps, weight, height, duration)
	case "Бег":
		calories, err = RunningSpentCalories(steps, weight, height, duration)
	default:
		return "", ErrUnknownActivity
	}

	if err != nil {
		return "", err
	}

	hours := duration.Hours()

	result := fmt.Sprintf(
		"Тип тренировки: %s\nДлительность: %.2f ч.\nДистанция: %.2f км.\nСкорость: %.2f км/ч\nСожгли калорий: %.2f\n",
		activityType, hours, dist, speed, calories,
	)
	return result, nil
}
