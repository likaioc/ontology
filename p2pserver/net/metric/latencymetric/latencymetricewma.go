/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package latencymetric

import (
	"fmt"
	"sync"
	"time"

	"github.com/ontio/ontology/p2pserver/common"
)

const (
	LATENCY_METRIC_ALPHA = 0.1
)

type  LatencyMetricEWMA struct {
    lock sync.RWMutex
	latencyMetricMap map[common.P2PNodeID]time.Duration
}

func NewLatencyMetricEWMA() LatencyMetric {
	return &LatencyMetricEWMA{latencyMetricMap: make(map[common.P2PNodeID]time.Duration)}
}

//the function implements EWMA(Exponentially Weighted Moving Average)
func (this* LatencyMetricEWMA) StoreLatency(p2pID common.P2PNodeID, latency time.Duration) error {

	this.lock.Lock()
	defer this.lock.Unlock()

	if latV, ok := this.latencyMetricMap[p2pID]; ok {
		ewmaCur := float64(latency)
		ewmaPre := float64(latV)
		ewmaNew := LATENCY_METRIC_ALPHA * ewmaCur + (1 - LATENCY_METRIC_ALPHA) * ewmaPre
		this.latencyMetricMap[p2pID] = time.Duration(ewmaNew)
	} else {
		this.latencyMetricMap[p2pID] = latency
	}

	return nil
}

func (this* LatencyMetricEWMA) QueryLatency(p2pID common.P2PNodeID) (time.Duration, error) {

	this.lock.Lock()
	defer this.lock.Unlock()

	if latV, ok := this.latencyMetricMap[p2pID]; ok {
		return latV, nil
	}

	return 0, fmt.Errorf("can't query the respondign latency of %s", p2pID)
}
