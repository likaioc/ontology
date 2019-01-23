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
	"io"

	"github.com/ontio/ontology/common"
	pCom "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/net/routing/dht/types"
)

type DHTPong struct {
	Version      uint16
	FromID       types.NodeID
	SrcEndPoint  EndPoint
	DestEndPoint EndPoint
}

func (this *DHTPong) CmdType() string {
	return pCom.DHT_PONG
}

//Serialize message
func (this DHTPong) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint16(this.Version)
	sink.WriteVarBytes(this.FromID[:])

	sink.WriteVarBytes(this.SrcEndPoint.Addr[:])
	sink.WriteUint16(this.SrcEndPoint.UDPPort)
	sink.WriteUint16(this.SrcEndPoint.TCPPort)
	sink.WriteVarBytes(this.DestEndPoint.Addr[:])
	sink.WriteUint16(this.DestEndPoint.UDPPort)
	sink.WriteUint16(this.DestEndPoint.TCPPort)
	return nil
}

//Deserialize message
func (this *DHTPong) Deserialization(source *common.ZeroCopySource) error {
	var (
		eof       bool
		irregular bool
		buf       []byte
	)

	this.Version, eof = source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}

	buf, _, irregular, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}
	copy(this.FromID[:], buf)

	buf, _, irregular, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}
	copy(this.SrcEndPoint.Addr[:], buf)

	this.SrcEndPoint.UDPPort, eof = source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.SrcEndPoint.TCPPort, eof = source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}

	buf, _, irregular, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}
	copy(this.DestEndPoint.Addr[:], buf)

	this.DestEndPoint.UDPPort, eof = source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}

	this.DestEndPoint.TCPPort, eof = source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
