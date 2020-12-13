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
	ctx                 context.Context
	outChan             chan (map[string]interface{})
	inChan              chan (interface{})
	outInternalChan     chan (map[string]interface{})
	timeoutInternalChan chan (bool)
}

// NewCoinbaseProWebsocket creates a new Coinbase Pro websocket feed. The feed will only start running once `.Start()` is called on the websocket.
// The `product` should be a Coinbase Pro ticket (example: "ETH-USD"), the `outChan` and `inChan` passed in allow the feed to send / receive messages
// and allow any external component to interact with the websocket service.
// This websocket is also fault-tolerant, if an update is not received within `heartbeatTTLSecond