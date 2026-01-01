package cachedrepo

import (
	"context"
	"log"
	"math"
	"strconv"

	"github.com/alimx07/Distributed_Microservices_Backend/feed_service/models"
	"github.com/redis/go-redis/v9"
)

type redisRepo struct {
	r   *redis.ClusterClient
	ctx context.Context
}

func NewRedisRepo(ctx context.Context, config models.RedisConfig) (*redisRepo, error) {
	r := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    config.ClusterAddr,
		Password: config.Password,
	})

	if err := r.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &redisRepo{
		r:   r,
		ctx: ctx,
	}, nil
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
	err := rs.r.ZAdd(rs.ctx, f.UserId, member).Err()
	if err != nil {
		log.Printf("Error in Inserting {%v} for user{%v}--> %v\n", f.PostId, f.UserId, err.Error())
		return err
	}
	return nil
}

func (rs *redisRepo) Get(c models.Cursor) ([]models.FeedItem, string, error) {
	opt := redis.ZRangeArgs{
		Key:     c.UserId,
		Start:   c.Cursor,
		Stop:    "+inf",
		Offset:  0,
		Count:   int64(c.PageSize),
		ByScore: true,
		Rev:     true,
	}
	res, err := rs.r.ZRangeArgsWithScores(rs.ctx, opt).Result()
	if err != nil {
		log.Println("Error fetching posts from cache: ", err.Error())
		return nil, "", err
	}
	items := make([]models.FeedItem, 0, c.PageSize)
	score := strconv.FormatInt(int64(math.MaxInt64), 10)
	if len(res) > 0 {
		score = strconv.FormatInt(int64(res[len(res)-1].Score), 10)
	}
	for _, val := range res {
		post_id, ok := val.Member.(string)
		if !ok {
			log.Printf("Failed to assert member to string: %v", val.Member)
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

func (rs *redisRepo) Close() error {
	return rs.r.Close()
}
