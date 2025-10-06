package main

import (
	"context"
	"encoding/base64"
	"log"
	"slices"
	"strconv"
	"time"

	cachedrepo "github.com/alimx07/Distributed_Microservices_Backend/feed_service/cachedRepo"
	"github.com/alimx07/Distributed_Microservices_Backend/feed_service/models"
	pb "github.com/alimx07/Distributed_Microservices_Backend/services_bindings_go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FeedService struct {
	pb.UnimplementedFeedServiceServer
	ctx          context.Context
	config       models.ServerConfig
	fw           *FanoutWriter
	cache        cachedrepo.Cache
	postClient   *PostClient
	followClient *FollowClient
	userClient   *UserClient
}

func NewFeedService(ctx context.Context, config models.ServerConfig, Kconfig models.KafkaConfig, cache cachedrepo.Cache) (*FeedService, error) {
	fs := &FeedService{
		ctx:    ctx,
		config: config,
		cache:  cache,
	}
	pc, err := NewPostClient(config.PostService)
	if err != nil {
		log.Fatal("Failed to intiallize connection with PostService", err.Error())
	}
	fc, err := NewFollowClient(config.FollowService)
	if err != nil {
		log.Fatal("Failed to intiallize connection with FollowService", err.Error())
	}
	uc, err := NewUserClient(config.UserService)
	if err != nil {
		log.Fatal("Failed to intiallize connection with UserService", err.Error())
	}
	fs.postClient = pc
	fs.followClient = fc
	fs.userClient = uc
	fw, err := NewFanoutWriter(ctx, Kconfig, cache, fc, 100)
	if err != nil {
		log.Println("Failed to intiallize FanoutWriter ", err.Error())
	}
	fs.fw = fw
	go fs.fw.WriteFanout(fs.ctx)
	return fs, nil
}

// We will get the feed from hybird fanout and
// Merge them and return top N posts (sorted by timestamp) and cursor for next fetch
func (fs *FeedService) GetFeed(ctx context.Context, req *pb.GetFeedRequest) (*pb.GetFeedResponse, error) {
	c := models.Cursor{
		UserId:   req.GetUserId(),
		Cursor:   req.GetCursor(),
		PageSize: req.GetPageSize(),
	}

	c.Cursor, c.PageSize = ValidateCursorData(c.Cursor, c.PageSize)

	items := make([]models.FeedItem, 0, c.PageSize*2)
	// First Get UserFeed Posts
	items, curr_score, err := fs.cache.Get(c)
	if err != nil {
		log.Println("Error in Getting UserFeed Cache Items:", err.Error())
	}

	timeoutCtx, c1 := context.WithTimeout(ctx, 5*time.Second)
	defer c1()
	celebs, err := fs.followClient.GetCeleb(timeoutCtx, c.UserId)

	// we assign a new cursor while we go in the loop to avoid
	// fetching un needed old data
	// if curr_score after fetching is X
	// We only get posts newer than X in other caches

	var celeb_items []models.FeedItem
	for _, celeb := range celebs {
		celeb_items, curr_score, _ = fs.cache.Get(models.Cursor{
			UserId:   celeb,
			Cursor:   curr_score,
			PageSize: c.PageSize,
		})
		items = append(items, celeb_items...)
	}

	// Now we sort them and and get Most Recent N posts
	slices.SortFunc(items, func(a, b models.FeedItem) int {
		if a.Created_at > b.Created_at {
			return -1
		}
		if b.Created_at > a.Created_at {
			return 1
		}
		return 0
	})

	var nextCursor string
	if len(items) > int(c.PageSize) {
		// The next cursor is the timestamp of the last item on the current page
		lastItemTimestamp := items[c.PageSize-1].Created_at
		nextCursor = encode(strconv.FormatInt(lastItemTimestamp, 10))
		items = items[:c.PageSize]
	} else {
		nextCursor = "" //end
	}

	postsCtx, c2 := context.WithTimeout(ctx, 5*time.Second)
	defer c2()
	posts, user_ids, err := fs.postClient.GetPosts(postsCtx, items)
	if err != nil {
		log.Println("feed generation failed. We can`t get items from DB", err.Error())
		return nil, status.Error(codes.Internal, "Failed to getFeed Due to internal Issues")
	}

	// Now We should Get Users Metadata(username)
	userCtx, c3 := context.WithTimeout(context.Background(), 5*time.Second)
	defer c3()
	users, err := fs.userClient.GetUsersData(userCtx, user_ids)
	if err != nil {
		log.Println("Failed To get Users Metadata. Error in Users Service connection or internals", err.Error())
	} else {
		for i := range posts {
			if v, ok := users[user_ids[i]]; ok {
				posts[i].Username = v
			}
		}
	}

	return &pb.GetFeedResponse{
		Posts:      posts,
		NextCursor: nextCursor,
	}, nil
}

func encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
