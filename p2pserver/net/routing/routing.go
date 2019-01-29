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
	"github.com/ontio/ontology/p2pserver/common"
	putils "github.com/ontio/ontology/p2pserver/message/utils"
	ontNet "github.com/ontio/ontology/p2pserver/net"
)

type Routing interface {
	Init(netL ontNet.NetLayer, mRouterRegister putils.MessageRouterRegister) error
	Start() error
	Stop() error
	SetFallbackNodes(recentPeers map[uint32][]string)
	GetFeedCh() chan *ontNet.FeedEvent
	GetNbrPeers(peerId common.P2PNodeID)([]common.P2PNodeID, error)
}
