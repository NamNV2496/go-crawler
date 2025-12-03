package mq

import (
	"context"
	"encoding/json"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/pkg/logging"
	"github.com/segmentio/kafka-go"
)

type IProducer interface {
	Publish(ctx context.Context, topic, key string, value any) error
}

type Producer struct {
	client map[string]*kafka.Writer
}

func NewKafkaProducer(
	conf *configs.Config,
) IProducer {
	clients := make(map[string]*kafka.Writer, 0)
	for _, topic := range conf.KafkaProducerConfig.Topic {
		clients[topic] = &kafka.Writer{
			Addr:                   kafka.TCP(conf.KafkaConsumerConfig.Brokers...),
			Balancer:               &kafka.LeastBytes{},
			Topic:                  topic,
			AllowAutoTopicCreation: true,
		}
	}
	return &Producer{
		client: clients,
	}
}

func (p *Producer) Publish(ctx context.Context, topic, key string, value any) error {
	deferFunc := logging.AppendPrefix("Publish")
	defer deferFunc()
	producer := p.client[topic]
	if producer == nil {
		logging.Error(ctx, "topic %s not found", topic)
		return nil
	}
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return producer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(jsonData),
		},
	)
}
