package data

import (
	"time"
)

type Rate struct {
	ID    uint
	Date  time.Time
	Value float64
}
