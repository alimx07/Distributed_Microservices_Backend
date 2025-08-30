package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/alimx07/Distributed_Microservices_Backend/services_bindings_go"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	repo   *UserRepo
	config Config
}

func NewUserServer(repo *UserRepo, config Config) *UserServer {
	server := &UserServer{
		repo:   repo,
		config: config,
	}
	return server
}

func (s *UserServer) start() error {
	log.Printf("Starting gRPC server on %s:%s", s.config.ServerHost, s.config.ServerPort)
	listener, err := net.Listen("tcp", net.JoinHostPort(s.config.ServerHost, s.config.ServerPort))
	if err != nil {
		log.Fatal("Can not intialized Connection on Host:", net.JoinHostPort(s.config.ServerHost, s.config.ServerPort))
	}
	grpcserver := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcserver, s)
	return grpcserver.Serve(listener)
}

func (s *UserServer) Register(ctx context.Context, rreq *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Register called for email: %s", rreq.GetEmail())
	user := User{
		UserName: rreq.GetUsername(),
		Email:    rreq.GetEmail(),
		Password: rreq.GetPassword(),
	}
	if err := check(user); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	userData := User{
		UserName: user.UserName,
		Password: string(hashed),
		Email:    user.Email,
	}
	err := s.repo.CreateUser(userData)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, "User Already Exists")
	}

	return &pb.RegisterResponse{
		Message: "User Created Successfully",
	}, nil
}

func (s *UserServer) Login(ctx context.Context, lreq *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Login called for email: %s", lreq.GetEmail())
	user := User{
		Email:    lreq.GetEmail(),
		Password: lreq.GetPassword(),
	}

	userData, err := s.repo.GetUserData(user.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, "User does not exists")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(user.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid Credintionals")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
		"iss": "users_service", // mock Issuer
		"sub": user.UserID,
		"exp": time.Now().Add(time.Hour * 1).Unix(),
		"aud": []string{"api_gateway"}, // mock audiance
	})

	tokenString, _ := token.SignedString(s.config.JWTSecret)
	return &pb.LoginResponse{
		Token: tokenString,
	}, nil
}

// Mock Decleration for now
// This Method will talk with Post Service to get user posts
func (s *UserServer) GetUserData(ctx context.Context, usreq *pb.GetUserDataRequest) (*pb.GetUserDataResponse, error) {
	return &pb.GetUserDataResponse{
		User: &pb.User{},
	}, nil
}
