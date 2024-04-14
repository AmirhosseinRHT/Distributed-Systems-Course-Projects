package main

import (
	"fmt"
	"log"
	"strings"

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
	for {
		userInput := 0
		fmt.Printf("Please enter 1 for Server Streaming and 2 for Bidirectional Streaming and 0 to exit: ")
		fmt.Scan(&userInput)

		if userInput == 0 { break }

		var inputNmaes string
		fmt.Printf("please enter Order from this list(note values must be comma seperated and use no space):{banana, apple, orange, grape, red apple, kiwi, mango, pear, cherry, green apple} \n")
		fmt.Scan(&inputNmaes)
		if inputNmaes == "" { continue }

		names := strings.Split(inputNmaes, ",")

		orders := &pb.NamesList{
			Names: names,
		}

		if userInput == 1 {
			callGetOrderServerStream(client, orders)
		} else if userInput == 2 {
			callGetOrderBidirectionalStream(client, orders)
		} else {
			break
		}
	}
}
