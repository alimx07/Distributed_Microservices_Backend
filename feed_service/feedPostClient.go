package main

import (
	"context"
	"log"

	"github.com/alimx07/Distributed_Microservices_Backend/feed_service/models"
	pb "github.com/alimx07/Distributed_Microservices_Backend/services_bindings_go"
	"google.golang.org/grpc"
)

type PostClient struct {
	conn   *grpc.ClientConn
	client pb.PostSeriveClient
}

func NewPostClient(target string) (*PostClient, error) {
	conn, err := grpc.NewClient(target)
	if err != nil {
		log.Println("Error in Connection to Post Service: ", err.Error())
		return nil, err
	}
	client := pb.NewPostSeriveClient(conn)
	return &PostClient{
		conn:   conn,
		client: client,
	}, nil
}

func (pc *PostClient) GetPosts(ctx context.Context, items []models.FeedItem) ([]*pb.FeedPost, []int32, error) {
	var ids []int64
	for _, item := range items {
		ids = append(ids, item.PostId)
	}
	req := &pb.GetPostRequest{
		PostId: ids,
	}
	res, err := pc.client.GetPosts(ctx, req)
	if err != nil {
		log.Println("Error in Fetching posts: ", err.Error())
	}
	posts := make([]*pb.FeedPost, 0, len(res.Post))
	usersID := make([]int32, 0, len(res.Post))
	for _, pbPost := range res.Post {
		posts = append(posts, &pb.FeedPost{
			Content:       pbPost.Content,
			CreatedAt:     pbPost.CreatedAt,
			LikesCount:    pbPost.LikesCount,
			CommentsCount: pbPost.CommentsCount,
		})
		usersID = append(usersID, pbPost.UserId)
	}
	return posts, usersID, nil
}
func (pc *PostClient) Close() error {
	if pc.conn != nil {
		return pc.conn.Close()
	}
	return nil
}
