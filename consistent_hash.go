package ringhash

import (
	"errors"
	"fmt"
)

const RING uint64 = 360 //default RING size

func (lhm *LinkedHashMap[K, V]) DistributeMap(servers []Server[K, V]) error {
	//variable used for error checking, if server doesn't exist
	//not the robust method, just used as an example
	var emptyServer Server[K, V]

	if len(servers) == 0 {
		return errors.New("no servers available for distribution")
	}
	firstLayerBuckets := lhm.buckets

	for _, bucket := range firstLayerBuckets {
		for _, kv := range bucket {
			hash := Hasher(bytify(kv.key)) % uint64(len(servers))
			//if hash = number of servers, it's considered out of bounds
			if hash >= uint64(len(servers)) {
				return errors.New("hash overflows the amount of designated servers")
			}
			focusedServer := servers[hash]

			if focusedServer == emptyServer {
				return fmt.Errorf("server at index %d is nil", hash)
			}

			focusedServer.table.Insert(kv.key, kv.value)
		}
	}

	return nil
}

// Consistent hashing was designed to avoid the problem of having to reassign every BLOB when a server is added or
// removed throughout the cluster. The central idea is to use a hash function that maps both the BLOB and servers to a
// unit circle, usually 2π radians
func (lhm *LinkedHashMap[K, V]) DistrbuteConsistentHash(servers []Server[K, V], ring uint64) error {
	if len(servers) == 0 {
		return errors.New("no servers available for distribution")
	}
	serverDistance := ring / uint64(len(servers))

	var determineServer func(uint64, uint64, uint64) int
	determineServer = func(hash, ring, dist uint64) int {
		initialServer := 1

		for i := int(dist); i < int(ring); i += int(dist) {
			if hash < uint64(i) {
				return initialServer
			}
			initialServer++
		}

		return 0
	}

	for _, bucket := range lhm.buckets {
		for _, kv := range bucket {
			hash := Hasher(bytify(kv.key)) % ring //[0, 360)
			focusedServer := servers[determineServer(hash, serverDistance, ring)]
			focusedServer.table.Insert(kv.key, kv.value)
			continue
		}
	}
	return nil
}
