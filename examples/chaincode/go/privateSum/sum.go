package privateSum

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SendRecv interface {
	P2PSend(payload []byte, to ...shim.PeerIdentity)

	P2PRecv(shim.PeerIdentity) (payload []byte)
}

type PrivateSum struct {
	X         uint32 // Our private number
	Neighbors []shim.PeerIdentity
	Comm      SendRecv
}

func (ps *PrivateSum) Sum() uint32 {
	// Create random shares for all neighbors.
	var shares []uint32
	var sumShares uint32
	for i := 0; i < len(ps.Neighbors); i++ {
		randomShare := randomIntN()
		shares = append(shares, randomShare)
		sumShares = sumShares + randomShare
	}

	// Send each neighbor its corresponding share.
	for i, neighbor := range ps.Neighbors {
		payload := make([]byte, 4)
		binary.LittleEndian.PutUint32(payload, shares[i])
		ps.Comm.P2PSend(payload, neighbor)
	}

	var sumNeighborsShares uint32
	// Collect shares from all neighbors.
	for _, neighbor := range ps.Neighbors {
		neighborShare := binary.LittleEndian.Uint32(ps.Comm.P2PRecv(neighbor))
		sumNeighborsShares = sumNeighborsShares + neighborShare
	}

	// Send: X - Sum of our random shares + sumNeighborsShares  to all neighbors.
	payload := make([]byte, 4)
	intermediate := ps.X + sumNeighborsShares - sumShares

	binary.LittleEndian.PutUint32(payload, intermediate)
	ps.Comm.P2PSend(payload, ps.Neighbors...)

	// Receive back the same from all neighbors.
	totalSum := intermediate
	// Collect shares from all neighbors.
	for _, neighbor := range ps.Neighbors {
		neighborShare := binary.LittleEndian.Uint32(ps.Comm.P2PRecv(neighbor))
		totalSum = totalSum + neighborShare
	}

	return totalSum
}

func randomIntN() uint32 {
	b := make([]byte, 4)
	rand.Read(b)
	return binary.LittleEndian.Uint32(b)
}
