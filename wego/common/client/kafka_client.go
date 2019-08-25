package client

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	kazoo "github.com/wvanbergen/kazoo-go"
)

const (
	defaultPartitions string = "all"
	defaultBufferSize int    = 256
	defaultRetryMax   int    = 5
)

type KafkaClient struct {
	Topic         string
	BrokerList    string
	ConsumerGroup string
	Zookeeper     string
	Partitions    string
	BufferSize    int
	Debug         bool
	Logger        func(string)

	zookeeperNodes []string
	running        bool // indicate KafkaClient whether stuil to consum data
}

func (k *KafkaClient) Stop() {
	k.Logger(fmt.Sprintf("[KafkaClient][Stop] Stopping the KafkaClient."))
	k.running = false
}

func (k *KafkaClient) Init() {
	if k.Logger == nil {
		k.Logger = func(msg string) {
			fmt.Printf(msg)
		}
	}
	if k.Partitions == "" {
		k.Partitions = defaultPartitions
	}
	if k.BufferSize <= 0 {
		k.BufferSize = defaultBufferSize
	}
}

func (k *KafkaClient) RunConsumer(consumeFunc func([]byte), consumeErrorsFunc func(error)) error {
	return k.runConsumer(consumeFunc, consumeErrorsFunc, "newest")
}

func (k *KafkaClient) RunConsumerFromResetNewest(consumeFunc func([]byte), consumeErrorsFunc func(error)) error {
	return k.runConsumer(consumeFunc, consumeErrorsFunc, "reset_newest")
}

func (k *KafkaClient) RunConsumerFromNewest(consumeFunc func([]byte), consumeErrorsFunc func(error)) error {
	return k.runConsumer(consumeFunc, consumeErrorsFunc, "newest")
}

func (k *KafkaClient) RunConsumerFromResetOldest(consumeFunc func([]byte), consumeErrorsFunc func(error)) error {
	return k.runConsumer(consumeFunc, consumeErrorsFunc, "reset_oldest")
}

func (k *KafkaClient) RunConsumerFromOldest(consumeFunc func([]byte), consumeErrorsFunc func(error)) error {
	return k.runConsumer(consumeFunc, consumeErrorsFunc, "oldest")
}

func (k *KafkaClient) runConsumer(consumeFunc func([]byte), consumeErrorsFunc func(error), from string) error {
	config := consumergroup.NewConfig()
	switch from {
	case "reset_newest":
		config.Offsets.ResetOffsets = true
		config.Offsets.Initial = sarama.OffsetNewest
	case "newest":
		config.Offsets.Initial = sarama.OffsetNewest
	case "reset_oldest":
		config.Offsets.ResetOffsets = true
		config.Offsets.Initial = sarama.OffsetOldest
	case "oldest":
		config.Offsets.Initial = sarama.OffsetOldest
	default:
		k.Logger("Error! Invalid `from`!")
	}
	config.Offsets.ProcessingTimeout = 1 * time.Second
	k.zookeeperNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(k.Zookeeper)
	kafkaTopics := []string{k.Topic}
	consumer, err := consumergroup.JoinConsumerGroup(k.ConsumerGroup, kafkaTopics, k.zookeeperNodes, config)
	if err != nil {
		return fmt.Errorf("[KafkaClient][RunConsumer][Init] error=[%v]", err.Error())
	}
	k.running = true
	go func() {
		for err := range consumer.Errors() {
			consumeErrorsFunc(err)
			if !k.running {
				break
			}
		}
	}()
	go func() {
		for msg := range consumer.Messages() {
			consumeFunc(msg.Value)
			_ = consumer.CommitUpto(msg)
			if !k.running {
				break
			}
		}
	}()
	return nil
}

func (k *KafkaClient) RunConsumerWithoutLoadBalance(consumeFunc func([]byte), consumeErrorsFunc func(error)) error {
	consumer, err := sarama.NewConsumer(strings.Split(k.BrokerList, ","), nil)
	if err != nil {
		return fmt.Errorf("KafkaClient RunConsumerWithoutLoadBalance InitConsumer error=%v", err.Error())
	}
	partitionList, err := k.getPartitions(consumer)
	if err != nil {
		return fmt.Errorf("KafkaClient RunConsumerWithoutLoadBalance getPartitions error=%v", err.Error())
	}
	k.running = true
	var messages = make(chan *sarama.ConsumerMessage, k.BufferSize)
	for _, partition := range partitionList {
		pc, err := consumer.ConsumePartition(k.Topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("KafkaClient RunConsumerWithoutLoadBalance StartConsumer partition=%v error=%v", partition, err.Error())
		}
		go func(pc sarama.PartitionConsumer) {
			for message := range pc.Messages() {
				messages <- message
				if !k.running {
					break
				}
			}
			for err := range pc.Errors() {
				consumeErrorsFunc(err)
			}
		}(pc)
	}
	go func() {
		for msg := range messages {
			consumeFunc(msg.Value)
			if !k.running {
				break
			}
		}
	}()
	return nil
}

func (k *KafkaClient) getPartitions(consumer sarama.Consumer) ([]int32, error) {
	if k.Partitions == "all" {
		return consumer.Partitions(k.Topic)
	}
	tmp := strings.Split(k.Partitions, ",")
	var pList []int32
	for i := range tmp {
		val, err := strconv.ParseInt(tmp[i], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("KafkaClient GetPartitions error=%v", err.Error())
		}
		pList = append(pList, int32(val))
	}
	return pList, nil
}

func (k *KafkaClient) ProduceSyncMsg2Kafka(msg string) error {
	t0 := time.Now()
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = defaultRetryMax
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(strings.Split(k.BrokerList, ","), config)
	if err != nil {
		return fmt.Errorf("KafkaClient ProduceSyncMsg2Kafka InitProducer error=%v", err.Error())
	}
	defer func() {
		if err := producer.Close(); err != nil {
			k.Logger(fmt.Sprintf("KafkaClient ProduceSyncMsg2Kafka CloseProducer error=%v", err.Error()))
		}
	}()
	message := &sarama.ProducerMessage{
		Topic: k.Topic,
		Value: sarama.StringEncoder(msg),
	}
	partition, offset, sendMsgErr := producer.SendMessage(message)
	if sendMsgErr != nil {
		return fmt.Errorf("KafkaClient ProduceSyncMsg2Kafka SendMsg error=%v", sendMsgErr.Error())
	}
	t1 := time.Now()
	if k.Debug {
		k.Logger(fmt.Sprintf("KafkaClient ProduceSyncMsg2Kafka Message is Stored! topic=%v cost=%.3f partition=%v offset=%v",
			k.Topic, t1.Sub(t0).Seconds()*1000, partition, offset))
	}
	return nil
}
