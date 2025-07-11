package redpanda

type GroupAddedEvent struct {
	GroupId   string `json:"group_id"`
	GroupName string `json:"group_name"`
	TrainerId string `json:"trainer_id"`
	StudentId string `json:"student_id"`
	Link      string `json:"link"`
}
