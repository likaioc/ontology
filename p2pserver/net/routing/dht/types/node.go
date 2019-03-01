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

package types

import (
	"container/heap"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ontio/ontology/p2pserver/common"
)

const (
	NODE_ID_BITS = common.P2PNodeIDSize * 8
)

// NodeID is a unique identifier for each node.
type NodeID common.RawP2PNodeID

// Bytes returns a byte slice representation of the NodeID
func (n NodeID) Bytes() []byte {
	return n[:]
}

// NodeID prints as a long hexadecimal number.
func (n NodeID) String() string {
	return string(common.ConvertToP2PNodeID(common.RawP2PNodeID(n)))
}

var NilID = NodeID{}

func (n NodeID) IsNil() bool {
	return n == NilID
}

func StringID(in string) (NodeID, error) {
	var id NodeID
	b, err := hex.DecodeString(strings.TrimPrefix(in, "0x"))
	if err != nil {
		return id, err
	} else if len(b) > len(id) {
		return id, fmt.Errorf("wrong length, want %d hex chars", len(b)*2)
	}
	copy(id[:], b)
	return id, nil
}

type Item struct {
	Entry    *Node
	Distance int
	Index    int
}

type ClosestList []*Item

func (cl ClosestList) Len() int { return len(cl) }

func (cl ClosestList) Less(i, j int) bool {
	return cl[i].Distance > cl[j].Distance
}

func (cl ClosestList) Swap(i, j int) {
	cl[i], cl[j] = cl[j], cl[i]
	cl[i].Index = i
	cl[j].Index = j
}

func (cl *ClosestList) Push(x interface{}) {
	n := len(*cl)
	item := x.(*Item)
	item.Index = n
	*cl = append(*cl, item)
}

func (cl *ClosestList) Pop() interface{} {
	old := *cl
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*cl = old[0 : n-1]
	return item
}

func (cl *ClosestList) Update(item *Item, entry *Node, distance int) {
	item.Entry = entry
	item.Distance = distance
	heap.Fix(cl, item.Index)
}
