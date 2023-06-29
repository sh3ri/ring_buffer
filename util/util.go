package util

import (
	"circular_buffer/models"

	"golang.org/x/exp/constraints"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func TrySend[T models.ANY](data *T, channel chan *T) bool {
	select {
	case channel <- data:
		return true
	default:
		return false
	}
}
