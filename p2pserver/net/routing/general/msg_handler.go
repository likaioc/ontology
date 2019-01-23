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
	"net"
	"strconv"

	evtActor "github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"

	msgCommon "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/msg_pack"
	msgTypes "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
)

// AddrReqHandle handles the neighbor address request from peer
func addrReqHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive addr request message", data.Addr, data.Id)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in AddrReqHandle")
		return
	}

	var addrStr []msgCommon.PeerAddr
	addrStr = p2p.GetNeighborAddrs()
	//check mask peers
	mskPeers := config.DefConfig.P2PNode.ReservedCfg.MaskPeers
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(mskPeers) > 0 {
		for i := 0; i < len(addrStr); i++ {
			var ip net.IP
			ip = addrStr[i].IpAddr[:]
			address := ip.To16().String()
			for j := 0; j < len(mskPeers); j++ {
				if address == mskPeers[j] {
					addrStr = append(addrStr[:i], addrStr[i+1:]...)
					i--
					break
				}
			}
		}

	}
	msg := msgpack.NewAddrs(addrStr)
	err := p2p.Send(remotePeer, msg, false)
	if err != nil {
		log.Warn(err)
		return
	}
}

// AddrHandle handles the neighbor address response message from peer
func addrHandle(data *msgTypes.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]handle addr message", data.Addr, data.Id)

	var msg = data.Payload.(*msgTypes.Addr)
	for _, v := range msg.NodeAddrs {
		var ip net.IP
		ip = v.IpAddr[:]
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))

		if v.ID == p2p.GetID() {
			continue
		}

		if p2p.NodeEstablished(v.ID) {
			continue
		}

		if ret := p2p.GetPeerFromAddr(address); ret != nil {
			continue
		}

		if v.Port == 0 {
			continue
		}
		if p2p.IsAddrFromConnecting(address) {
			continue
		}
		log.Debug("[p2p]connect ip address:", address)
		go p2p.Connect(address, false)
	}
}
