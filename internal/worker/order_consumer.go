package worker

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

type OrderConsumer struct {
	log       Logger
	processor OrderProcessor
}

func NewOrderConsumer(p OrderProcessor, log Logger) *OrderConsumer {
	return &OrderConsumer{
		processor: p,
		log:       log,
	}
}

func (c *OrderConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *OrderConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *OrderConsumer) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for msg := range claim.Messages() {
		var ev OrderEvent
		if err := json.Unmarshal(msg.Value, &ev); err != nil {
			c.log.Warnw("bad kafka message", "err", err)
			session.MarkMessage(msg, "")
			continue
		}

		if err := c.processor.Process(session.Context(), ev); err != nil {
			c.log.Errorw("order event failed", "order", ev.OrderID, "err", err)
			continue
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
