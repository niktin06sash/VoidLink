package node

import (
	"log"

	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func (n *Node) startServer() {
	prefix := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1,
	}
	key, err := prefix.Sum([]byte(n.rendezvous))
	if err != nil {
		log.Printf("server: cid error: %v", err)
		return
	}
	err = n.DHT.Provide(n.ctx, key, true)
	if err != nil {
		log.Printf("server: provide error: %v", err)
	} else {
		log.Printf("server: providing key=%s for rendezvous=%s", key.String(), n.rendezvous)
	}
}
