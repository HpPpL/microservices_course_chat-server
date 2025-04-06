package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/jackc/pgx/v4/pgxpool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/HpPpL/microservices_course_chat-server/internal/config"
	desc "github.com/HpPpL/microservices_course_chat-server/pkg/chat_server_v1"
)

// Path to config
var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

type server struct {
	desc.UnimplementedChatServerV1Server
	pool *pgxpool.Pool
}

// MessageInfo is a message structure from user to server
type MessageInfo struct {
	from      string
	text      string
	timestamp time.Time
}

var (
	// General PG errors
	errFailedBuildQuery = errors.New("failed to build query")
	errChatDoesntExist  = errors.New("user with current id doesn't exist")

	// Create errors
	errFailedInsertChat = errors.New("failed to insert user")
	errUsernamesISEmpty = errors.New("usernames is empty")

	// Delete errors
	errFailedDeleteUser = errors.New("failed to delete user")
)

// Create part
func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	log.Print("There is create request!")

	users := req.GetUsernames()
	if len(users) == 0 {
		log.Printf("Usernames is empty")
		return &desc.CreateResponse{}, errUsernamesISEmpty
	}

	builderInsert := sq.Insert("chats").
		PlaceholderFormat(sq.Dollar).
		Columns("users").
		Values(users).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return &desc.CreateResponse{}, errFailedBuildQuery
	}

	var ChatID int64
	err = s.pool.QueryRow(ctx, query, args...).Scan(&ChatID)
	if err != nil {
		log.Printf("failed to insert user: %v", err)
		return &desc.CreateResponse{}, errFailedInsertChat
	}

	log.Printf("inserted user with id: %v", ChatID)
	return &desc.CreateResponse{
		Id: ChatID,
	}, nil
}

// Delete part
func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	ChatID := req.GetId()

	builderDelete := sq.Delete("chats").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": ChatID})

	query, args, err := builderDelete.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return &emptypb.Empty{}, errFailedBuildQuery
	}

	res, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("failed to delete chat: %v", err)
		return &emptypb.Empty{}, errFailedDeleteUser
	}

	if res.RowsAffected() == 0 {
		return &emptypb.Empty{}, errChatDoesntExist
	}

	log.Printf("Deleted %d rows", res.RowsAffected())
	return &emptypb.Empty{}, nil
}

// SendMessage part
func (s *server) SendMessage(_ context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	message := MessageInfo{
		from:      req.GetMessage().GetFrom(),
		text:      req.GetMessage().GetText(),
		timestamp: req.GetMessage().GetTimestamp().AsTime(),
	}

	log.Printf("Message for Admin:\n %+v", message)

	return &emptypb.Empty{}, nil
}

func main() {
	flag.Parse()
	ctx := context.Background()

	err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	grpcConfig, err := config.NewGRPCConfig()
	if err != nil {
		log.Fatalf("failed to get grpc config")
	}

	pgConfig, err := config.NewPGConfig()
	if err != nil {
		log.Fatalf("failed to get pg config")
	}

	lis, err := net.Listen("tcp", grpcConfig.Address())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	pool, err := pgxpool.Connect(ctx, pgConfig.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterChatServerV1Server(s, &server{pool: pool})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
