package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/nimaCod/orderingsystem/pkg/orderingsystem"

	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewOrderingsystemClient(conn)

	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SearchOrders(ctx, &pb.OrderRequest{Query:"helloo"})
	if err != nil {
		log.Fatalf("Could not greet: %v", err)
	}
	log.Printf("Response: %s , %s", r.Items[0],r.Timestamp)
}
