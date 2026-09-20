package spentcalories

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
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
		return 0, "", 0, errors.New("invalid data format")
	}

	parts := strings.Split(data, ",")
	if len(parts) != 3 {
		return 0, "", 0, errors.New("invalid data format")
	}

	stepsStr := strings.TrimSpace(parts[0])
	activity := strings.TrimSpace(parts[1])
	durationStr := strings.TrimSpace(parts[2])

	s, err := strconv.ParseInt(stepsStr, 10, 64)
	if err != nil {
		return 0, "", 0, errors.New("invalid steps value")
	}
	if s <= 0 {
		return 0, "", 0, errors.New("invalid steps value")
	}
	steps := int(s)

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, "", 0, errors.New("invalid duration format")
	}
	if duration <= 0 {
		return 0, "", 0, errors.New("invalid duration value")
	}

	return steps, activity, duration, nil
}

func distance(steps int, height float64) float64 {
	if steps <= 0 {
		return 0
	}

	var stepLen float64
	if height > 0 {
		stepLen = height * stepLengthCoefficient
	} else {
		stepLen = lenStep
	}

	return float64(steps) * stepLen / mInKm
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
	if steps <= 0 {
		return 0, errors.New("invalid steps value")
	}
	if weight <= 0 {
		return 0, errors.New("invalid weight value")
	}
	if duration <= 0 {
		return 0, errors.New("invalid duration value")
	}
	if height <= 0 {
		return 0, errors.New("invalid height value")
	}

	speed := meanSpeed(steps, height, duration)
	calories := weight * speed * duration.Hours() * walkingCaloriesCoefficient
	return calories, nil
}

func RunningSpentCalories(steps int, weight float64, height float64, duration time.Duration) (float64, error) {
	if steps <= 0 {
		return 0, errors.New("invalid steps value")
	}
	if weight <= 0 {
		return 0, errors.New("invalid weight value")
	}
	if duration <= 0 {
		return 0, errors.New("invalid duration value")
	}
	if height <= 0 {
		return 0, errors.New("invalid height value")
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

	dist := distance(steps, height)
	speed := meanSpeed(steps, height, duration)

	var calories float64
	switch activityType {
	case "Ходьба":
		calories, err = WalkingSpentCalories(steps, weight, height, duration)
	case "Бег":
		calories, err = RunningSpentCalories(steps, weight, height, duration)
	default:
		return "", errors.New("неизвестный тип тренировки")
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
