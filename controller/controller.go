
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
) *FeedController {
	aUUID, _ := uuid.NewUUID()
	orderbook := feed.NewOrderbookFeed(product)
	newContext, stopFn := context.WithCancel(ctx)
	return &FeedController{
		uuid:      aUUID.String(),
		orderbook: orderbook,
		stopFn:    stopFn,
		ctx:       newContext,
		started:   false,
		outChan:   make(chan (map[string]interface{}), CHANNEL_BUFFER_SIZE),
		inChan:    make(chan (interface{}), CHANNEL_BUFFER_SIZE),
		product:   product,
	}
}

func (fc *FeedController) Start() error {
	if fc.started {
		return errors.New("Feed Controller is already started and cannot be restarted. Please create a new instance")
	}
	fc.startLock.Lock()
	defer fc.startLock.Unlock()

	fc.started = true
	fc.websocket = datasource.NewCoinbaseProWebsocket(fc.ctx, fc.product, fc.outChan, fc.inChan)
	fc.websocket.Start()

	go fc.runOrderbookReporter()
	go fc.runLoop()
	return nil
}

func (fc *FeedController) runOrderbookReporter() {
	timer := time.NewTicker(ORDERBOOK_REPORT_TICKER_SECS * time.Second)
	for {
		select {
		case <-fc.ctx.Done():
			log.Warning("Orderbook reporter shutdown")
			return
		case <-timer.C: