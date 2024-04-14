package main

import (
	"log"
	"net"

	pb "github.com/m-hariri/basic-go-grpc/proto"
	"google.golang.org/grpc"
)

const (
	port = ":8080"
)

type orderServer struct {
	pb.OrderServiceServer
}

var ServerOrders = []string{"banana", "apple", "orange", "grape", "red apple",
	"kiwi", "mango", "pear", "cherry", "green apple"}

func main() {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to start server %v", err)
	}
	grpcServer := grpc.NewServer()

	pb.RegisterOrderServiceServer(grpcServer, &orderServer{})
	log.Printf("Server started at %v", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
}
