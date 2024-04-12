package main

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	pb "github.com/nimaCod/orderingsystem/pkg/orderingsystem"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedOrderingsystemServer
}

func (s *server) SearchOrders(ctx context.Context, in *pb.OrderRequest) (stream *pb.OrderResponse) {
	timestampStr := strconv.Itoa(int(time.Now().UnixNano()))
	items := []string{"Got: " + in.GetQuery()}
	return &pb.OrderResponse{Items: items, Timestamp: timestampStr}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterOrderingsystemServer(s, &server{})

	log.Println("Server started on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
