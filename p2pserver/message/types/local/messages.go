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

 package local

import (
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver/common"
	ptypes "github.com/ontio/ontology/p2pserver/message/types"
)

type AppendPeerID struct {
	ID common.P2PNodeID // The peer id
}

type RemovePeerID struct {
	ID common.P2PNodeID // The peer id
}

type AppendHeaders struct {
	FromID  common.P2PNodeID          // The peer id
	Headers []*types.Header // Headers to be added to the ledger
}

type AppendBlock struct {
	FromID    common.P2PNodeID       // The peer id
	BlockSize uint32                 // Block size
	Block     *types.Block           // Block to be added to the ledger
}

type RelayTransmitConsensusMsgReq struct {
	Target common.P2PNodeID
	Msg    ptypes.Message
}

