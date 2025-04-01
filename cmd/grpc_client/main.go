package main

import (
	"context"
	"log"
	"time"

	"github.com/fatih/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	desc "github.com/HpPpL/microservices_course_chat-server/pkg/chat_server_v1"
)

const (
	address = "localhost:50052"
	authID  = 151417018088165515
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

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
