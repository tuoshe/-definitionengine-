package feed

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	TIMEOUT_STALE_BOOK     = 5
	INSUFFICIENT_LIQUIDITY = "INSUFFICIENT_LIQUIDITY"
	BIDS                   = "BIDS"
	ASKS                   = "ASKS"
)

// OrderbookFeed is the primary struct responsible for storage and access of the bids and asks.
// Use this class alongside a websocket feed to keep an up-to-date orderbook, or  you can also
// use this class for one-off orderbook queries.
type OrderbookFeed struct {
	ProductID                string
	bids, asks               sortByOrderbookPrice
	bidsSizeMap, asksSizeMap map[string]float64
	lastEpochSeen            int64
	updateLock               *sync.RWMutex
	snapshotWasSet           bool
}

// GetProduct returns the base and quote assets.
func (of *OrderbookFeed) GetProduct() (string, string) {
	items := strings.Split(of.ProductID, "-")
	if len(items) != 2 {
		panic("Expected 2 items")
	}
	return items[0], items[1]
}

// BuyQuote simulates a market buy of a certain amount. For example, in a
// BTC-USD book, BuyQuote(usdAmount) will return btcToSell.
func (of *OrderbookFeed) BuyQuote(amount float64) (float64, int64, error) {
	return of.performMarketOperationOnQuote(amount, of.bids, of.bidsSizeMap)
}

// SellQuote simulates a market sell of a certain amount. For example, in a
// BTC-USD book, SellQuote(usdAmount) will return btcToBuy.
func (of *OrderbookFeed) SellQuote(amount float64) (float64, int64, error) {
	return of.performMarketOperationOnQuote(amount, of.asks, of.asksSizeMap)
}

// CleanUpOrderbook performs housekeeping on the books, by merging and removing
// orders that have no size.
func (of *OrderbookFeed) CleanUpOrderbook() {
	of.updateLock.Lock()
	defer of.updateLock.Unlock()

	// Process bids
	var newBids, newAsks []*orderbookSortedKey
	for _, bid := range of.bids {
		bidSize := of.bidsSizeMap[bid.Key]
		if bidSize > 0 {
			newBids = append(newBids, bid)
		}
	}
	for _, ask := range of.asks {
		askSize := of.asksSizeMap[ask.Key]
		if askSize > 0 {
			newAsks = append(newAsks, ask)
		}
	}
	of.bids = newBids
	of.asks = newAsks
}

func (of *OrderbookFeed) performMarketOperationOnQuote(amount float64, book sortByOrderbookPrice, sizeMap map[string]float64) (float64, int64, error) {
	if !of.snapshotWasSet {
		return -1, of.lastEpochSeen, errors.New("A snapshot was never set, therefore the orderbook is inaccurate")
	}
	if (time.Now().Unix() - of.lastEpochSeen) > TIMEOUT_STALE_BOOK {
		return -1, of.lastEpochSeen, errors.New("Orderbook is stale")
	}
	if amount <= 0 {
		return -1, of.lastEpochSeen, errors.New("Amount invalid")
	}

	remaining := amount
	baseAmountToPay := 0.0
	for _, orderSet := range book {

		if remaining <= 0 {
			break
		}

		of.updateLock.RLock()
		size, ok := sizeMap[orderSet.Key]
		of.updateLock.RUnlock()
		if !ok {
			log.WithField("key", orderSet.Key).Errorln("Key cannot be found in lookup table.")
			continue
		}
		maxQuoteAmount := orderSet.Value * size
		amountToPurchase := maxQuoteAmount
		if amountToPurchase > remaining {
			amountToPurchase = remaining
		}

		// Perform the transaction
		remaining -= amountToPurchase
		baseAmountToPay += amountToPurchase / orderSet.Value
	}
	if remaining == 0 {
		return baseAmountToPay, of.lastEpochSeen, nil
	}

	return -1, of.lastEpochSeen, errors.New(INSUFFICIENT_LIQUIDITY)
}

// BuyBase simulates a market buy of a certain amount. For example, in a
// BTC-USD book, BuyBase(btcToBuy) will return usdSold.
func (of *OrderbookFeed) BuyBase(amount float64) (float64, int64, error) {
	return of.performMarketOperationOnBase(amount, of.asks, of.asksSizeMap)
}

// SellBase simulates a market buy of a certain amount. For example, in a
// BTC-USD book, SellBase(btcToSell) will return usdPurchased.
func (of *OrderbookFeed) SellBase(amount float64) (float64, int64, error) {
	return of.performMarketOperationOnBase(amount, of.bids, of.bidsSizeMap)
}

func (of *OrderbookFeed) performMarketOperationOnBase(amount float64, book sortByOrderbookPrice, sizeMap map[string]float64) (float64, int64, error) {
	if !of.snapshotWasSet {
		return -1, of.lastEpochSeen, errors.New("A snapshot was never set, therefore the orderbook is inaccurate")
	}
	if (time.Now().Unix() - of.lastEpochSeen) > TIMEOUT_STALE_BOOK {
		return -1, of.lastEpochSeen, errors.New("Orderbook is stale")
	}
	if amount <= 0 {
		return -1, of.lastEpochSeen, errors.New("Amount invalid")
	}
	remainingAmt := amount
	profitMade := 0.0
	for _, orderSet := range book {

		of.updateLock.RLock()
		orderSize := sizeMap[orderSet.Key]
		of.updateLock.RUnlock()

		amountToConsume := orderSize
		if remainingAmt <= amountToConsume {
			amountToConsume = remainingAmt
		}
		remainingAmt -= amountToConsume
		profitMade += amountToConsume * orderSet.Value

		if remainingAmt < 0 {
			return -1, of.lastEpochSeen, errors.New("Implementation error")
		}

		if remainingAmt <= 0 {
			break
		}

	}
	if remainingAmt == 0 {
		return profitMade, of.lastEpochSeen, nil
	}
	return -1, of.lastEpochSeen, errors.New(INSUFFICIENT_LIQUIDITY)
}

func (of *OrderbookFeed) writeUpdate(updates []*Update, side string) bool {
	var selectedBookPtr *sortByOrderbookPrice
	var selectedMap map[string]float64
	if side == BIDS {
		selectedBookPtr = &of.bids
		selectedMap = of.bidsSizeMap
	} else if side == ASKS {
		selectedBookPtr = &of.asks
		selectedMap = of.asksSizeMap
	} else {
		panic("Unsupported side: " + side)
	}

	performedInsert := false
	for _, update := range updates {
		parsedSize, err := strconv.ParseFloat(update.Size, 64)
		if err != nil {
			log.WithField("msg", err.Error()).Errorln("Skipped update due to error")
			continue
		}
		_, ok := selectedMap[update.Price]
		selectedMap[update.Price] = parsedSize
		if !ok {
			parsedPrice, err := strconv.ParseFloat(update.Price, 64)
			if err != nil {
				log.WithField("msg", err.Error()).Errorln("Skipped update due to error")
				continue
			}
			*selectedBookPtr = append(*selectedBookPtr, &orderbookSortedKey{
				Key:   update.Price,
				Value: parsedPrice,
			})
			performedInsert = true
		}
	}
	return performe