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

package common

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/pkg/errors"
	"math/bits"
	"time"

	blake2bhash "github.com/minio/blake2b-simd"
	"github.com/ontio/ontology/common"
)

const (
	P2PNODEID_DIFFICULT = 40
	P2PNODEID_SIZE      = 32
	P2PNODEID_TIMEOUT   = 60 * time.Second
	P2PNODEID_BLANK     = ""
)

type RawP2PNodeID [P2PNODEID_SIZE]byte
type P2PNodeID string

type FinalP2PNodeIDInfo struct {
	Nonce     *common.Uint256
	Difficult uint64
	P2PNodeID RawP2PNodeID
	Err       error
}

func calcHashDifficult(hashByte RawP2PNodeID) uint64 {

	for i, b := range hashByte {
		if b != 0 {
			return uint64(i*8 + bits.LeadingZeros8(uint8(b)))
		}
	}
	return uint64(len(hashByte) * 8)
}

func hashDataWithNonce(data []byte, nonce *common.Uint256) RawP2PNodeID {

	first := blake2bhash.Sum512(append(data, nonce.ToArray()...))
	return RawP2PNodeID(sha256.Sum256(first[:]))
}

//_, pubKey, err := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
//serialPubKey := keypair.SerializePublicKey(pubKey)
func generateRawP2PNodeIDSub(data []byte, difficult uint64, startNonce* common.Uint256, stop chan struct{}, nodeIDInfoReceiver chan FinalP2PNodeIDInfo) {

	var finalP2PNodeIdInfo FinalP2PNodeIDInfo
	for nonce := startNonce; ; nonce.Inc() {
		select {
		case <- stop:
			finalP2PNodeIdInfo.Err = errors.New("GenerateRawP2PNodeID timeout")
			nodeIDInfoReceiver <- finalP2PNodeIdInfo
			return
		default:
			rP2PNodeId := hashDataWithNonce(data, nonce)
			currentDifficult := calcHashDifficult(rP2PNodeId)
			if calcHashDifficult(rP2PNodeId) >= difficult {
				finalP2PNodeIdInfo.Nonce     = nonce
				finalP2PNodeIdInfo.Difficult = currentDifficult
				finalP2PNodeIdInfo.P2PNodeID = rP2PNodeId
				finalP2PNodeIdInfo.Err       = nil
				nodeIDInfoReceiver <- finalP2PNodeIdInfo
				return
			}

			if currentDifficult > finalP2PNodeIdInfo.Difficult {
				finalP2PNodeIdInfo.Nonce     = nonce
				finalP2PNodeIdInfo.Difficult = currentDifficult
				finalP2PNodeIdInfo.P2PNodeID = rP2PNodeId
			}
		}
	}
}

func GenerateRawP2PNodeID()(RawP2PNodeID, keypair.PublicKey, error) {

	stop := make(chan struct{}, 1)
	nodeIDInfoReceiver := make(chan FinalP2PNodeIDInfo, 1)
	_, pubKey, err := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	if err == nil {
		serialPubKey := keypair.SerializePublicKey(pubKey)
		go generateRawP2PNodeIDSub(serialPubKey, P2PNODEID_DIFFICULT, &common.Uint256{}, stop, nodeIDInfoReceiver)
		time.Sleep(P2PNODEID_TIMEOUT)
		stop <- struct{}{}

		p2pNodeInfoID, ok := <- nodeIDInfoReceiver
		if ok {
			return p2pNodeInfoID.P2PNodeID, pubKey, nil
		}
	}

	return RawP2PNodeID{}, pubKey, err
}

func ConvertToP2PNodeID(rP2PNodeID RawP2PNodeID) P2PNodeID {

	for i := 0; i < P2PNODEID_SIZE/2; i++ {
		rP2PNodeID[i], rP2PNodeID[P2PNODEID_SIZE-1-i] = rP2PNodeID[P2PNODEID_SIZE-1-i], rP2PNodeID[i]
	}
	return P2PNodeID(hex.EncodeToString(rP2PNodeID[:]))
}

func ConvertToRawP2PNodeID(p2pId P2PNodeID) (*RawP2PNodeID, error) {

	hexBytes, err := hex.DecodeString(string(p2pId))
	if err == nil {
		l := len(hexBytes)
		if l != P2PNODEID_SIZE {
			return nil, errors.New("Invalid p2pId")
		}
		for i := 0; i < l; i++ {
			hexBytes[i], hexBytes[l-1-i] = hexBytes[l-1-i], hexBytes[i]
		}

		rP2PNodeID := new(RawP2PNodeID)
		copy(rP2PNodeID[:], hexBytes)
	}

	return nil, err
}

func GenerateP2PNodeID() (P2PNodeID, keypair.PublicKey, error) {

	rawP2PNodeID, pubKey, err := GenerateRawP2PNodeID()
	if err == nil {
		return ConvertToP2PNodeID(rawP2PNodeID), pubKey, nil
	}

	return "", pubKey, err
}


