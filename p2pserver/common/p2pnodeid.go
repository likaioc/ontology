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
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/bits"
	"strconv"
	"time"
)

const (
	P2PNodeIDBlank     = ""
	P2PNodeIDSize      = 32
	P2PNodeIDDifficult = 40
	P2PNodeIDTimeOut   = 60 * time.Second
)

type RawP2PNodeID           [P2PNodeIDSize]byte
type P2PNodeIDDynamicFactor [P2PNodeIDSize]byte
type P2PNodeID string

type FinalP2PNodeIDInfo struct {
	P2PNodeID        RawP2PNodeID
	P2PNodeIDDF      P2PNodeIDDynamicFactor
	DifficultStatic  uint64
	DifficultDynamic uint64
	Err              error
}

func calcDifficult(bytes []byte) uint64 {

	for i, b := range bytes {
		if b != 0 {
			return uint64(i*8 + bits.LeadingZeros8(uint8(b)))
		}
	}
	return uint64(len(bytes) * 8)
}

func xor (data1 []byte, data2 []byte)[]byte {

	if len(data1) !=len(data2) {
		return nil
	}
	rData := make([]byte, len(data1))
	for i:= 0; i < len(data1); i++ {
		rData[i] = data1[i] ^ data2[i]
	}

	return rData
}

func GenerateRawNodeIDStatic() ([]byte, uint64, error) {

	rsaPriKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, 0, err
	}
	rsaPubKey    := rsaPriKey.PublicKey
	rsaPubKeyStr := rsaPubKey.N.String() + strconv.Itoa(rsaPubKey.E)
	rNodeID      := sha256.New().Sum([]byte(rsaPubKeyStr))
	hashBytes    := sha256.New().Sum(rNodeID)
	curDifficult := calcDifficult(hashBytes)

	return rNodeID, curDifficult, nil
}

func GenerateRawNodeIDDynamic(
						difficult           uint64,
						rNodeID             []byte,
						finalP2PNodeIdInfo* FinalP2PNodeIDInfo,
						stop                chan struct{},
						nodeIDInfoReceiver  chan FinalP2PNodeIDInfo) {

	for {
		select {
		case <- stop:
			finalP2PNodeIdInfo.Err = errors.New("GenerateRawP2PNodeID timeout")
			nodeIDInfoReceiver <- *finalP2PNodeIdInfo
		default:
			var dynamicFactor [P2PNodeIDSize]byte
			rand.Read(dynamicFactor[:])
			dynamicFactorHash := sha256.New().Sum(xor(dynamicFactor[:], rNodeID))
			curDifficultDynamic := calcDifficult(dynamicFactorHash)

			finalP2PNodeIdInfo.DifficultDynamic = curDifficultDynamic
			copy(finalP2PNodeIdInfo.P2PNodeIDDF[:], dynamicFactor[:])
			if curDifficultDynamic >= difficult {
				finalP2PNodeIdInfo.Err = nil
				nodeIDInfoReceiver <- *finalP2PNodeIdInfo
				return
			}
		}
	}
}

func GenerateRawNodeID(difficult uint64, stop chan struct{}, nodeIDInfoReceiver chan FinalP2PNodeIDInfo){

	var finalP2PNodeIdInfo FinalP2PNodeIDInfo
	for {
		select {
		case <- stop:
			finalP2PNodeIdInfo.Err = errors.New("GenerateRawP2PNodeID timeout")
			nodeIDInfoReceiver <- finalP2PNodeIdInfo
			return
		default:
			rNodeID, curDifficultStatic, err := GenerateRawNodeIDStatic()
			if err != nil {
				finalP2PNodeIdInfo.Err = err
				nodeIDInfoReceiver <- finalP2PNodeIdInfo
				return
			}
			finalP2PNodeIdInfo.DifficultStatic = curDifficultStatic
			copy(finalP2PNodeIdInfo.P2PNodeID[:], rNodeID[:])
			if curDifficultStatic >= difficult {
				GenerateRawNodeIDDynamic(difficult, rNodeID, &finalP2PNodeIdInfo, stop, nodeIDInfoReceiver)
			}
		}
	}
}

func GenerateRawP2PNodeID()(RawP2PNodeID, P2PNodeIDDynamicFactor, error) {

	stop               := make(chan struct{}, 1)
	nodeIDInfoReceiver := make(chan FinalP2PNodeIDInfo, 1)

	go GenerateRawNodeID(P2PNodeIDDifficult, stop, nodeIDInfoReceiver)
	time.Sleep(P2PNodeIDTimeOut)
	stop <- struct{}{}
	p2pNodeInfoID, ok := <- nodeIDInfoReceiver
	if ok {
		return p2pNodeInfoID.P2PNodeID, p2pNodeInfoID.P2PNodeIDDF, nil
	}

	return RawP2PNodeID{}, P2PNodeIDDynamicFactor{}, errors.New("can't generate RawP2PNodeID")
}

func ConvertToP2PNodeID(rP2PNodeID RawP2PNodeID) P2PNodeID {

	for i := 0; i < P2PNodeIDSize/2; i++ {
		rP2PNodeID[i], rP2PNodeID[P2PNodeIDSize-1-i] = rP2PNodeID[P2PNodeIDSize-1-i], rP2PNodeID[i]
	}
	return P2PNodeID(hex.EncodeToString(rP2PNodeID[:]))
}

func ConvertToRawP2PNodeID(p2pId P2PNodeID) (*RawP2PNodeID, error) {

	hexBytes, err := hex.DecodeString(string(p2pId))
	if err == nil {
		l := len(hexBytes)
		if l != P2PNodeIDSize {
			return nil, errors.New("invalid p2pId")
		}
		for i := 0; i < l; i++ {
			hexBytes[i], hexBytes[l-1-i] = hexBytes[l-1-i], hexBytes[i]
		}

		rP2PNodeID := new(RawP2PNodeID)
		copy(rP2PNodeID[:], hexBytes)
	}

	return nil, err
}

func GenerateP2PNodeID() (P2PNodeID, P2PNodeIDDynamicFactor, error) {

	rawP2PNodeID, rawP2PNodeIDDF, err := GenerateRawP2PNodeID()
	if err == nil {
		return ConvertToP2PNodeID(rawP2PNodeID), rawP2PNodeIDDF, nil
	}

	return P2PNodeIDBlank, P2PNodeIDDynamicFactor{}, err
}

func VerifyRawP2PNodeID(rNodeID RawP2PNodeID, dfP2PNodeID P2PNodeIDDynamicFactor) bool {

	dfHash             := sha256.New().Sum(xor(dfP2PNodeID[:], rNodeID[:]))
	curDifficultDynamic := calcDifficult(dfHash)
	if curDifficultDynamic >= P2PNodeIDDifficult {
		return true
	}else {
		return false
	}
}

