package redpanda

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/hesoyamTM/apphelper-schedule/internal/clients"
	"github.com/hesoyamTM/apphelper-schedule/internal/models"
)

func (r *RedPanda) ScheduleCreatedEvent(ctx context.Context, schedule *models.Schedule) error {
	const op = "redpanda.RedPanda.ScheduleCreateEvent"

	value, err := json.Marshal(schedule)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	message := &sarama.ProducerMessage{
		Topic: scheduleCreatedTopic,
		Value: sarama.ByteEncoder(value),
	}

	select {
	case r.messageChan <- message:
		return nil
	case <-r.stopChan:
		return fmt.Errorf("%s: %w", op, clients.ErrClosedChannel)
	default:
		return fmt.Errorf("%s: %w", op, clients.ErrClosedChannel)
	}
}

func (r *RedPanda) ScheduleUpdatedEvent(ctx context.Context, schedule *models.Schedule) error {
	const op = "redpanda.RedPanda.ScheduleUpdatedEvent"

	value, err := json.Marshal(schedule)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	message := &sarama.ProducerMessage{
		Topic: scheduleUpdatedTopic,
		Value: sarama.ByteEncoder(value),
	}

	select {
	case r.messageChan <- message:
		return nil
	case <-r.stopChan:
		return fmt.Errorf("%s: %w", op, clients.ErrClosedChannel)
	default:
		return fmt.Errorf("%s: %w", op, clients.ErrClosedChannel)
	}
}

func (r *RedPanda) GroupAddedEvent(ctx context.Context, group *GroupAddedEvent) error {
	const op = "redpanda.RedPanda.GroupAddedEvent"

	value, err := json.Marshal(group)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	message := &sarama.ProducerMessage{
		Topic: groupAddedTopic,
		Value: sarama.ByteEncoder(value),
	}

	select {
	case r.messageChan <- message:
		return nil
	case <-r.stopChan:
		return fmt.Errorf("%s: %w", op, clients.ErrClosedChannel)
	default:
		return fmt.Errorf("%s: %w", op, clients.ErrClosedChannel)
	}
}
