package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/feed_service/models"
	cachedrepo "github.com/alimx07/Distributed_Microservices_Backend/services/feed_service/cachedRepo"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// FollowClientInterface defines the interface for follow client operations
type FollowClientInterface interface {
	GetFollowers(ctx context.Context, id string) ([]string, error)
	IsCeleb(ctx context.Context, id string) (bool, error)
}

// The consumer will be the responsable
// about fanout on write, as it will read from topics
// and insert into userFeed Cache
type FanoutWriter struct {
	c               *kafka.Consumer
	ctx             context.Context
	cache           cachedrepo.Cache
	followClient    FollowClientInterface
	wg              *sync.WaitGroup
	workerThreshold int
}

func NewFanoutWriter(ctx context.Context, config models.KafkaConfig, cache cachedrepo.Cache, fc FollowClientInterface, workerThreshold int) (*FanoutWriter, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootStrapServers,
		"group.id":          config.GroupID,

		// for better batching
		"fetch.min.bytes":   config.FetchMinBytes,
		"auto.offset.reset": config.OffsetReset,

		// for message delivery grantunee , i will go with at-least-ocne
		// We can get at-least-once even with auto commit if we consume all messages correctly before poll or close
		// Kafka Docs : https://kafka.apache.org/25/javadoc/org/apache/kafka/clients/consumer/KafkaConsumer.html?#:~:text=Note%3A%20Using%20automatic,is%20considered%20%22consumed.%22
		// but I will use manual for more control
		"auto.commit.enable": "false",
	})
	if err != nil {
		log.Println("Error in intiallizing a kakfa consumer: ", err.Error())
		return nil, err
	}
	err = c.SubscribeTopics(config.Topics, nil)
	if err != nil {
		log.Println("Error in subcribtion to topic: ", err.Error())
		return nil, err
	}

	return &FanoutWriter{
		c:               c,
		ctx:             ctx,
		cache:           cache,
		followClient:    fc,
		workerThreshold: workerThreshold,
		wg:              &sync.WaitGroup{},
	}, nil

}

func (fw *FanoutWriter) WriteFanout() {

	for {
		select {
		case <-fw.ctx.Done():
			return
		default:
			ev := fw.c.Poll(100)
			switch e := ev.(type) {
			case *kafka.Message:
				fw.wg.Add(1)
				go func() {
					defer fw.wg.Done()
					err := fw.ProcessMessage(e)
					if err != nil {
						log.Println("Error Processing Message", err.Error())
					} else {
						fw.c.Commit()
					}
				}()
			case kafka.Error:
				log.Println("Error in Consuming events: ", e)
			}
		}
	}
}

// TODO:
// Add more robust error control for message processing
func (fw *FanoutWriter) ProcessMessage(msg *kafka.Message) error {
	var item models.FeedItem
	err := json.Unmarshal(msg.Value, &item)
	if err != nil {
		return err
	}
	// log.Println("POSTID --> ", item.PostId)
	ctx, c1 := context.WithTimeout(fw.ctx, 5*time.Second)
	defer c1()
	celeb, _ := fw.followClient.IsCeleb(ctx, item.UserId)
	if celeb {
		fw.cache.Set(item)
		return nil
	}
	ctx2, c2 := context.WithTimeout(fw.ctx, 5*time.Second)
	defer c2()
	followers, err := fw.followClient.GetFollowers(ctx2, item.UserId)
	if err != nil {
		return err
	}
	i := int32(0)
	var wg sync.WaitGroup
	for int(atomic.LoadInt32(&i)) < len(followers) {
		wg.Add(1)
		endIdx := min(int(atomic.LoadInt32(&i))+fw.workerThreshold, len(followers))
		go func(ids []string) {
			defer wg.Done()
			for _, id := range ids {
				fw.cache.Set(models.FeedItem{
					UserId:     id,
					PostId:     item.PostId,
					Created_at: item.Created_at,
				})
			}
		}(followers[i:endIdx])
		atomic.AddInt32(&i, int32(fw.workerThreshold))
	}
	wg.Wait()
	return nil
}

func (fw *FanoutWriter) close() error {
	// wait until all goroutines end
	fw.wg.Wait()
	return fw.c.Close()
}
