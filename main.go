package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"pirosb3/real_feed/controller"
	"pirosb3/real_feed/rpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	market := os.Getenv("MARKET")
	port := "8000"
	ctx, _ := context.WithCancel(context.Background())

	// Start feed controller
	fc := controller.NewFeedController(c