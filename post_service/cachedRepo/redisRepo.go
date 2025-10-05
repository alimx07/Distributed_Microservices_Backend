package cachedRepo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
	"github.com/alimx07/Distributed_Microservices_Backend/post_service/postRepo"
	"github.com/redis/go-redis/v9"
)

type redisRepo struct {
	presistanceDB postRepo.PersistenceDB
	redisClient   *redis.Client
	ctx           context.Context
}

func NewRedisRepo(repo postRepo.PersistenceDB, host, port, pass string) *redisRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(host, port),
		Password: pass,
	})
	return &redisRepo{
		presistanceDB: repo,
		redisClient:   client,
	}
}

func (rs *redisRepo) CachePost(ctx context.Context, post models.CachedPost) error {
	data, _ := json.Marshal(post)
	err := rs.redisClient.Set(ctx, postKey(post.Id), data, 12*time.Hour).Err()
	return err
}

func (rs *redisRepo) DeletePost(ctx context.Context, id int64) error {
	// Remove from cache
	err := rs.redisClient.Del(ctx, postKey(id)).Err()
	return err
}

func (rs *redisRepo) UpdateLikesCounter(ctx context.Context, id int64) {
	pipe := rs.redisClient.Pipeline()
	pipe.HIncrBy(ctx, counterKey(id), "likes", 1)
	pipe.SAdd(ctx, "counters:set", counterKey(id))
	pipe.Exec(ctx)
}

func (rs *redisRepo) UpdateCommentsCounter(ctx context.Context, id int64) {
	pipe := rs.redisClient.Pipeline()
	pipe.HIncrBy(ctx, counterKey(id), "comments", 1)
	pipe.SAdd(ctx, "counters:set", counterKey(id))
	pipe.Exec(ctx)
}

func (rs *redisRepo) GetPosts(ctx context.Context, ids []int64) ([]models.Post, error) {
	if len(ids) == 0 {
		return []models.Post{}, nil
	}

	keys := make([]string, len(ids))
	cmds := make([]*redis.MapStringStringCmd, len(ids))
	pipe := rs.redisClient.Pipeline()

	for i, id := range ids {
		keys[i] = fmt.Sprintf("post:%d", id)
		cmds[i] = pipe.HGetAll(ctx, counterKey(id))
	}
	// From Cache First
	posts := make([]models.Post, len(ids))
	var missedPostIDs []int64
	var missedCounter []int64

	mgetvalues := pipe.MGet(ctx, keys...)
	_, err := pipe.Exec(ctx)

	if err != nil {
		log.Println("Error in Loading data from redis caches:", err.Error())
	}

	postVals := mgetvalues.Val()

	idx := make(map[int64]int)
	for i, val := range postVals {

		// post cache miss
		if val == nil {
			missedPostIDs = append(missedPostIDs, ids[i])
			continue
		}
		var p models.Post
		if err := json.Unmarshal([]byte(val.(string)), &p); err != nil {
			log.Println("Error in UnMarshal post: ", err.Error())
			missedPostIDs = append(missedPostIDs, ids[i])
			continue
		}
		cnt, err := cmds[i].Result()

		// in case of cache miss
		// instead on go to db for this counter only
		// store it in map , go to db once get all couters
		// and map counters back to thier posts
		if err != nil || len(cnt) == 0 {
			log.Printf("Cache miss for counters of post:%v\n", err.Error())
			missedCounter = append(missedCounter, p.Id)
			idx[p.Id] = i
		} else {
			if v, ok := cnt["likes"]; ok {
				if n, err := strconv.ParseInt(v, 10, 64); err == nil {
					p.Likes_count = n
				}
			}
			if v, ok := cnt["comments"]; ok {
				if n, err := strconv.ParseInt(v, 10, 64); err == nil {
					p.Comments_count = n
				}
			}
		}
		posts = append(posts, p)
	}
	// Get missed counters
	if len(missedCounter) > 0 {
		dbcounters, err := rs.presistanceDB.GetCounters(ctx, missedCounter)
		if err != nil {
			log.Println("Error in fetching counters form db: ", err.Error())
		}
		pipe := rs.redisClient.Pipeline()
		for _, cnt := range dbcounters {
			i := idx[cnt.Id]
			posts[i].Likes_count = cnt.Likes
			posts[i].Comments_count = cnt.Comments
			pipe.HMSet(ctx, counterKey(cnt.Id), "likes", cnt.Likes, "comments", cnt.Comments)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Println("Error while putting data in counters cache: ", err.Error())
		}
	}
	// From Database
	if len(missedPostIDs) > 0 {
		dbPosts, err := rs.presistanceDB.GetPosts(ctx, missedPostIDs)
		if err != nil {
			return posts, err
		}

		pipe := rs.redisClient.Pipeline()
		// Finally Cache DBPosts
		for _, post := range dbPosts {
			data, _ := json.Marshal(post.CachedPost)

			// populate data back to caches
			pipe.Set(ctx, fmt.Sprintf("post:%v", post.Id), data, 24*time.Hour)
			pipe.MSet(ctx, fmt.Sprintf("post:%v:likes", post.Id), post.Likes_count, fmt.Sprintf("post:%v:comments", post.Id), post.Comments_count)
			posts = append(posts, post)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Printf("Error while putting new values in cache %v", err.Error())
		}
	}
	return posts, nil
}

