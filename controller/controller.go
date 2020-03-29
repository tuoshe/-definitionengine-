
package controller

import (
	"context"
	"errors"
	"pirosb3/real_feed/datasource"
	"pirosb3/real_feed/feed"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

const ORDERBOOK_REPORT_TICKER_SECS = 2
const CHANNEL_BUFFER_SIZE = 20
const TS_LAYOUT = "2006-01-02T15:04:05.000000Z"

func DateStringToUnixEpoch(timestamp string) (int64, error) {
	t, err := time.Parse(TS_LAYOUT, timestamp)
	if err != nil {
		return -1, err
	}
	return int64(t.Unix()), nil
}

var (
	heartbeatTicker = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "heartbeat",
		Help:      "Counts heartbeats from the websocket feed",
		Namespace: "feed",
	}, []string{"uuid", "market"})
	orderbookDepthGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "orderbookDepth",
		Help:      "Orderbook Depth",
		Namespace: "feed",
	}, []string{"uuid", "market", "side"})
)

type FeedController struct {
	orderbook *feed.OrderbookFeed
	websocket *datasource.CoinbaseProWebsocket
	ctx       context.Context
	startLock sync.Mutex
	stopFn    context.CancelFunc
	started   bool
	outChan   chan (map[string]interface{})
	inChan    chan (interface{})
	product   string
	uuid      string
}

func NewFeedController(
	ctx context.Context,
	product string,