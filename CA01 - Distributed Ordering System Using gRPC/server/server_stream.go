package main

import (
	"log"
	"strings"
	"time"

	pb "github.com/m-hariri/basic-go-grpc/proto"
)

func (s *orderServer) GetOrderServerStreaming(req *pb.NamesList, stream pb.OrderService_GetOrderServerStreamingServer) error {
	log.Printf("Got request with names: %v", req.Names)
	for _, name := range req.Names {
		found := false

		for _, item := range ServerOrders {
			if strings.Contains(item, name) {
				found = true

				res := &pb.OrderResponse{
					Message: "Item found: " + item,
				}
				if err := stream.Send(res); err != nil {
					return err
				}
			}
		}
		if !found {
			res := &pb.OrderResponse{
				Message: "Item not found for: " + name,
			}
			if err := stream.Send(res); err != nil {
				return err
			}
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}
