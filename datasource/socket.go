package datasource

import (
	"context"
	"errors"
	"net/http"
	"pirosb3/real_feed/feed"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_gol