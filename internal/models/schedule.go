package models

import "time"

type Schedule struct {
	GroupName string
	GroupId   int64
	StudentId int64
	TrainerId int64
	Date      time.Time
}
