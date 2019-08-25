package client

import (
	"os"
	"testing"
)

func TestKafKaClient(t *testing.T) {
	kafkaClient := KafkaClient{
		Topic:         "KafkaTopicName",
		BrokerList:    os.Getenv("KAFKA_PEERS"),
		ConsumerGroup: "KafkaConsumerGroupName",
		Zookeeper:     "127.0.0.1:8988,127.0.0.1:8986,127.0.0.1:8987",
		Partitions:    "all",
		BufferSize:    256,
	}
	kafkaClient.Init()
	_ = kafkaClient.RunConsumer(func(msgFromKafka []byte) {
		// Consume a Message from Kafka
	}, func(err error) {
		// Get an Error When Consuming Kafka
	})
}
