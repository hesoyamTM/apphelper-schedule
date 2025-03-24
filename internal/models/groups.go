package models

type Group struct {
	Id        int64
	Name      string
	TrainerId int64
	Students  []int64
	Link      string
}
