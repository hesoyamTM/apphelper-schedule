package models

import (
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	Id        uuid.UUID `json:"id"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	GroupName string    `json:"group_name"`
	Title     string    `json:"title"`
	GroupId   uuid.UUID `json:"group_id"`
	StudentId uuid.UUID `json:"student_id"`
	TrainerId uuid.UUID `json:"trainer_id"`
}
