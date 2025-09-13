package cachedRepo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

func NewRedisRepo(repo postRepo.PersistenceDB, addr, pass string) *redisRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
	})
	return &redisRepo{
		presistanceDB: repo,
		redisClient:   client,
	}
}

func (rs *redisRepo) CachePost(ctx context.Context, post models.CachedPost) error {
	data, _ := json.Marshal(post)
	err := rs.redisClient.Set(ctx, fmt.Sprintf("post:%v", post.Id), data, 0).Err()
	return err
}

func (rs *redisRepo) DeletePost(ctx context.Context, id int64) error {
	// Remove from cache
	err := rs.redisClient.Del(ctx, fmt.Sprintf("post:%v", id)).Err()
	return err
}

func (rs *redisRepo) UpdateLikesCounter(ctx context.Context, id int64) {
	rs.redisClient.HIncrBy(ctx, fmt.Sprintf("post:%v:counters", id), "likes", 1)
}

func (rs *redisRepo) UpdateCommentsCounter(ctx context.Context, id int64) {
	rs.redisClient.HIncrBy(ctx, fmt.Sprintf("post:%v:counters", id), "comments", 1)
}

// func (rs *redisRepo) GetCounters(ctx context.Context , keys []string) *redis.SliceCmd {

// }

// func (rs *redisRepo) GetCommentsCounter(ctx context.Context , keys []string) *redis.SliceCmd {

// 	for _ , key := range keys {
// 		commentKeys[i] = fmt.Sprintf("%v:comments" , key)
// 	}
// 	return rs.redisClient.MGet(ctx , commentKeys...)
// }

// This function will be moved later
func (rs *redisRepo) GetPosts(ctx context.Context, ids []int64) ([]models.Post, error) {
	if len(ids) == 0 {
		return []models.Post{}, nil
	}

	keys := make([]string, len(ids))
	cmds := make([]*redis.MapStringStringCmd, len(ids))
	pipe := rs.redisClient.Pipeline()

	for i, id := range ids {
		keys[i] = fmt.Sprintf("post:%d", id)
		cmds[i] = pipe.HGetAll(ctx, fmt.Sprintf("posts:%v:counters", id))
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
		cnt := cmds[i].Val()

		// in cache of cache miss
		// instead on go to db for this counter only
		// store it in map , go to db once get all couters
		// and map counters back to thier posts
		if len(cnt) == 0 {
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
			pipe.HSet(ctx, fmt.Sprintf("post:%v:counters", cnt.Id), map[string]interface{}{
				"likes":    cnt.Likes,
				"comments": cnt.Comments,
			}, 5*time.Minute)
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
			pipe.HSet(ctx, fmt.Sprintf("post:%v:counters", post.Id), map[string]interface{}{
				"likes":    post.Likes_count,
				"comments": post.Comments_count,
			}, 5*time.Minute)
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
	for {
		select {
		case <-rs.ctx.Done():
			return
		case <-ticker.C:
			var cursor uint64
			for {
				var keys []string
				var err error
				keys, cursor, err = rs.redisClient.SScan(rs.ctx, "counters:set", cursor, "*", 100).Result()
				if err != nil {
					continue
				}
				if cursor == 0 {
					break
				}
				pipe := rs.redisClient.Pipeline()
				cmds := make(map[string]*redis.MapStringStringCmd)
				for _, key := range keys {
					cmds[key] = pipe.HGetAll(rs.ctx, key)
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
					if v, ok := cnt["likes"]; ok {
						if n, err := strconv.ParseInt(v, 10, 64); err == nil {
							cachedcounter.Likes = n
						}
					}
					if v, ok := cnt["comments"]; ok {
						if n, err := strconv.ParseInt(v, 10, 64); err == nil {
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
