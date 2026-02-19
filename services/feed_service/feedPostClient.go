package main

import (
	"context"
	"log"

	"github.com/alimx07/Distributed_Microservices_Backend/services/feed_service/models"
	pb "github.com/alimx07/Distributed_Microservices_Backend/services/services_bindings_go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PostClient struct {
	conn   *grpc.ClientConn
	client pb.PostSeriveClient
}

func NewPostClient(target string) (*PostClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (pc *PostClient) GetPosts(ctx context.Context, items []models.FeedItem) ([]*pb.FeedPost, []string, error) {
	var ids []string
	for _, item := range items {
		ids = append(ids, item.PostId)
	}
	req := &pb.GetPostRequest{
		PostId: ids,
	}
	res, err := pc.client.GetPosts(ctx, req)
	if err != nil {
		log.Println("Error in Fetching posts: ", err.Error())
		return nil, nil, err
	}
	posts := make([]*pb.FeedPost, 0, len(res.Post))
	usersID := make([]string, 0, len(res.Post))
	for _, pbPost := range res.Post {
		posts = append(posts, &pb.FeedPost{
			UserId:        pbPost.UserId,
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
