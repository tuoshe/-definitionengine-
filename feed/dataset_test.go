
package feed

import (
	"encoding/json"
	"math"
	"net/http"
	"testing"
	"time"
)

func transformToUpdate(input [][]interface{}) []*Update {
	newUpdates := make([]*Update, len(input))
	for idx, val := range input {
		newUpdates[idx] = &Update{
			Price: val[0].(string),
			Size:  val[1].(string),
		}
	}
	return newUpdates
}

func TestCanInitialize(t *testing.T) {
	ob := NewOrderbookFeed("ETH-DAI")
	base, quote := ob.GetProduct()
	if base != "ETH" || quote != "DAI" {