package worker

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type KafkaConsumer struct {
	group   sarama.ConsumerGroup
	topic   string
	handler sarama.ConsumerGroupHandler
	log     *zap.SugaredLogger
}

func NewKafkaConsumer(
	group sarama.ConsumerGroup,
	topic string,
	handler sarama.ConsumerGroupHandler,
	log *zap.SugaredLogger,
) *KafkaConsumer {
	return &KafkaConsumer{
		group:   group,
		topic:   topic,
		handler: handler,
		log:     log,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	go func() {
		defer c.group.Close()

		for {
			if err := c.group.Consume(ctx, []string{c.topic}, c.handler); err != nil {
				c.log.Error("Kafka consume error", zap.Error(err))

				select {
				case <-time.After(time.Second):
				case <-ctx.Done():
					return
				}
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()
}
