package entity

import "time"

type Order struct {
	ID         string
	UserID     string
	Number     string
	Accrual    uint32
	Status     string
	UploadedAt time.Time
}
