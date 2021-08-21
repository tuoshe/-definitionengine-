package datasource

import (
	"context"
	"testing"
	"time"
)

func TestContextShutsDown(t *testing.T) {
	outChan := make(chan (map[string]interface{}))
	inChan := make(chan (interface{}))

	ctx, cancelFn := con