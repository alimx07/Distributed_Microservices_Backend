package main

import (
	"context"
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	cachedrepo "github.com/alimx07/Distributed_Microservices_Backend/feed_service/cachedRepo"
	"github.com/alimx07/Distributed_Microservices_Backend/feed_service/models"
	pb "github.com/alimx07/Distributed_Microservices_Backend/services_bindings_go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FeedService struct {
	pb.UnimplementedFeedServiceServer
	ctx          context.Context
	cancel       context.CancelFunc
	config       models.ServerConfig
	fw           *FanoutWriter
	cache        cachedrepo.Cache
	postClient   *PostClient
	followClient *FollowClient
	userClient   *UserClient
	grpcServer   *grpc.Server
	httpServer   *http.Server
	wg           *sync.WaitGroup
	serviceOFF   atomic.Bool
}

func NewFeedService(config models.ServerConfig, Kconfig models.KafkaConfig, cacheConfig models.RedisConfig) (*FeedService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	fs := &FeedService{
		config: config,
		ctx:    ctx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
	}
	pc, err := NewPostClient(config.PostService)
	if err != nil {
		fs.closeClients()
		log.Fatal("Failed to intiallize connection with PostService", err.Error())
	}
	fc, err := NewFollowClient(config.FollowService)
	if err != nil {
		// fs.closeClients()
		log.Println("Failed to intiallize connection with FollowService", err.Error())
	}
	uc, err := NewUserClient(config.UserService)
	if err != nil {
		fs.closeClients()
		log.Fatal("Failed to intiallize connection with UserService", err.Error())
	}
	fs.postClient = pc
	fs.followClient = fc
	fs.userClient = uc

	cache, err := cachedrepo.NewRedisRepo(ctx, cacheConfig)
	if err != nil {
		fs.closeClients()
		fs.closeResources()
		log.Fatal("Error in Connecting to cache: ", err.Error())
	}
	fs.cache = cache
	fw, err := NewFanoutWriter(ctx, Kconfig, cache, fc, 100)
	if err != nil {
		fs.closeClients()
		fs.closeResources()
		log.Println("Failed to intiallize FanoutWriter ", err.Error())
	}
	fs.fw = fw

	fs.wg.Add(1)
	go func() {
		fs.fw.WriteFanout()
		fs.wg.Done()
	}()
	fs.serviceOFF.Store(false)
	return fs, nil
}

func (fs *FeedService) Start() error {
	log.Printf("Start Listening to Feed Service on %v\n", net.JoinHostPort(fs.config.ServerHost, fs.config.ServerPort))
	conn, err := net.Listen("tcp", net.JoinHostPort(fs.config.ServerHost, fs.config.ServerPort))
	if err != nil {
		log.Printf("Error in Starting listener for FeedService on %v\n", net.JoinHostPort(fs.config.ServerHost, fs.config.ServerPort))
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterFeedServiceServer(grpcServer, fs)
	fs.grpcServer = grpcServer
	return grpcServer.Serve(conn)
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

	// items := make([]models.FeedItem, 0, c.PageSize*2)

	// First Get precomputed UserFeed Posts
	log.Println(fs == nil, fs.cache == nil)
	items, curr_score, err := fs.cache.Get(c)
	if err != nil {
		log.Println("Error in Getting UserFeed Cache Items:", err.Error())
	}

	timeoutCtx, c1 := context.WithTimeout(ctx, 2*time.Second)
	defer c1()

	// Second Get Most Recents posts in celebs caches
	var celebs []string
	if fs.followClient != nil {
		celebs, _ = fs.followClient.GetCeleb(timeoutCtx, c.UserId)
	}

	// we assign a new cursor while we go in the loop to avoid
	// fetching unnecessary old data
	// if curr_score after fetching CacheX is Y
	// We only get posts newer than Y in other caches

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
	userCtx, c3 := context.WithTimeout(ctx, 5*time.Second)
	defer c3()
	users, err := fs.userClient.GetUsersData(userCtx, user_ids)
	if err != nil {
		log.Println("Failed To get Users Metadata. Error in Users Service connection or internals", err.Error())
	} else {
		for i := range posts {
			if v, ok := users[posts[i].UserId]; ok {
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

func (fs *FeedService) StartHealthServer() error {
	router := http.NewServeMux()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if fs.serviceOFF.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status": "Down" , "service": "feed_service"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok" , "service": "feed_service"}`))
		}
	})

	var handler http.Handler = router
	server := &http.Server{
		Addr:    net.JoinHostPort(fs.config.ServerHost, fs.config.ServerHTTPPort),
		Handler: handler,
	}
	fs.httpServer = server
	log.Printf("PostServer HTTP starting on %s:%s\n", fs.config.ServerHost, fs.config.ServerHTTPPort)
	return server.ListenAndServe()
}

func (fs *FeedService) close() {
	log.Println("Starting shutdown feedService.....")

	// mark service as down
	fs.serviceOFF.Store(true)

	// wait until new state reflected in api_gateway
	time.Sleep(5 * time.Second)

	// Now no new traffic will come

	// start ending current requests
	fs.grpcServer.GracefulStop()
	log.Println("Grpc Server Closed...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := fs.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error in closing HttpServer: %v", err)
	} else {
		log.Println("Http Server Closed...")
	}

	fs.cancel()

	// wait to any pending requests
	done := make(chan struct{})
	go func() {
		fs.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All Pending Requests end successfully")
	case <-time.After(30 * time.Second):
		log.Println("Timeout waiting pending requests")
	}

	// close clients
	fs.closeClients()

	// close any resources
	fs.closeResources()

	log.Println("ShutDown Completed...")

}

func (fs *FeedService) closeClients() {
	if fs.postClient != nil {
		if err := fs.postClient.Close(); err != nil {
			log.Printf("Error Closing post client: %v", err)
		}
	}
	if fs.followClient != nil {
		if err := fs.followClient.Close(); err != nil {
			log.Printf("Error Closing follw client: %v", err)
		}
	}
	if fs.userClient != nil {
		if err := fs.userClient.Close(); err != nil {
			log.Printf("Error Closing user client: %v", err)
		}
	}
}

func (fs *FeedService) closeResources() {
	if fs.fw != nil {
		if err := fs.fw.close(); err != nil {
			log.Printf("Error Closing Kafka Clinet: %v", err)
		}
	}
	if fs.cache != nil {
		if err := fs.cache.Close(); err != nil {
			log.Printf("Error closing cache Client: %v", err)
		}
	}
}
