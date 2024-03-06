package ringhash

import (
	"errors"
)
// Consistent hashing was designed to avoid the problem of having to reassign every BLOB when a server is added or
// removed throughout the cluster. The central idea is to use a hash function that maps both the BLOB and servers to a
// unit circle, usually 2π radians


//BLOB: a computer data storage approach that manages data as "blobs" or "objects", as opposed to other storage
//architectures like file systems which manages data as a file hierarchy, and block storage which manages data as blocks
//within sectors and tracks. (https://en.wikipedia.org/wiki/Object_storage)

const RING uint64 = 360 //default RING size

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
			hash := Hasher(bytify(kv.key)) % ring 
			focusedServer := servers[determineServer(hash, serverDistance, ring)]
			focusedServer.table.Insert(kv.key, kv.value)
			continue
		}
	}
	return nil
}
