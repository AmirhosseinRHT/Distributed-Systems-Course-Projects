package main

import ( 
	"log"

	pb "github.com/m-hariri/basic-go-grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	port = ":8080"
)

func main() {
	conn, err := grpc.Dial("localhost"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	orders := &pb.NamesList{
		Names: []string{"salt", "mango", "ban"},
	}


	//callGetOrderServerStream(client, orders)
	callGetOrderBidirectionalStream(client, orders)
}
