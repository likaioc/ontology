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

package routing

import (
	"errors"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	ptypes "github.com/ontio/ontology/p2pserver/message/types"
	putils "github.com/ontio/ontology/p2pserver/message/utils"
	ontNet "github.com/ontio/ontology/p2pserver/net"
	"github.com/ontio/ontology/p2pserver/net/routing/dht"
	"github.com/ontio/ontology/p2pserver/net/routing/general"
	netRouting "github.com/ontio/ontology/p2pserver/net/routing/interface"
)

type RoutingFacade interface {
	Init(rMode constants.RoutingModeType,  netL ontNet.NetLayer, mRouterRegister putils.MessageRouterRegister) error
	Start() error
	Stop() error
	SetFallbackNodes(recentPeers map[uint32][]string)
	TransmitMsgReq(targetPeerId uint64, msg ptypes.Message) error
}

type routingFacade struct {
	netL ontNet.NetLayer
	routingMap map[constants.RoutingModeType]netRouting.Routing
}

func NewRoutingFacade() RoutingFacade {
	return &routingFacade{}
}

func (this* routingFacade) initGeneralRouting(netL ontNet.NetLayer, mRouterRegister putils.MessageRouterRegister) error {

	genRouting :=  general.NewRouting()
	err := genRouting.Init(netL, mRouterRegister)
	this.routingMap[constants.P2P_ROUTING_GENERAL] = genRouting

	return err
}

func (this* routingFacade) initDHTRouting(netL ontNet.NetLayer, mRouterRegister putils.MessageRouterRegister) error {

	dhtRouting := dht.NewRouting(this.netL.GetID())
	err := dhtRouting.Init(netL, mRouterRegister)
	if err != nil {
		log.Error("[p2p]Init DHT routing error")
		return err
	}

	feedChan := dhtRouting.GetFeedCh()
	if feedChan != nil {
		this.netL.SetFeedCh(feedChan)
	}else {
		log.Error("[p2p]The DHT routing hasn't feed chan,  so the network can't receive from DHT routing")
		return errors.New("[p2p]The DHT routing hasn't feed chan,  so the network can't receive from DHT routing")
	}

	this.routingMap[constants.P2P_ROUTING_DHT] = dhtRouting

	return err
}

func (this* routingFacade) Init(rMode constants.RoutingModeType, netL ontNet.NetLayer, mRouterRegister putils.MessageRouterRegister) error {

	log.Infof("[p2p]Init routing, routing mode=%s", constants.RModeMap[rMode])

	this.netL = netL

	this.routingMap = make(map[constants.RoutingModeType]netRouting.Routing)
	switch rMode {
	case constants.P2P_ROUTING_GENERAL:
		return this.initGeneralRouting(netL, mRouterRegister)
	case constants.P2P_ROUTING_DHT:
		return this.initDHTRouting(netL, mRouterRegister)
	case constants.P2P_ROUTING_ALL:
		err1 := this.initGeneralRouting(netL, mRouterRegister)
		err2 := this.initDHTRouting(netL, mRouterRegister)

		var errs common.Errors
		errs = append(errs, err1)
		errs = append(errs, err2)

		return errs.Err()
	default:
		log.Error("[p2p]Unknown routing mode, rMode=%d", rMode)
		return errors.New("[p2p]Unknown routing mode")
	}

	return nil
}

func (this* routingFacade) Start() error {

	for _, r := range this.routingMap {
		if r.GetFeedCh() != nil && this.netL != nil {
			go this.netL.LoopRecvRoutingMsg()
		}
		r.Start()
	}

	return nil
}

func (this* routingFacade) Stop() error {

	var errs common.Errors
	for _, r := range this.routingMap {
		er := r.Stop()
		errs = append(errs, er)
	}

	return errs.Err()
}

func (this* routingFacade) SetFallbackNodes(recentPeers map[uint32][]string) {

	if r, ok := this.routingMap[constants.P2P_ROUTING_DHT]; ok {
		r.SetFallbackNodes(recentPeers)
	}else {
		log.Errorf("[p2p]there is no valid routing for SetFallbackNodes")
	}
}

func (this* routingFacade) TransmitMsgReq(targetPeerId uint64, msg ptypes.Message) error {

	if r, ok := this.routingMap[constants.P2P_ROUTING_DHT]; ok {
		nbrPeers, _ := r.GetNbrPeers(targetPeerId)
		for _, id := range nbrPeers {
			peer := this.netL.GetPeer(id)
			if peer == nil{
				continue
			}
			err := this.netL.Send(peer, msg, true)
			if err != nil {
				log.Warnf("[p2p]can`t transmit msg to %s, send msg err: %s", peer.GetAddr(), err)
			}
		}
	}else {
		log.Errorf("[p2p]there is no valid routing for transmitting Msg " + msg.CmdType())
		return errors.New("[p2p]there is no valid routing for transmitting Msg " + msg.CmdType())
	}

	return nil
}


