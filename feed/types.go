package feed

import "time"

type Update struct {
	Price string
	Size  string
}

type orderbookSortedKey str