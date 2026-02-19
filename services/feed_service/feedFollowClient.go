package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/alimx07/Distributed_Microservices_Backend/services/services_bindings_go"
)

type FollowClient struct {
	conn   *grpc.ClientConn
	client pb.FollowServiceClient
}

func NewFollowClient(target string) (*FollowClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("Error in Connection to Follow Service: ", err.Error())
		return nil, err
	}
	client := pb.NewFollowServiceClient(conn)

	return &FollowClient{
		conn:   conn,
		client: client,
	}, nil
}

func (fc *FollowClient) GetFollowers(ctx context.Context, id string) ([]string, error) {
	req := &pb.GetFollowersReq{
		UserId: id,
	}
	res, err := fc.client.GetFollowers(ctx, req)
	if err != nil {
		log.Printf("Error in Fetching Follower IDs for user: %v , Err: %v", id, err.Error())
		return nil, err
	}
	return res.FollowerID, nil
}

func (fc *FollowClient) IsCeleb(ctx context.Context, id string) (bool, error) {
	req := &pb.IsCelebReq{
		UserId: id,
	}
	res, err := fc.client.IsCeleb(ctx, req)
	if err != nil {
		log.Printf("Error in Fetching Follower IDs for user: %v , Err: %v", id, err.Error())
		return false, err
	}
	return res.IsCeleb, nil
}

func (fc *FollowClient) GetCeleb(ctx context.Context, id string) ([]string, error) {
	req := &pb.GetFollowersReq{
		UserId: id,
	}
	res, err := fc.client.GetFollowers(ctx, req)
	if err != nil {
		log.Printf("Error in Fetching Follower IDs for user: %v , Err: %v", id, err.Error())
		return nil, err
	}
	return res.FollowerID, nil
}

func (fc *FollowClient) Close() error {
	if fc.conn != nil {
		return fc.conn.Close()
	}
	return nil
}
