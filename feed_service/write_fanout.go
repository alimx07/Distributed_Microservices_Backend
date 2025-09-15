package main

import (
	"context"
	"encoding/json"
	"log"

	cachedrepo "github.com/alimx07/Distributed_Microservices_Backend/feed_service/cachedRepo"
	"github.com/alimx07/Distributed_Microservices_Backend/feed_service/models"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// The consumer will be the responsable
// about fanout on write, as it will read from topics
// and insert into userFeed Cache
type FanoutWriter struct {
	c     *kafka.Consumer
	ctx   context.Context
	cache cachedrepo.Cache
}

func NewConsumer(ctx context.Context, config models.KafkaConfig, cache cachedrepo.Cache, topics []string) (*FanoutWriter, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootStrapServers,
		"group.id":          config.GroupID,

		// for better batching
		"fetch.min.bytes":   config.FetchMinBytes,
		"auto.offset.reset": config.OffsetReset,

		// for message delivery grantunee , i will go with at-least-ocne
		// We can get at-least-once even with auto commit if we consume all messages correctly before poll or close
		// KAFKA DOCS:
		// Note: Using automatic offset commits can also give you "at-least-once" delivery,
		// but the requirement is that you must consume all data returned from each call to poll(Duration)
		// before any subsequent calls, or before closing the consumer. If you fail to do either of these,
		// it is possible for the committed offset to get ahead of the consumed position, which results in missing records.
		// The advantage of using manual offset control is that you have direct control over when a record is considered "consumed."
		"auto.commit.enable": "enable",
	})
	if err != nil {
		log.Println("Error in intiallizing a kakfa consumer: ", err.Error())
		return nil, err
	}
	err = c.SubscribeTopics(topics, nil)
	if err != nil {
		log.Println("Error in subcribtion to topic: ", err.Error())
		return nil, err
	}

	return &FanoutWriter{
		c:     c,
		ctx:   ctx,
		cache: cache,
	}, nil

}

func (fw *FanoutWriter) WriteFanout(ctx context.Context) {

	run := true
	for run {
		select {
		case <-fw.ctx.Done():
			return
		default:
			ev := fw.c.Poll(100)
			switch e := ev.(type) {
			case *kafka.Message:
				err := fw.ProcessMessage(e)
				log.Println("Error Processing Message", err.Error())
			case kafka.Error:
				log.Println("Error in Consuming events: ", e)
			}
		}
	}
}

func (fw *FanoutWriter) ProcessMessage(msg *kafka.Message) error {
	var item models.FeedItem
	err := json.Unmarshal(msg.Value, &item)
	if err != nil {
		return err
	}
	fw.cache.Set(item)
	return nil
}

func (fw *FanoutWriter) GetPosts(c models.Cursor) ([]models.FeedItem, error) {
	return fw.cache.Get(c)
}
