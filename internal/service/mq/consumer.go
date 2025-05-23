package mq

import (
	"github.com/namnv2496/crawler/internal/configs"
	"github.com/segmentio/kafka-go"
)

type IConsumer interface {
	GetConsumer() []*kafka.Reader
}
type Consumer struct {
	client []*kafka.Reader
}

func NewKafkaConsumer(
	conf *configs.Config,
) *Consumer {
	consumers := make([]*kafka.Reader, 0)
	for _, topic := range conf.KafkaConsumerConfig.Topic {
		consumers = append(consumers, kafka.NewReader(kafka.ReaderConfig{
			Brokers:   conf.KafkaConsumerConfig.Brokers,
			Topic:     topic,
			GroupID:   conf.KafkaConsumerConfig.GroupID,
			Partition: 0,
			MaxBytes:  10e6, // 10MB
		}))
	}
	return &Consumer{
		client: consumers,
	}
}

func (c *Consumer) GetConsumer() []*kafka.Reader {
	return c.client
}
