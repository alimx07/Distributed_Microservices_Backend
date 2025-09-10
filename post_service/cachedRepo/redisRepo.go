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
	repo        postRepo.PostRepo // presistance db
	redisClient *redis.Client
	ctx         context.Context
}

func NewRedisRepo(repo postRepo.PostRepo, addr, pass string) *redisRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
	})
	return &redisRepo{
		repo:        repo,
		redisClient: client,
	}
}
func (rs *redisRepo) CreatePost(ctx context.Context, post models.Post) error {
	id, err := rs.repo.CreatePost(ctx, post)
	if err != nil {
		return err
	}
	post.Id = id
	data, _ := json.Marshal(post)
	rs.redisClient.Set(ctx, fmt.Sprintf("post:%v", id), data, 0)
	return nil
}

func (rs *redisRepo) CreateComment(ctx context.Context, comment models.Comment) error {
	err := rs.repo.CreateComment(ctx, comment)
	if err != nil {
		return err
	}

	// no caching for comments now
	return nil
}

func (rs *redisRepo) CreateLike(ctx context.Context, like models.Like) error {
	err := rs.repo.CreateLike(ctx, like)
	if err != nil {
		return err
	}

	// no caching for likes now
	return nil
}

func (rs *redisRepo) DeletePost(ctx context.Context, id int64, userID int32) error {
	err := rs.repo.DeletePost(ctx, id, userID)
	if err != nil {
		return err
	}

	// Remove from cache
	rs.redisClient.Del(ctx, fmt.Sprintf("post:%v", id))
	return nil
}

func (rs *redisRepo) DeleteComment(ctx context.Context, id int64, userID int32) error {

	err := rs.repo.DeleteComment(ctx, id, userID)
	if err != nil {
		return err
	}

	return nil
}

func (rs *redisRepo) DeleteLike(ctx context.Context, postID int64, userID int32) error {
	err := rs.repo.DeleteLike(ctx, postID, userID)
	if err != nil {
		return err
	}

	return nil
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
		dbcounters, err := rs.repo.GetCounters(ctx, missedCounter)
		if err != nil {
			log.Println("Error in fetching counters form db: ", err.Error())
		}
		pipe := rs.redisClient.Pipeline()
		for _, cnt := range dbcounters {
			i := idx[cnt.Id]
			posts[i].Likes_count = cnt.Likes_count
			posts[i].Comments_count = cnt.Comments_count
			pipe.HSet(ctx, fmt.Sprintf("post:%v:counters", cnt.Id), map[string]interface{}{
				"likes":    cnt.Likes_count,
				"comments": cnt.Comments_count,
			}, 5*time.Minute)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Println("Error while putting data in counters cache: ", err.Error())
		}
	}
	// From Database
	if len(missedPostIDs) > 0 {
		dbPosts, err := rs.repo.GetPosts(ctx, missedPostIDs)
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

func (rs *redisRepo) GetComments(ctx context.Context, id int64) ([]models.Comment, error) {

	// No Caching for comments Now
	// just read from DB
	return rs.repo.GetComments(ctx, id)

}

func (rs *redisRepo) GetLikes(ctx context.Context, id int64) ([]models.Like, error) {
	// No Caching for likes Now
	// just read from DB
	return rs.repo.GetLikes(ctx, id)
}

func (rs *redisRepo) GetCouters(ctx context.Context, id int64) (map[string]int64, error) {
	cnts, err := rs.redisClient.HGetAll(ctx, fmt.Sprintf("post:%v:counters", id)).Result()
	if err != nil {
		return nil, err
	}
	res := make(map[string]int64)
	for key, value := range cnts {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			log.Printf("Error in parsing %v counter of post:%v: %v\n", key, id, err.Error())
			res[key] = 0
		}
		res[key] = intValue
	}
	return res, nil
}

func (rs *redisRepo) syncCounters() {
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
				rs.repo.UpdateCounters(rs.ctx, cnts)
			}
		}
	}
}
