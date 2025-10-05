package service

import (
	models "metrify/internal/model"
	"slices"
)

var validTypes = []string{models.Counter, models.Gauge}

func ValidateMetricType(name string) bool {
	return slices.Contains(validTypes, name)
}
