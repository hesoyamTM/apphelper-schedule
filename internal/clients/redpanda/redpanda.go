package redpanda

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/IBM/sarama"
	"github.com/hesoyamTM/apphelper-notification/pkg/redpanda"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"go.uber.org/zap"
)

const (
	scheduleCreatedTopic = "schedule.schedule.created"
	scheduleUpdatedTopic = "schedule.schedule.updated"
	groupAddedTopic      = "group.group.added"
)

type RedPanda struct {
	producer    sarama.AsyncProducer
	messageChan chan *sarama.ProducerMessage
	stopChan    chan struct{}
}

func NewRedPanda(ctx context.Context, cfg redpanda.RedpandaConfig) (*RedPanda, error) {
	const op = "redpanda.NewRedPanda"

	redpandaCfg := redpanda.NewSaramaConfig(cfg)
	producer, err := redpanda.NewSaramaAsyncProducer(redpandaCfg, cfg.Brokers)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &RedPanda{
		producer: *producer,
		stopChan: make(chan struct{}),
	}, nil
}

func (r *RedPanda) Start(ctx context.Context) error {
	const op = "redpanda.RedPanda.Start"
	log := logger.GetLoggerFromCtx(ctx)

	var enqueued, successes, failures int32

	go func() {
		for range r.producer.Successes() {
			atomic.AddInt32(&successes, 1)
		}
	}()

	go func() {
		for err := range r.producer.Errors() {
			atomic.AddInt32(&failures, 1)
			log.Error(ctx, "failed to send message to redpanda", zap.Error(err))
		}
	}()

	r.messageChan = make(chan *sarama.ProducerMessage)

	for {
		select {
		case message := <-r.messageChan:
			select {
			case r.producer.Input() <- message:
				atomic.AddInt32(&enqueued, 1)
			case <-r.stopChan:
				return nil
			}
		case <-r.stopChan:
			return nil
		}
	}
}

func (r *RedPanda) Stop(ctx context.Context) error {
	const op = "redpanda.RedPanda.Stop"

	close(r.stopChan)

	if err := r.producer.Close(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log := logger.GetLoggerFromCtx(ctx)
	log.Info(ctx, "stopped redpanda producer")

	return nil
}
