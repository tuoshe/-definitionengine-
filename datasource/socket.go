package datasource

import (
	"context"
	"errors"
	"net/http"
	"pirosb3/real_feed/feed"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

const heartbeatTTLSeconds = 4

var (
	pricingProm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "pricing",
		Help:      "Orderbook visualization for 50 ETH",
		Namespace: "feed",
	}, []string{"uuid", "market"})

	updatesCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "updates",
		Help:      "Shows the frequency of orderbook updates coming out of the websocket",
		Namespace: "feed",
	}, []string{"uuid", "market"})

	droppedPacketsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "droppedPackets",
		Help:      "Shows the amount of dropped packets",
		Namespace: "feed",
	}, []string{"uuid", "market"})

	timeoutsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "timeout",
		Help:      "Shows the frequency of timeouts",
		Namespace: "feed",
	}, []string{"uuid", "market"})
	wsLatency = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:      "websocketUpdateFrequency",
		Help:      "Shows the frequency of websocket responses",
		Namespace: "feed",
	}, []string{"uuid", "market"})
)

type CoinbaseProWebsocket struct {
	uuid                string
	startLock           sync.Mutex
	websocketConn       *websocket.Conn
	product             string
	running             bool
	ctx                 con