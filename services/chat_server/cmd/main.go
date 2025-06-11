package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"

	chatAPI "github.com/mfisher1411/cli-chat/libraries/api/chat/v1"
)

const grpcPort = 50051

const (
	dbDSN = "host=chat-server_pg port=5432 dbname=chat user=chat-user password=chat-password sslmode=disable"
)

type server struct {
	chatAPI.UnimplementedChatV1Server
	pool *pgxpool.Pool
}

// NullTimeToTimestamp converts a sql.NullTime into a *timestamppb.Timestamp.
func NullTimeToTimestamp(t sql.NullTime) *timestamppb.Timestamp {
	if t.Valid {
		return timestamppb.New(t.Time)
	}
	return nil
}

// Create ...
func (s *server) CreateChat(ctx context.Context, req *chatAPI.CreateChatRequest) (*chatAPI.CreateChatResponse, error) {
	log.Printf("Received create chat request: %+v", req)

	// Делаем запрос на вставку записи в таблицу chat
	builder := sq.Insert(`"chat"`).
		PlaceholderFormat(sq.Dollar).
		Columns("name").
		Values(req.GetName()).
		Suffix("RETURNING id")

	query, args, err := builder.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	var chatID int64
	err = s.pool.QueryRow(ctx, query, args...).Scan(&chatID)
	if err != nil {
		log.Printf("failed to insert chat: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert chat")
	}

	log.Printf("inserted chat with id: %d", chatID)

	return &chatAPI.CreateChatResponse{
		Id: chatID,
	}, nil
}

func (s *server) DeleteChat(ctx context.Context, req *chatAPI.DeleteChatRequest) (*emptypb.Empty, error) {
	log.Printf("Received delete chat request: %+v", req)

	// Валидация ID
	if req.GetId() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid chat id")
	}

	// Делаем запрос на удаление записи в таблице chat
	builder := sq.Delete(`"chat"`).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.GetId()})

	query, args, err := builder.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	res, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("failed to delete chat: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete chat")
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no chat found with id: %d", req.GetId())
		return nil, status.Errorf(codes.NotFound, "chat not found")
	}

	log.Printf("deleted %d row(s) for chat id: %d", rowsAffected, req.GetId())

	log.Printf("deleted chat with id: %d", req.GetId())

	return &emptypb.Empty{}, nil
}

func (s *server) AddUserToChat(ctx context.Context, req *chatAPI.AddUserToChatRequest) (*emptypb.Empty, error) {
	log.Printf("Adding user %d to chat %d", req.GetUserId(), req.ChatId)

	builder := sq.Insert(`"chat_member"`).
		PlaceholderFormat(sq.Dollar).
		Columns("user_id", "chat_id").
		Values(req.GetUserId(), req.GetChatId()).
		Suffix("ON CONFLICT DO NOTHING") // чтобы не падало, если у же есть

	query, args, err := builder.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	_, err = s.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add user to chat")
	}

	return &emptypb.Empty{}, nil
}

func (s *server) SendMessage(ctx context.Context, req *chatAPI.SendMessageRequest) (*emptypb.Empty, error) {
	log.Printf("Received send message chat request: %+v", req)

	builder := sq.Insert(`"message"`).
		PlaceholderFormat(sq.Dollar).
		Columns("chat_id", "sender_id", "content").
		Values(req.GetChatId(), req.GetSenderId(), req.GetContent()).
		Suffix("ON CONFLICT DO NOTHING") // чтобы не падало, если у же есть

	query, args, err := builder.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	res, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add message to table")
	}

	log.Printf("inserted %d rows", res.RowsAffected())

	return &emptypb.Empty{}, nil
}

func (s *server) GetMessages(ctx context.Context, req *chatAPI.GetMessagesRequest) (*chatAPI.GetMessagesResponse, error) {
	log.Printf("Getting list messages from the chat with id %d", req.GetChatId())

	builder := sq.Select("id", "sender_id", "content", "sent_at").
		From(`"message"`).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"chat_id": req.GetChatId()})

	query, args, err := builder.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		log.Printf("failed to select messages: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get messages")
	}

	defer rows.Close()

	var messages []*chatAPI.Message

	for rows.Next() {
		var id int
		var senderID int
		var content string
		var sentAT time.Time

		err = rows.Scan(&id, &senderID, &content, &sentAT)
		if err != nil {
			log.Printf("failed to scan message: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to read message row")
		}

		messages = append(messages, &chatAPI.Message{
			Id:       int64(id),
			ChatId:   req.ChatId,
			SenderId: int64(senderID),
			Content:  content,
			SentAt:   timestamppb.New(sentAT),
		})
	}

	return &chatAPI.GetMessagesResponse{Messages: messages}, nil
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
	srv := &server{
		pool: pool,
	}
	chatAPI.RegisterChatV1Server(s, srv)

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
