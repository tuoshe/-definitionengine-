package controller

import (
	"context"
	"fmt"

	"pirosb3/real_feed/rpc"
)

type OrderbookGrpcController struct {
	rpc.UnimplementedOrderbookServiceServer
	feedController *FeedController
	product        string
}

func NewOrderbookGrpcController(feedController *FeedController, product string) *OrderbookGrpcController {
	return &Orderboo