package controller

import (
	"context"
	"fmt"

	"pirosb3/real_feed/rpc"
)

type OrderbookGrpcController struct {
	rpc.UnimplementedOrderbookServiceServer
	feedController *FeedController
	