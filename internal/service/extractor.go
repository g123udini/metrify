package service

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrInvalidPathFormat = errors.New("invalid path format")
)

func split(path string) ([]string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 4 {
		return nil, ErrInvalidPathFormat
	}

	return parts, nil
}

func ExtractType(path string) (string, error) {
	parts, err := split(path)

	if err != nil {
		return "", err
	}

	return parts[1], err
}

func ExtractName(path string) (string, error) {
	parts, err := split(path)

	if err != nil {
		return "", err
	}

	return parts[2], err
}

func ExtractCounterValue(path string) (int64, error) {
	parts, err := split(path)

	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(parts[3], 10, 64)
}

func ExtractGaugeValue(path string) (float64, error) {
	parts, err := split(path)

	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(parts[3], 64)
}
