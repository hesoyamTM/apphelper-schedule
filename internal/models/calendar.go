package models

import "time"

type Calendar struct {
	Title string
	Id    string
}

type CalendarEvent struct {
	EventId    string
	CalendarId string
	Title      string
	Start      time.Time
	End        time.Time
}
