package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/cachedRepo"
	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
	"github.com/alimx07/Distributed_Microservices_Backend/post_service/postRepo"
	pb "github.com/alimx07/Distributed_Microservices_Backend/services_bindings_go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type postService struct {
	pb.UnimplementedPostSeriveServer
	presistanceDB postRepo.PersistenceDB
	cache         cachedRepo.CachedRepo
	config        models.Config
}

func NewPostService(presistance postRepo.PersistenceDB, cache cachedRepo.CachedRepo, config models.Config) *postService {
	return &postService{
		presistanceDB: presistance,
		cache:         cache,
		config:        config,
	}
}

func (ps *postService) start() error {
	log.Printf("Starting gRPC server on %s:%s", ps.config.ServerHost, ps.config.ServerPort)
	listener, err := net.Listen("tcp", net.JoinHostPort(ps.config.ServerHost, ps.config.ServerPort))
	if err != nil {
		log.Fatal("Can not intialized Connection on Host:", net.JoinHostPort(ps.config.ServerHost, ps.config.ServerPort))
	}
	grpcserver := grpc.NewServer()
	pb.RegisterPostSeriveServer(grpcserver, ps)
	// reflection.Register(grpcserver)
	return grpcserver.Serve(listener)
}

func (ps *postService) CreatePost(ctx context.Context, req *pb.Post) (*pb.Response, error) {
	log.Println("DATA: ", req.GetContent(), "  ", req.GetUserId())
	post := models.CachedPost{User_id: req.GetUserId(),
		Content: req.GetContent()}

	log.Println(post.User_id, post.Content)
	id, err := ps.presistanceDB.CreatePost(ctx, models.Post{CachedPost: post})
	if err != nil {
		log.Printf("Failed to create post for user{%v}: {%v}\n", post.User_id, err.Error())
		return nil, status.Error(codes.Internal, "Post Can`t be Created Due to internal Issues")
	}
	post.Id = id
	err = ps.cache.CachePost(ctx, post)
	if err != nil {
		log.Printf("Failed to cache post{%v} for user{%v}\n : {%v}", post.Id, post.User_id, err.Error())
	}
	return &pb.Response{
		Message: "Post Created Successfully",
	}, nil
}

func (ps *postService) CreateComment(ctx context.Context, req *pb.Comment) (*pb.Response, error) {
	comment := models.Comment{
		User_id: req.GetUserId(),
		Post_id: req.GetPostId(),
		Content: req.GetComment(),
	}
	id, err := ps.presistanceDB.CreateComment(ctx, comment)
	if err != nil {
		log.Printf("Failed to create comment on post{%v} by user{%v}: %v", comment.Post_id, comment.User_id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to create comment Due to internal Issues")
	}
	ps.cache.UpdateCommentsCounter(ctx, id, 1)
	return &pb.Response{
		Message: "Comment Created Successfully",
	}, nil
}

func (ps *postService) CreateLike(ctx context.Context, req *pb.Like) (*pb.Response, error) {
	like := models.Like{
		User_id: req.UserId,
		Post_id: req.PostId,
	}
	err := ps.presistanceDB.CreateLike(ctx, like)
	if err != nil {
		log.Printf("Failed to create Like on post{%v} by user{%v} : %v", like.Post_id, like.User_id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to create like Due to internal Issues")
	}
	ps.cache.UpdateLikesCounter(ctx, like.Post_id, 1)
	return &pb.Response{
		Message: "Like Created Successfully",
	}, nil
}

func (ps *postService) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.Response, error) {
	id := req.GetPostId()
	err := ps.presistanceDB.DeletePost(ctx, id)
	if err != nil {
		log.Printf("Failed to delete post{%v}: %v\n", id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to Delete Post Due to Internal Issues")
	}

	err = ps.cache.DeletePost(ctx, id)
	if err != nil {
		// in case of cache failing , then there will be inconsistency as users can still see the post until expiration
		// even it is deleted. a background proccess can be spin up to take cache failed operations and
		// retry them again. for simplecity just log the error now
		log.Printf("Failed to Delete post {%v} from the cache: {%v}\n", id, err.Error())
	}

	return &pb.Response{
		Message: "Post Deleted Successfully",
	}, nil
}

func (ps *postService) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.Response, error) {
	id := req.GetCommentId()

	err := ps.presistanceDB.DeleteComment(ctx, id)
	if err != nil {
		log.Printf("Failed to delete Comment{%v}: %v\n", id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to Delete Comment Due to Internal Issues")
	}
	ps.cache.UpdateCommentsCounter(ctx, id, -1)
	return &pb.Response{
		Message: "Comment Deleted Successfully",
	}, nil
}

func (ps *postService) DeleteLike(ctx context.Context, req *pb.DeleteLikeRequest) (*pb.Response, error) {
	id := req.GetPostId()
	user_id := req.GetUserId()

	err := ps.presistanceDB.DeleteLike(ctx, id, user_id)
	if err != nil {
		log.Printf("Failed to delete Like{%v} for user{%v}: %v\n", id, user_id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to Delete Like Due to Internal Issues")
	}
	ps.cache.UpdateLikesCounter(ctx, id, -1)
	return &pb.Response{
		Message: "Like Deleted Successfully",
	}, nil
}

func (ps *postService) GetPosts(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	ids := req.GetPostId()
	var posts []models.Post
	var err error
	posts, err = ps.cache.GetPosts(ctx, ids)
	if err != nil {
		log.Println("Error in Get Posts from Cache: ", err.Error())
		posts, err = ps.presistanceDB.GetPosts(ctx, ids)
		if err != nil {
			log.Println("Error in GetPosts From DB, ", err.Error())
			return nil, err
		}
	}
	resPosts := make([]*pb.Post, 0, len(posts))
	for _, post := range posts {
		resPosts = append(resPosts, &pb.Post{
			UserId:        post.CachedPost.User_id,
			Content:       post.CachedPost.Content,
			CreatedAt:     post.CachedPost.Created_at.Unix(),
			LikesCount:    post.Likes_count,
			CommentsCount: post.Comments_count,
		})
	}
	return &pb.GetPostResponse{
		Post: resPosts,
	}, nil
}

func (ps *postService) StartHealthServer() error {
	router := http.NewServeMux()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
		w.Write([]byte(`{"service": "post_service"}`))
	})

	var handler http.Handler = router
	server := &http.Server{
		Addr:    net.JoinHostPort(ps.config.ServerHost, ps.config.ServerHttpPort),
		Handler: handler,
	}
	log.Printf("PostServer HTTP starting on %s:%s\n", ps.config.ServerHost, ps.config.ServerHttpPort)
	return server.ListenAndServe()
}
