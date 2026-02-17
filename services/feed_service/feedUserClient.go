package main

import (
	"context"
	"log"

	pb "github.com/alimx07/Distributed_Microservices_Backend/services/services_bindings_go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

func NewUserClient(target string) (*UserClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("Error in Connection to User Service: ", err.Error())
		return nil, err
	}
	client := pb.NewUserServiceClient(conn)
	return &UserClient{
		conn:   conn,
		client: client,
	}, nil
}

func (uc *UserClient) GetUsersData(ctx context.Context, ids []string) (map[string]string, error) {
	req := &pb.GetUsersDataRequest{UserId: ids}
	res, err := uc.client.GetUsersData(ctx, req)
	if err != nil {
		log.Println("Failed to get userData from Users Service:", err.Error())
		return nil, err
	}
	users := make(map[string]string)
	for i := range len(res.UserId) {
		users[res.UserId[i]] = res.Username[i]
	}
	return users, nil
}

func (uc *UserClient) Close() error {
	if uc.conn != nil {
		uc.conn.Close()
	}
	return nil
}
