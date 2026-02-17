package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/services/feed_service/models"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Mock Cache implementation for testing
type MockCache struct{}

func (m *MockCache) Set(item models.FeedItem) error {
	// Simulate some work
	time.Sleep(100 * time.Microsecond)
	return nil
}

func (m *MockCache) Get(cursor models.Cursor) ([]models.FeedItem, string, error) {
	return nil, "", nil
}

func (m *MockCache) Close() error {
	return nil
}

// Mock FollowClient implementation for testing
type MockFollowClient struct {
	followers []string
	isCeleb   bool
}

func (m *MockFollowClient) GetFollowers(ctx context.Context, id string) ([]string, error) {
	return m.followers, nil
}

func (m *MockFollowClient) IsCeleb(ctx context.Context, id string) (bool, error) {
	return m.isCeleb, nil
}

func (m *MockFollowClient) GetCeleb(ctx context.Context, id string) ([]string, error) {
	return m.followers, nil
}

func (m *MockFollowClient) Close() error {
	return nil
}

func createTestMessage(userId string, postId string) *kafka.Message {
	item := models.FeedItem{
		UserId:     userId,
		PostId:     postId,
		Created_at: time.Now().Unix(),
	}
	data, _ := json.Marshal(item)
	return &kafka.Message{
		Value: data,
	}
}

func generateFollowers(count int) []string {
	followers := make([]string, count)
	for i := 0; i < count; i++ {
		followers[i] = "user_" + string(rune('0'+i%10))
	}
	return followers
}

// Benchmark with single worker (workerThreshold = total followers count)
func Benchmark_SingleWorker_5000Followers(b *testing.B) {
	ctx := context.Background()
	cache := &MockCache{}
	followers := generateFollowers(5000)
	followClient := &MockFollowClient{
		followers: followers,
		isCeleb:   false,
	}

	fw := &FanoutWriter{
		ctx:             ctx,
		cache:           cache,
		followClient:    followClient,
		workerThreshold: len(followers), // All followers in one worker
	}

	msg := createTestMessage("author123", "post456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := fw.ProcessMessage(msg)
		if err != nil {
			b.Fatalf("ProcessMessage failed: %v", err)
		}
	}
}

func Benchmark_100PerWorker_5000Followers(b *testing.B) {
	ctx := context.Background()
	cache := &MockCache{}
	followers := generateFollowers(5000)
	followClient := &MockFollowClient{
		followers: followers,
		isCeleb:   false,
	}

	fw := &FanoutWriter{
		ctx:             ctx,
		cache:           cache,
		followClient:    followClient,
		workerThreshold: 100, // 100 followers per worker
	}

	msg := createTestMessage("author123", "post456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := fw.ProcessMessage(msg)
		if err != nil {
			b.Fatalf("ProcessMessage failed: %v", err)
		}
	}
}

// Parallel benchmark to test concurrency
func Benchmark_SingleWorker_5000Followers_Parallel(b *testing.B) {
	ctx := context.Background()
	cache := &MockCache{}
	followers := generateFollowers(5000)
	followClient := &MockFollowClient{
		followers: followers,
		isCeleb:   false,
	}

	fw := &FanoutWriter{
		ctx:             ctx,
		cache:           cache,
		followClient:    followClient,
		workerThreshold: len(followers),
	}
	msg := make([]*kafka.Message, 50)
	for i := range 50 {
		msg[i] = createTestMessage(fmt.Sprintf("author_%d", i), fmt.Sprintf("post_%d", i))
	}
	idx := int32(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := fw.ProcessMessage(msg[atomic.LoadInt32(&idx)%50])
			if err != nil {
				b.Fatalf("ProcessMessage failed: %v", err)
			}
			atomic.AddInt32(&idx, 1)
		}
	})
}
func Benchmark_100PerWorker_5000Followers_Parallel(b *testing.B) {
	ctx := context.Background()
	cache := &MockCache{}
	followers := generateFollowers(5000)
	followClient := &MockFollowClient{
		followers: followers,
		isCeleb:   false,
	}

	fw := &FanoutWriter{
		ctx:             ctx,
		cache:           cache,
		followClient:    followClient,
		workerThreshold: 100,
	}

	msg := make([]*kafka.Message, 50)
	for i := range 50 {
		msg[i] = createTestMessage(fmt.Sprintf("author_%d", i), fmt.Sprintf("post_%d", i))
	}
	idx := int32(0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := fw.ProcessMessage(msg[atomic.LoadInt32(&idx)%50])
			if err != nil {
				b.Fatalf("ProcessMessage failed: %v", err)
			}
			atomic.AddInt32(&idx, 1)
		}
	})
}
