package models

import "github.com/google/uuid"

type Group struct {
	Id        uuid.UUID
	Name      string
	TrainerId uuid.UUID
	Students  []uuid.UUID
	Link      string
}
