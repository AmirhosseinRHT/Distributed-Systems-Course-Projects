package main

import (
	"io"
	"log"
	"strconv"
	"strings"

	pb "github.com/m-hariri/basic-go-grpc/proto"
)

func (s *orderServer) GetOrderBidirectionalStreaming(stream pb.OrderService_GetOrderBidirectionalStreamingServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		log.Printf("Got request with name : %v", req.Name)

		found := false
		for i, item := range ServerOrders {
			if strings.Contains(item, req.Name) {
				found = true
				res := &pb.OrderResponse{
					Message: "Item found! Item number: " + strconv.Itoa(i+1) + ", Item name: " + item,
				}
				if err := stream.Send(res); err != nil {
					return err
				}
			}
		}
		if !found {
			res := &pb.OrderResponse{
				Message: "Item not found for: " + req.Name,
			}
			if err := stream.Send(res); err != nil {
				return err
			}
		}
	}
}
