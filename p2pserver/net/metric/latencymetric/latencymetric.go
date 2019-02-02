package latencymetric

import (
	"time"

	"github.com/ontio/ontology/p2pserver/common"
)

type LatencyMetric interface {
	StoreLatency(p2pID common.P2PNodeID, latency time.Duration) error
	QueryLatency(p2pID common.P2PNodeID) (time.Duration, error)
}
