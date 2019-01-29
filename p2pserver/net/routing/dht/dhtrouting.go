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
	"errors"
	"github.com/ontio/ontology/p2pserver/common"

	"github.com/ontio/ontology/common/log"
	putils "github.com/ontio/ontology/p2pserver/message/utils"
	ontNet "github.com/ontio/ontology/p2pserver/net"
	"github.com/ontio/ontology/p2pserver/net/routing/dht/types"
)

type dhtRouting struct{
	dht *DHT
}

func (this* dhtRouting) Init(netL ontNet.NetLayer, mRouterRegister putils.MessageRouterRegister) error {

	if this.dht == nil {
		log.Warnf("[p2p]can`t stop: no dht object")
		return errors.New("[p2p]can`t stop: no dht object")
	}

	this.dht.init()

	return nil
}

func (this* dhtRouting) Start() error {

	if this.dht == nil {
		log.Warnf("[p2p]can`t stop: no dht object")
		return errors.New("[p2p]can`t stop: no dht object")
	}

	this.dht.start()

	return nil
}

func (this* dhtRouting) Stop() error {

	if this.dht == nil {
		log.Warnf("[p2p]can`t stop: no dht object")
		return errors.New("[p2p]can`t stop: no dht object")
	}

	this.dht.stop()

	return nil
}

func (this* dhtRouting) SetFallbackNodes(recentPeers map[uint32][]string) {

	if this.dht == nil {
		log.Warnf("[p2p]can`t SetFallbackNodes: no dht object")
	}

	this.dht.setFallbackNodes(recentPeers)
}

func (this* dhtRouting) GetFeedCh() chan *ontNet.FeedEvent {

	if this.dht == nil {
		log.Warnf("[p2p]can`t GetFeedCh: no dht object")
	}

	return this.dht.getFeedCh()
}

func (this* dhtRouting) GetNbrPeers(peerId common.P2PNodeID)([]common.P2PNodeID, error) {

	if this.dht == nil {
		log.Warnf("[p2p]can`t GetNbrPeers: no dht object")
		return nil, errors.New("[p2p]can`t GetNbrPeers: no dht object")
	}
	closestList := this.dht.Resolve(peerId)
	if closestList.Len() == 0 {
		log.Warnf("[p2p]GetNbrPeers: no valid neighbor peer: %d\n", peerId)
		return nil, errors.New("[p2p]GetNbrPeers: no valid neighbor peer")
	}

	nbrPeerIds := make([]common.P2PNodeID, 0)
	for _, item := range closestList {
		id := common.ConvertToP2PNodeID(common.RawP2PNodeID(item.Entry.ID))
		nbrPeerIds = append(nbrPeerIds, id)
	}

	return nbrPeerIds, nil
}

func NewRouting(id common.P2PNodeID) *dhtRouting{

	nodeIDR, err := common.ConvertToRawP2PNodeID(id)
	if err != nil {
		log.Errorf("[p2p]%s", err.Error())
		return nil
	}

	dht := NewDHT(types.NodeID(*nodeIDR))
	return &dhtRouting{dht}
}
