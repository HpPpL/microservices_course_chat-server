package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/HpPpL/microservices_course_chat-server/internal/config"
	desc "github.com/HpPpL/microservices_course_chat-server/pkg/chat_server_v1"
)

// Path to config
var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

func main() {
	flag.Parse()

	err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	grpcConfig, err := config.NewGRPCConfig()
	if err != nil {
		log.Fatalf("failed to get grpc config")
	}

	conn, err := grpc.Dial(grpcConfig.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	c := desc.NewChatServerV1Client(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Create(ctx, &desc.CreateRequest{
		Usernames: []string{"Excepteur", "exercitation"}})

	if err != nil {
		log.Fatalf("failed to create chat: %v", err)
	}
	log.Printf(color.RedString("Chat info:\n"), color.GreenString("%+v", r.GetId()))
}
