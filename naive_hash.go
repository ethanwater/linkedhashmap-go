package ringhash

import (
	"errors"
	"fmt"
)


func (lhm *LinkedHashMap[K, V]) DistributeMap(servers []Server[K, V]) error {
	//variable used for error checking, if server doesn't exist
	//not the preffered method, just used as an example
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


