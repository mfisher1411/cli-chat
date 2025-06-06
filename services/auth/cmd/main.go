package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	sq "github.com/Masterminds/squirrel"
	userAPI "github.com/mfisher1411/cli-chat/libraries/api/user/v1"
)

const grpcPort = 50051

const (
	dbDSN = "host=localhost port=54321 dbname=auth user=auth-user password=auth-password sslmode=disable"
)

type server struct {
	userAPI.UnimplementedUserV1Server
	pool *pgxpool.Pool
}

func NullTimeToTimestamp(t sql.NullTime) *timestamppb.Timestamp {
	if t.Valid {
		return timestamppb.New(t.Time)
	}
	return nil
}

// Get ...
func (s *server) Get(ctx context.Context, req *userAPI.GetRequest) (*userAPI.GetResponse, error) {
	log.Printf("Received get user request: %+v", req)

	// Делаем запрос на выборку записей из таблицы user
	builderSelect := sq.Select("id", "name", "email", "role", "created_at", "updated_at").
		From(`"user"`).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.GetId()})

	query, args, err := builderSelect.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	var id int
	var name, email string
	var role int
	var createdAt time.Time
	var updatedAt sql.NullTime

	err = s.pool.QueryRow(ctx, query, args...).Scan(&id, &name, &email, &role, &createdAt, &updatedAt)
	if err == pgx.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "user not found")
	} else if err != nil {
		log.Printf("failed to select user: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
	return &userAPI.GetResponse{
		Id:        int64(id),
		Name:      name,
		Email:     email,
		Role:      userAPI.UserRole(role),
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: NullTimeToTimestamp(updatedAt),
	}, nil

}

func (s *server) Create(ctx context.Context, req *userAPI.CreateRequest) (*userAPI.CreateResponse, error) {
	log.Printf("Received create user request: %+v", req)
	// return &userAPI.CreateResponse{
	// 	Id: gofakeit.Int64(),
	// }, nil

	// Делаем запрос на вставку записи в таблицу user
	builderInsert := sq.Insert(`"user"`).
		PlaceholderFormat(sq.Dollar).
		Columns("name", "email", "role").
		Values(req.GetName(), req.GetEmail(), req.GetRole()).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	var userID int64
	err = s.pool.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		log.Printf("failed to insert note: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert user")
	}

	log.Printf("inserted note with id: %d", userID)

	return &userAPI.CreateResponse{
		Id: userID,
	}, nil
}

func (s *server) Update(ctx context.Context, req *userAPI.UpdateRequest) (*emptypb.Empty, error) {
	log.Printf("Received update user request: %+v", req)
	// Делаем запрос на обновление записи в таблице user
	builderUpdate := sq.Update(`"user"`).
		PlaceholderFormat(sq.Dollar).
		Set("name", req.GetName().Value).
		Set("email", req.GetEmail().Value).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": req.GetId()})

	query, args, err := builderUpdate.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	res, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("failed to update note: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update user")
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no user found with id: %d", req.GetId())
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	log.Printf("updated %d rows", res.RowsAffected())
	return &emptypb.Empty{}, nil
}

func (s *server) Delete(ctx context.Context, req *userAPI.DeleteRequest) (*emptypb.Empty, error) {
	log.Printf("received delete user request: %+v", req)
	// Делаем запрос на удаление записи в таблице user
	builderDelete := sq.Delete(`"user"`).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.GetId()})

	query, args, err := builderDelete.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	res, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("failed to delete user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no user found with id: %d", req.GetId())
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	log.Printf("deleted user with id: %d", req.GetId())

	return &emptypb.Empty{}, nil
}

func main() {

	ctx := context.Background()

	// Создаем пул соединений с базой данных
	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	srv := &server{pool: pool}
	userAPI.RegisterUserV1Server(s, srv)

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
