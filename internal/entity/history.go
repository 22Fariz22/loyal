package entity

import "time"

type History struct {
	ID          string
	UserID      string
	Number      string
	Status      string
	Sum         int
	ProcessedAt time.Time
}
