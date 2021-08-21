package datasource

import (
	"context"
	"testing"
	"time"
)

func TestContextShutsDown(t *testing.T) {
	outChan := make(chan (map[string]interfa