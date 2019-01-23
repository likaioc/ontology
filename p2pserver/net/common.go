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

package net

import (
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/peer"
)

type feedType uint8

const (
	Add feedType = iota
	Del
)

type FeedEvent struct {
	EvtType feedType
	Event   interface{}
}

type FeedInfo struct {
	ID      uint64
	IP      string
	TCPPort uint16
}

type NetLayer interface {
	IsOwnAddress(addr string) bool
	GetID() uint64
	SetFeedCh(chan *FeedEvent)
	LoopRecvRoutingMsg()
	GetPeer(uint64) *peer.Peer
	GetNp() *peer.NbrPeers
	IsPeerEstablished(p *peer.Peer) bool
	Connect(addr string, isConsensus bool) error
	Send(p *peer.Peer, msg types.Message, isConsensus bool) error
}



