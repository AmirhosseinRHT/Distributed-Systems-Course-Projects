package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/m-hariri/basic-go-grpc/proto"
)

func callGetOrderBidirectionalStream(client pb.OrderServiceClient, orders *pb.NamesList) {
	log.Printf("Bidirectional Streaming started")
	stream, err := client.GetOrderBidirectionalStreaming(context.Background())
	if err != nil {
		log.Fatalf("Could not send orders: %v", err)
	}
	
	waitc := make(chan struct{}) 
	go func() {
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
		close(waitc)
	}()

	for _, name := range orders.Names {
		req := &pb.OrderRequest{  
			Name: name,
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("Error while sending %v", err)
		}
		time.Sleep(2 * time.Second)
	}

	stream.CloseSend()
	<-waitc
	log.Printf("Bidirectional Streaming finished")
}
