package main 

import (
	"context"
	"io"
	"log"

	pb "github.com/m-hariri/basic-go-grpc/proto"
)

func callGetOrderServerStream(client pb.OrderServiceClient, orders *pb.NamesList) {
	log.Printf("Server streaming started")
	stream, err := client.GetOrderServerStreaming(context.Background(), orders) 
	if err != nil {
		log.Fatalf("Could not send orders: %v", err)
	}

	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while streaming %v", err)
		}
		log.Println(message)
	}

	log.Printf("Server streaming finished")
}
