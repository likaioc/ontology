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

package dht

import (
	"container/heap"
	"sync"

	"github.com/ontio/ontology/p2pserver/dht/types"
)

type bucket struct {
	entries []*types.Node
}

type routingTable struct {
	mu      sync.RWMutex
	id      types.NodeID          // Local node id
	buckets []*bucket             // Hold routing table
	feedCh  chan *types.FeedEvent // The channel between dht and netserver
}

// init initializes a routing table
func (this *routingTable) init(id types.NodeID, ch chan *types.FeedEvent) {
	this.buckets = make([]*bucket, types.BUCKET_NUM)
	for i := range this.buckets {
		this.buckets[i] = &bucket{
			entries: make([]*types.Node, 0, types.BUCKET_SIZE),
		}
	}

	this.id = id
	this.feedCh = ch
}

// locateBucket locates the bucket with a given id
func (this *routingTable) locateBucket(id types.NodeID) (int, *bucket) {
	dist := logdist(this.id, id)
	if dist == 0 {
		return 0, this.buckets[0]
	}
	return dist - 1, this.buckets[dist-1]
}

// queryNode checks whether exist a node with a given id
func (this *routingTable) queryNode(id types.NodeID) (*types.Node, int) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	index, bucket := this.locateBucket(id)
	for _, node := range bucket.entries {
		if node.ID == id {
			return node, index
		}
	}
	return nil, index
}

// add node to bucket, if bucket contains the node, move it to bucket head
func (this *routingTable) addNode(node *types.Node, bucketIndex int) bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	bucket := this.buckets[bucketIndex]
	for i, entry := range bucket.entries {
		if entry.ID == node.ID {
			copy(bucket.entries[1:], bucket.entries[:i])
			bucket.entries[0] = node
			if this.feedCh != nil {
				feed := &types.FeedEvent{
					EvtType: types.Add,
					Event:   node,
				}
				this.feedCh <- feed
			}
			return true
		}
	}

	// Todo: if the bucket is full, use LRU to replace
	if len(bucket.entries) >= types.BUCKET_SIZE {
		// bucket is full
		return false
	}

	bucket.entries = append(bucket.entries, node)

	copy(bucket.entries[1:], bucket.entries[:])
	bucket.entries[0] = node
	if this.feedCh != nil {
		feed := &types.FeedEvent{
			EvtType: types.Add,
			Event:   node,
		}
		this.feedCh <- feed
	}

	return true
}

// removeNode removes a node with a given id
func (this *routingTable) removeNode(id types.NodeID) {
	this.mu.Lock()
	defer this.mu.Unlock()
	_, bucket := this.locateBucket(id)
	entries := bucket.entries[:0]
	var node *types.Node
	for _, entry := range bucket.entries {
		if entry.ID != id {
			entries = append(entries, entry)
		} else {
			node = entry
		}
	}
	bucket.entries = entries

	if node != nil && this.feedCh != nil {
		feed := &types.FeedEvent{
			EvtType: types.Del,
			Event:   node,
		}
		this.feedCh <- feed
	}
}

// getClosestNodes returns the num nodes in the table that are closest to a given id
func (this *routingTable) getClosestNodes(num int, targetID types.NodeID) types.ClosestList {
	this.mu.RLock()
	defer this.mu.RUnlock()
	closestList := make(types.ClosestList, 0, num)
	heap.Init(&closestList)
	index, _ := this.locateBucket(targetID)
	buckets := []int{index}
	i := index - 1
	j := index + 1

	for len(buckets) < types.BUCKET_NUM {
		if j < types.BUCKET_NUM {
			buckets = append(buckets, j)
		}
		if i >= 0 {
			buckets = append(buckets, i)
		}
		i--
		j++
	}

	for _, index := range buckets {
		for _, entry := range this.buckets[index].entries {
			push(entry, targetID, closestList, num)
			if closestList.Len() >= num {
				return closestList
			}
		}
	}
	return closestList
}

// getTotalNodeNumInBukcet returns the number of node in a specified bucket
func (this *routingTable) getTotalNodeNumInBukcet(bucket int) int {
	this.mu.RLock()
	defer this.mu.RUnlock()
	b := this.buckets[bucket]
	if b == nil {
		return 0
	}

	return len(b.entries)
}

// getLastNodeInBucket returns last node in a specified bucket
func (this *routingTable) getLastNodeInBucket(bucket int) *types.Node {
	this.mu.RLock()
	defer this.mu.RUnlock()
	b := this.buckets[bucket]
	if b == nil || len(b.entries) == 0 {
		return nil
	}

	return b.entries[len(b.entries)-1]
}

func (this *routingTable) totalNodes() int {
	this.mu.RLock()
	defer this.mu.RUnlock()
	var num int
	for _, bucket := range this.buckets {
		num += len(bucket.entries)
	}
	return num
}

// push adds the given node to the list, keeping the total size below maxElems
func push(n *types.Node, targetID types.NodeID, closestList types.ClosestList, maxElems int) {
	distance := getDistance(targetID, n.ID)
	if closestList.Len() < maxElems {
		item := &types.Item{
			Entry:    n,
			Distance: distance,
		}
		heap.Push(&closestList, item)
	} else {
		if closestList[0].Distance <= distance {
			return
		}
		closestList.Update(closestList[0], n, distance)
	}
}

// getDistance returns the distance between nodes
func getDistance(id1, id2 types.NodeID) int {
	dist := logdist(id1, id2)
	return dist
}

// table of leading zero counts for bytes [0..255]
var lzcount = [256]int{
	8, 7, 6, 6, 5, 5, 5, 5,
	4, 4, 4, 4, 4, 4, 4, 4,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// logdist returns the logarithmic distance between a and b, log2(a ^ b).
func logdist(a, b types.NodeID) int {
	lz := 0
	for i := range a {
		x := a[i] ^ b[i]
		if x == 0 {
			lz += 8
		} else {
			lz += lzcount[x]
			break
		}
	}
	return len(a)*8 - lz
}
