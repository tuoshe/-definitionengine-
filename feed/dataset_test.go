
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
		t.Errorf("Expected ETH-DAI, instead got %s-%s", base, quote)
	}
}

func TestFailsForGetPrice(t *testing.T) {
	ob := NewOrderbookFeed("ETH-DAI")
	ob.SetSnapshot(time.Now().Unix(), []*Update{}, []*Update{})
	_, _, err := ob.SellBase(1.2)
	if err == nil {
		t.Error("Expected error to exist, but it was nil")
	}
	if err.Error() != INSUFFICIENT_LIQUIDITY {
		t.Errorf("Error message is incorrect, was %s", err.Error())
	}
}

func TestAddItem(t *testing.T) {
	ob := NewOrderbookFeed("ETH-DAI")
	bids := []*Update{
		&Update{Price: "333.2", Size: "0.5"},
		&Update{Price: "320", Size: "0.5"},
		&Update{Price: "310", Size: "1.5"},
	}