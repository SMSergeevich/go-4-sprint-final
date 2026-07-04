package spentcalories

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidFormat   = errors.New("неверный формат данных")
	ErrUnknownActivity = errors.New("неизвестный тип тренировки")
	ErrInvalidSteps    = errors.New("недопустимое количество шагов")
	ErrInvalidDuration = errors.New("недопустимая продолжительность")
	ErrInvalidHeight   = errors.New("недопустимый рост")
)

const (
	lenStep                    = 0.65
	mInKm                      = 1000
	stepLengthCoefficient      = 0.45
	walkingCaloriesCoefficient = 0.5
	runningCaloriesCoefficient = 1.0
)

func parseTraining(data string) (int, string, time.Duration, error) {
	if data == "" {
		return 0, "", 0, ErrInvalidFormat
	}

	parts := strings.Split(data, ",")
	if len(parts) != 3 {
		return 0, "", 0, ErrInvalidFormat
	}

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

func parseDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`^((\d+(\.\d+)?)h)?((\d+(\.\d+)?)m)?$`)
	if !re.MatchString(s) {
		return 0, errors.New("недопустимый формат продолжительности")
	}

	var hours, minutes float64

	hMatch := regexp.MustCompile(`(\d+(\.\d+)?)h`).FindStringSubmatch(s)
	if len(hMatch) > 1 {
		val, err := strconv.ParseFloat(hMatch[1], 64)
		if err != nil {
			return 0, err
		}
		hours = val
	}

	mMatch := regexp.MustCompile(`(\d+(\.\d+)?)m`).FindStringSubmatch(s)
	if len(mMatch) > 1 {
		val, err := strconv.ParseFloat(mMatch[1], 64)
		if err != nil {
			return 0, err
		}
		minutes = val
	}

	totalSeconds := hours*3600 + minutes*60
	if totalSeconds <= 0 {
		return 0, errors.New("продолжительность должна быть больше 0")
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
		// Для TrainingInfo (где рост может быть 0) оставляем запасной вариант,
		// но в функциях калорий мы теперь не позволяем height <= 0.
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

// Теперь height проверяется на валидность
func WalkingSpentCalories(steps int, weight float64, height float64, duration time.Duration) (float64, error) {
	if steps <= 0 || weight <= 0 || duration <= 0 {
		return 0, ErrInvalidSteps
	}
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

	// Для TrainingInfo рост 0 допустим: будет использован шаг по умолчанию
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
