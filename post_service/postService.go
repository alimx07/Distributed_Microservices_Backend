package main

import (
	"context"
	"log"
	"net"

	"github.com/alimx07/Distributed_Microservices_Backend/post_service/models"
	"github.com/alimx07/Distributed_Microservices_Backend/post_service/postRepo"
	pb "github.com/alimx07/Distributed_Microservices_Backend/services_bindings_go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type postService struct {
	pb.UnimplementedPostSeriveServer
	repo   postRepo.PostRepo
	config models.Config
}

func NewPostService(repo postRepo.PostRepo, config models.Config) *postService {
	return &postService{
		repo:   repo,
		config: config,
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
	reflection.Register(grpcserver)
	return grpcserver.Serve(listener)
}
func (ps *postService) CreatePost(ctx context.Context, req *pb.Post) (*pb.Response, error) {
	post := models.Post{
		User_id: req.GetUserId(),
		Content: req.GetContent(),
	}
	err := ps.repo.CreatePost(ctx, post)
	if err != nil {
		log.Printf("Failed to create post for user{%v}\n : {%v}", post.User_id, err.Error())
		return nil, status.Error(codes.Internal, "Post Can`t be Created Due to internal Issues")
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
	err := ps.repo.CreateComment(ctx, comment)
	if err != nil {
		log.Printf("Failed to create comment on post{%v} by user{%v}: %v", comment.Post_id, comment.User_id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to create comment Due to internal Issues")
	}
	return &pb.Response{
		Message: "Comment Created Successfully",
	}, nil
}

func (ps *postService) CreateLike(ctx context.Context, req *pb.Like) (*pb.Response, error) {
	like := models.Like{
		User_id: req.UserId,
		Post_id: req.PostId,
	}
	err := ps.repo.CreateLike(ctx, like)
	if err != nil {
		log.Printf("Failed to create Like on post{%v} by user{%v} : %v", like.Post_id, like.User_id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to create like Due to internal Issues")
	}
	return &pb.Response{
		Message: "Like Created Successfully",
	}, nil
}

func (ps *postService) DeletePost(ctx context.Context, req *pb.Delete) (*pb.Response, error) {
	id := req.GetId()
	user_id := req.GetUserId()

	err := ps.repo.DeletePost(ctx, id, user_id)
	if err != nil {
		log.Printf("Failed to delete post{%v} for user{%v}: %v\n", id, user_id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to Delete Post Due to Internal Issues")
	}
	return &pb.Response{
		Message: "Post Deleted Successfully",
	}, nil
}
func (ps *postService) DeleteComment(ctx context.Context, req *pb.Delete) (*pb.Response, error) {
	id := req.GetId()
	user_id := req.GetUserId()

	err := ps.repo.DeleteComment(ctx, id, user_id)
	if err != nil {
		log.Printf("Failed to delete Comment{%v} for user{%v}: %v\n", id, user_id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to Delete Comment Due to Internal Issues")
	}
	return &pb.Response{
		Message: "Comment Deleted Successfully",
	}, nil
}

func (ps *postService) DeleteLike(ctx context.Context, req *pb.Delete) (*pb.Response, error) {
	id := req.GetId()
	user_id := req.GetUserId()

	err := ps.repo.DeleteLike(ctx, id, user_id)
	if err != nil {
		log.Printf("Failed to delete Like{%v} for user{%v}: %v\n", id, user_id, err.Error())
		return nil, status.Error(codes.Internal, "Failed to Delete Like Due to Internal Issues")
	}
	return &pb.Response{
		Message: "Like Deleted Successfully",
	}, nil
}
