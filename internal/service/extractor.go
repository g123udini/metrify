package service

import (
	"errors"
	"strings"
)

var (
	ErrInvalidPathFormat = errors.New("invalid path format")
)

func split(path string) ([]string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 2 {
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
