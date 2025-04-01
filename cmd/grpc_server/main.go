package main

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"sync"
	"time"

	"crypto/rand"
	"math/big"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	//"google.golang.org/protobuf/types/known/timestamppb"

	desc "github.com/HpPpL/microservices_course_chat-server/pkg/chat_server_v1"
)

const grpcPort = 50052

type server struct {
	desc.UnimplementedChatServerV1Server
}

type MessageInfo struct {
	from      string
	text      string
	timestamp time.Time
}
type ChatMap struct {
	elems map[int64][]string
	m     sync.RWMutex
}

var chats = &ChatMap{
	elems: make(map[int64][]string),
}

var (
	// Create errors
	usernamesIsEmpty   = errors.New("usernames is empty")
	idGenerationFailed = errors.New("id generation failed")

	// Delete errors
	badID = errors.New("id is incorrect")
)

// Create part
func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	users := req.GetUsernames()
	if len(users) == 0 {
		log.Printf("Usernames is empty")
		return &desc.CreateResponse{}, usernamesIsEmpty
	}

	maxNum := big.NewInt(0).Lsh(big.NewInt(1), 63)
	n, err := rand.Int(rand.Reader, maxNum)
	if err != nil {
		log.Print("Id generation failed")
		return &desc.CreateResponse{}, idGenerationFailed
	}
	// Можно потом добавить проверку на существование такого айди
	id := n.Int64()

	chats.m.Lock()
	defer chats.m.Unlock()
	chats.elems[id] = users

	return &desc.CreateResponse{Id: id}, nil
}

// Delete part
func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	id := req.GetId()

	if _, ok := chats.elems[id]; !ok {
		log.Printf("Bad id")
		return &emptypb.Empty{}, badID
	}

	chats.m.Lock()
	defer chats.m.Unlock()

	delete(chats.elems, id)
	return &emptypb.Empty{}, nil
}

// SendMessage part
func (s *server) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	message := MessageInfo{
		from:      req.GetMessage().GetFrom(),
		text:      req.GetMessage().GetText(),
		timestamp: req.GetMessage().GetTimestamp().AsTime(),
	}

	log.Printf("Message for Admin:\n %+v", message)

	return &emptypb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterChatServerV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
