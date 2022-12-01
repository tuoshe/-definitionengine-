package feed

import "time"

type Update struct {
	Price string
	Size  string
}

type orderbookSortedKey struct {
	Value float64
	Key   string
}

type sortByOrderbookPrice []*orderbookSortedKey

type LevelTwoOrderbook struct {
	Bids [][]interface{} `json:"bids"`
	Asks [][]interface{} `json:"asks"`
}

type TickerChannel struct {
	Name       string   `json:"name"`
	ProductIds []string `json:"product_ids"`
}

type WebsocketType struct {
	Type string `json:"type"`
}
type MessageSubscription struct {
	WebsocketType
	ProductIds []string      `json:"product_ids"`
	Channels   []int