func (rs *redisRepo) SyncCounters() {
	ticker := time.NewTicker(2 * time.Minute)

	defer ticker.Stop()
	for {
		select {
		case <-rs.ctx.Done():
			return
		case <-ticker.C:
			var cursor uint64
			for {
				var keys []string
				var err error
				keys, cursor, err = rs.redisClient.SScan(rs.ctx, "counters:set", cursor, "post:*:counters", 100).Result()
				if err != nil {
					continue
				}
				if cursor == 0 {
					break
				}
				pipe := rs.redisClient.Pipeline()
				cmds := make(map[string]*redis.StringSliceCmd)
				for _, key := range keys {

					// Get and Delete all keys from cache
					// now any new likes/comments will be added by updateFuncs and sync in the next cycle
					// but there is consistency issue if flushing to db failed
					// rename trick can be used or queue for consistency but i will go with GetDel solution for simplecity now
					cmds[key] = pipe.HGetDel(rs.ctx, key, "likes", "comments")
				}
				_, err = pipe.Exec(rs.ctx)
				if err != nil {
					log.Println("Error while Fetching Counters in background sync: ", err.Error())
				}
				var cnts []models.CachedCounter
				for key, cmd := range cmds {
					cnt, err := cmd.Result()
					if err != nil {
						log.Println(err.Error())
						continue
					}
					cnt_key, _ := strconv.ParseInt(strings.Split(key, ":")[1], 10, 64)
					cachedcounter := models.CachedCounter{
						Id: cnt_key,
					}

					if cnt[0] != "" {
						if n, err := strconv.ParseInt(cnt[0], 10, 64); err == nil {
							cachedcounter.Likes = n
						}
					}
					if cnt[1] != "" {
						if n, err := strconv.ParseInt(cnt[1], 10, 64); err == nil {
							cachedcounter.Comments = n
						}
					}
					cnts = append(cnts, cachedcounter)
				}
				rs.presistanceDB.UpdateCounters(rs.ctx, cnts)
			}
		}
	}
}

func postKey(id int64) string {
	return fmt.Sprintf("post:%v", id)
}

func counterKey(id int64) string {
	return fmt.Sprintf("post:%v:counters", id)
}

// func (rs *redisRepo) SyncCounters() {
// 	ticker := time.NewTicker(2 * time.Minute)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-rs.ctx.Done():
// 			return
// 		case <-ticker.C:
// 			if err := rs.syncIteration(); err != nil {
// 				log.Printf("Sync iteration failed: %v", err)
// 			}
// 		}
// 	}
// }
