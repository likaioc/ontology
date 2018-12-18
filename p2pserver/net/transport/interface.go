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

package transport

import (
	"io"
	"net"
	"time"
)

type RecvStream interface {
	io.Reader
	CanContinue() bool
}

type Connection interface {
	GetRecvStream() (RecvStream, error)
	GetTransportType() byte
	Write(cmdType string,  b []byte) (int, error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetWriteDeadline(t time.Time) error
}

type Listener interface {
	Accept() (Connection, error)
	Close() error
	Addr() net.Addr
}

type Transport interface {
	Dial(addr string) (Connection, error)
	DialWithTimeout(addr string, timeout time.Duration) (Connection, error)
	Listen(port uint16) (Listener, error)
	GetReqInterval() int
	ProtocolCode() int
	ProtocolName() string
}
