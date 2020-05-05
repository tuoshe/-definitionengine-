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
	return &OrderbookGrpcController{
		feedController: feedController,
		product:        product,
	}
}

func (ob *OrderbookGrpcController) handleResponse(response float64, lastUpdated int64, err error, productRequested string) (*rpc.PricingResponse, error) {
	if ob.product != productRequested {
		return &rpc.PricingResponse{
			Product: ob.product,
			Error:   fmt.Sprintf("Requested quote for feed '%s', but service is serving feed '%s'", productRequested, ob.product),
		}, 