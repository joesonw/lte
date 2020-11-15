package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/joesonw/distress/examples/grpc/message"
)

type Server struct {
	UnimplementedEchoServer
}

func (s *Server) Echo(ctx context.Context, req *message.Message) (*message.Message, error) {
	delay := time.Duration(rand.Float64() * 3 * float64(time.Millisecond))
	time.Sleep(delay)
	println("Received: " + req.GetBody())
	return &message.Message{
		Body: "You said: " + req.GetBody(),
	}, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	s := grpc.NewServer()
	RegisterEchoServer(s, &Server{})
	lis, err := net.Listen("tcp", ":10090")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.Serve(lis))
}
