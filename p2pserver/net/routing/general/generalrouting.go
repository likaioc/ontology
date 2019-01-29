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

package general

import (
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"

	msgCommon "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	putils "github.com/ontio/ontology/p2pserver/message/utils"
	ontNet "github.com/ontio/ontology/p2pserver/net"
	"github.com/ontio/ontology/p2pserver/peer"
)

type generalRouting struct{
	netL ontNet.NetLayer
}

func (this* generalRouting) Init(netL ontNet.NetLayer, mRouterRegister putils.MessageRouterRegister) error {

	this.netL = netL

	mRouterRegister.RegisterMsgHandler(msgCommon.GetADDR_TYPE, addrReqHandle)
	mRouterRegister.RegisterMsgHandler(msgCommon.ADDR_TYPE, addrHandle)
	return nil
}


//reqNbrList ask the peer for its neighbor list
func (this *generalRouting) reqNbrList(p *peer.Peer) {

	msg := msgpack.NewAddrReq()
	if this.netL != nil && this.netL.IsPeerEstablished(p) {
		go this.netL.Send(p, msg, false)
	}
}

func (this* generalRouting) Start() error {
	seedNodes := make([]string, 0)
	for _, n := range config.DefConfig.Genesis.SeedList {
		ip, err := msgCommon.ParseIPAddr(n)
		if err != nil {
			log.Warnf("[p2p]seed peer %s address format is wrong", n)
			continue
		}
		ns, err := net.LookupHost(ip)
		if err != nil {
			log.Warnf("[p2p]resolve err: %s", err.Error())
			continue
		}
		port, err := msgCommon.ParseIPPort(n)
		if err != nil {
			log.Warnf("[p2p]seed peer %s address format is wrong", n)
			continue
		}
		seedNodes = append(seedNodes, ns[0]+port)
	}

	connPeers := make(map[string]*peer.Peer)
	np := this.netL.GetNp()
	np.Lock()
	for _, tn := range np.List {
		ipAddr, _ := tn.GetAddr16()
		ip := net.IP(ipAddr[:])
		addrString := ip.To16().String() + ":" + strconv.Itoa(int(tn.GetSyncPort()))
		if tn.GetSyncState() == msgCommon.ESTABLISH {
			connPeers[addrString] = tn
		}
	}
	np.Unlock()

	seedConnList := make([]*peer.Peer, 0)
	seedDisconn := make([]string, 0)
	isSeed := false
	for _, nodeAddr := range seedNodes {
		if p, ok := connPeers[nodeAddr]; ok {
			seedConnList = append(seedConnList, p)
		} else {
			seedDisconn = append(seedDisconn, nodeAddr)
		}

		if this.netL.IsOwnAddress(nodeAddr) {
			isSeed = true
		}
	}

	if len(seedConnList) > 0 {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(seedConnList))
		this.reqNbrList(seedConnList[index])
		if isSeed && len(seedDisconn) > 0 {
			index := rand.Intn(len(seedDisconn))
			go this.netL.Connect(seedDisconn[index], false)
		}
	} else { //not found
		for _, nodeAddr := range seedNodes {
			go this.netL.Connect(nodeAddr, false)
		}
	}

	return nil
}

func (this* generalRouting) Stop() error {

	return nil
}

func (this* generalRouting) SetFallbackNodes(recentPeers map[uint32][]string) {

}

func (this* generalRouting) GetFeedCh() chan *ontNet.FeedEvent {

	return nil
}

func (this* generalRouting) GetNbrPeers(peerId msgCommon.P2PNodeID)([]msgCommon.P2PNodeID, error) {

	return nil, nil
}

func NewRouting() *generalRouting {
	return &generalRouting{}
}