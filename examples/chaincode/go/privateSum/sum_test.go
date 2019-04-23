package privateSum

import (
	"fmt"
	"sync"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type peer struct {
	id           int
	fromChannels map[int]chan []byte
	peers        map[int]*peer
}

func newPeer(id, neighborCount int, peers map[int]*peer) *peer {
	p := &peer{
		id:           id,
		fromChannels: make(map[int]chan []byte, neighborCount),
		peers:        peers,
	}

	for i := 0; i <= neighborCount; i++ {
		if i == p.id {
			continue
		}
		p.fromChannels[i] = make(chan []byte, 1)
	}

	peers[id] = p

	return p
}

func (p *peer) P2PSend(payload []byte, destinations ...shim.PeerIdentity) {
	for _, dest := range destinations {
		dst := int(dest[0])
		receiver := p.peers[dst]
		c := receiver.fromChannels[p.id]
		c <- payload
	}
}

func (p *peer) P2PRecv(id shim.PeerIdentity) (payload []byte) {
	src := int(id[0])
	x := <-p.fromChannels[src]
	return x
}

func TestSum(t *testing.T) {
	peers := make(map[int]*peer)

	n := 10

	var players []*PrivateSum

	for id := 0; id < n; id++ {
		p := newPeer(id, n-1, peers)
		players = append(players, &PrivateSum{
			X:         uint32(id),
			Comm:      p,
			Neighbors: makeNeighbors(id, n),
		})
	}

	var wg sync.WaitGroup
	wg.Add(n)

	for id := 0; id < n; id++ {
		go func(player *PrivateSum) {
			defer wg.Done()
			sum := player.Sum()
			fmt.Println(player.X, sum)
		}(players[id])
	}

	wg.Wait()
}

func makeNeighbors(selfID, total int) []shim.PeerIdentity {
	var res []shim.PeerIdentity
	for i := 0; i < total; i++ {
		if i == selfID {
			continue
		}
		res = append(res, shim.PeerIdentity{byte(i)})
	}
	return res
}
