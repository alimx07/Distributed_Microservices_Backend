package cachedrepo

import (
	"context"
	"fmt"
	"log"
	"math"
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

// There is two types of data in caches:
// 1- User precomputed Feed
// 2- Celebs Post IDs

// The Celebs one will be used to query and get postIDs from specific celebs that user follows
// and merge Posts with precomputed ones to have the final N posts for current feed request
// then a call for Post Service once for get all posts

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

func (rs *redisRepo) Get(c models.Cursor) ([]models.FeedItem, string, error) {
	user_id := strconv.FormatInt(c.UserId, 10)
	opt := redis.ZRangeArgs{
		Key:    user_id,
		Start:  c.Cursor,
		Stop:   "+inf",
		Offset: 0,
		Count:  int64(c.PageSize),
	}
	res, err := rs.r.ZRangeArgsWithScores(rs.ctx, opt).Result()
	if err != nil {
		log.Println("Error fetching posts from cache: ", err.Error())
		return nil, "", err
	}
	items := make([]models.FeedItem, 0, c.PageSize)
	score := strconv.FormatInt(int64(res[0].Score), 10)
	if len(res) > 0 {
		score = strconv.FormatInt(int64(math.MaxInt64), 10)
	}
	for _, val := range res {
		post_id_str, ok := val.Member.(string)
		if !ok {
			log.Printf("Failed to assert member to string: %v", val.Member)
			continue
		}
		post_id, err := strconv.ParseInt(post_id_str, 10, 64)
		if err != nil {
			log.Printf("Failed to parse post_id: %v", err)
			continue
		}
		item := models.FeedItem{
			PostId:     post_id,
			Created_at: int64(val.Score),
		}
		items = append(items, item)
	}
	return items, score, nil
}
