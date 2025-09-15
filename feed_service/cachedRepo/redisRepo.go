package cachedrepo

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/alimx07/Distributed_Microservices_Backend/feed_service/models"
	"github.com/redis/go-redis/v9"
)

type redisRepo struct {
	r   *redis.Client
	ctx context.Context
}

func NewRedisRepo(ctx context.Context, config models.RedisConfig) *redisRepo {
	r := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Addr, config.Port),
		Password: config.Password,
	})

	return &redisRepo{
		r:   r,
		ctx: ctx,
	}
}

func (rs *redisRepo) Set(f models.FeedItem) error {
	member := redis.Z{
		Score:  float64(f.Created_at),
		Member: f.PostId,
	}
	user_id := strconv.FormatInt(f.UserId, 10)
	err := rs.r.ZAdd(rs.ctx, user_id, member).Err()
	if err != nil {
		log.Printf("Error in Inserting {%v} for user{%v}\n", f.PostId, f.UserId)
		return err
	}
	return nil
}

func (rs *redisRepo) Get(c models.Cursor) ([]models.FeedItem, error) {
	cursor, pageSize, _ := DecodeCursor(c.Cursor)
	opt := &redis.ZRangeBy{
		Max:    cursor,
		Min:    "-inf",
		Offset: 0,
		Count:  pageSize,
	}
	user_id := strconv.FormatInt(c.UserId, 10)
	res, err := rs.r.ZRangeByScoreWithScores(rs.ctx, user_id, opt).Result()
	if err != nil {
		log.Println("Error fetching posts from cache: ", err.Error())
		return nil, err
	}
	items := make([]models.FeedItem, 0, pageSize)
	for _, val := range res {
		var postId int64
		if v, ok := val.Member.(int64); ok {
			postId = v
		}
		item := models.FeedItem{
			UserId: c.UserId,
			PostId: postId,
		}
		items = append(items, item)
	}
	return items, nil
}
