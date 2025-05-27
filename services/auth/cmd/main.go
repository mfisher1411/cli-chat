package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/brianvoe/gofakeit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	userAPI "github.com/mfisher1411/cli-chat/libraries/api/user/v1"
)

const grpcPort = 50051

type server struct {
	userAPI.UnimplementedUserV1Server
}

// Get ...
func (s *server) Get(ctx context.Context, req *userAPI.GetRequest) (*userAPI.GetResponse, error) {
	log.Printf("Received get user request: %+v", req)

	return &userAPI.GetResponse{
		Id:        req.GetId(),
		Name:      gofakeit.Name(),
		Email:     gofakeit.Email(),
		Role:      userAPI.UserRole_USER,
		CreatedAt: timestamppb.New(gofakeit.Date()),
		UpdatedAt: timestamppb.New(gofakeit.Date()),
	}, nil
}

func (s *server) Create(ctx context.Context, req *userAPI.CreateRequest) (*userAPI.CreateResponse, error) {
	log.Printf("Received create user request: %+v", req)
	return &userAPI.CreateResponse{
		Id: gofakeit.Int64(),
	}, nil
}

func (s *server) Update(ctx context.Context, req *userAPI.UpdateRequest) (*emptypb.Empty, error) {
	log.Printf("Received update user request: %+v", req)
	return &emptypb.Empty{}, nil
}

func (s *server) Delete(ctx context.Context, req *userAPI.DeleteRequest) (*emptypb.Empty, error) {
	log.Printf("received delete user request: %+v", req)
	return &emptypb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	userAPI.RegisterUserV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
