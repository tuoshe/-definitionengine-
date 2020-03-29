
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