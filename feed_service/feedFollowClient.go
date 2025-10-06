package main

import (
	"context"
	"log"

	pb "github.com/alimx07/Distributed_Microservices_Backend/services_bindings_go"
	"google.golang.org/grpc"
)

type FollowClient struct {
	conn   *grpc.ClientConn
	client pb.FollowServiceClient
}

func NewFollowClient(target string) (*FollowClient, error) {
	conn, err := grpc.NewClient(target)
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

func (fc *FollowClient) GetFollowers(ctx context.Context, id int64) ([]int64, error) {
	req := &pb.GetFollowersReq{
		UserID: id,
	}
	res, err := fc.client.GetFollowers(ctx, req)
	if err != nil {
		log.Printf("Error in Fetching Follower IDs for user: %v , Err: ", id, err.Error())
		return nil, err
	}
	return res.FollowerID, nil
}

func (fc *FollowClient) IsCeleb(ctx context.Context, id int64) (bool, error) {
	req := &pb.IsCelebReq{
		UserID: id,
	}
	res, err := fc.client.IsCeleb(ctx, req)
	if err != nil {
		log.Printf("Error in Fetching Follower IDs for user: %v , Err: ", id, err.Error())
		return false, err
	}
	return res.IsCeleb, nil
}

func (fc *FollowClient) GetCeleb(ctx context.Context, id int64) ([]int64, error) {
	req := &pb.GetFollowersReq{
		UserID: id,
	}
	res, err := fc.client.GetFollowers(ctx, req)
	if err != nil {
		log.Printf("Error in Fetching Follower IDs for user: %v , Err: ", id, err.Error())
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
