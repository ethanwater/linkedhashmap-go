package ringhash

import (
	"errors"
	"reflect"
	"strconv"
	"unsafe"
)

type Bucket[K comparable, V any] struct {
	key   K
	value V
}

type LinkedHashMap[K comparable, V any] struct {
	buckets [][]Bucket[K, V]
	items   uint64
}

func (*LinkedHashMap[K, V]) New() *LinkedHashMap[K, V] {
	linkedHashMap := LinkedHashMap[K, V]{
		buckets: make([][]Bucket[K, V], 0),
		items:   0,
	}

	return &linkedHashMap
}

func (lhm *LinkedHashMap[K, V]) Resize() {
	var targetSize int
	switch len(lhm.buckets) {
	case 0:
		targetSize = 1
	default:
		targetSize = len(lhm.buckets) * 2
	}

	newBuckets := make([][]Bucket[K, V], targetSize)
	for i := 0; i < targetSize; i++ {
		newBuckets[i] = []Bucket[K, V]{}
	}

	for _, bucket := range lhm.buckets {
		for _, kv := range bucket {
			key := kv.key
			value := kv.value
			hash := Hasher(bytify(key)) % uint64(len(newBuckets))
			newBuckets[hash] = append(newBuckets[hash], Bucket[K, V]{key, value})
		}
	}

	lhm.buckets = newBuckets
	return
}

func (lhm *LinkedHashMap[K, V]) Insert(key K, value V) error {
	if lhm.items == 0 || int(lhm.items) > 3*int(len(lhm.buckets))/4 {
		lhm.Resize()
	}

	bucketIndex, err := lhm.Bucket(key)
	if err != nil {
		return err
	}

	bucket := lhm.buckets[bucketIndex]
	for _, entry := range bucket {
		if reflect.DeepEqual(entry.key, key) {
			entry.key = key
			return nil
		}
	}

	lhm.items++
	bucket = append(bucket, Bucket[K, V]{key, value})

	return nil
}

func (lhm *LinkedHashMap[K, V]) Get(key K) (interface{}, error) {
	bucketIndex, err := lhm.Bucket(key)
	if err != nil {
		return nil, err
	}

	bucket := lhm.buckets[bucketIndex]
	for _, entry := range bucket {
		if entry.key == key {
			return entry.value, nil
		}
	}

	return nil, errors.New("invalid key, no associated buckets found")
}

func (lhm *LinkedHashMap[K, V]) Remove(key K) error {
	bucketIndex, err := lhm.Bucket(key)
	if err != nil {
		return err
	}

	bucket := lhm.buckets[bucketIndex]
	for i := 0; i < len(bucket); i++ {
		if reflect.DeepEqual(bucket[i].key, key) {
			bucket[i] = bucket[len(bucket)-1]
			lhm.buckets[bucketIndex] = bucket[:len(bucket)-1]
			return nil
		}
	}
	return errors.New("invalid key, no associated buckets found")
}

func (lhm *LinkedHashMap[K, V]) Bucket(key K) (uint64, error) {
	if lhm.items == 0 {
		return 0, errors.New("bucket is empty")
	}

	bucket := Hasher(bytify(key)) % uint64(len(lhm.buckets))

	if bucket >= uint64(len(lhm.buckets)) {
		return 0, errors.New("bucket does not exist")
	}
	return bucket, nil
}

type Server[K comparable, V any] struct {
	table *LinkedHashMap[K, V]
}

// doesn't cover all types*
func bytify[T any](data T) []byte {
	value := reflect.ValueOf(data)

	switch value.Kind() {
	case reflect.String:
		str := *(*string)(unsafe.Pointer(&value))
		return []byte(str)
	case reflect.Int:
		valueInt := value.Int()
		return []byte(strconv.FormatInt(valueInt, 10))
	case reflect.Uint:
		valueInt := value.Uint()
		return []byte(strconv.FormatUint(valueInt, 10))
	case reflect.Float64:
		valueFloat := value.Float()
		return []byte(strconv.FormatFloat(valueFloat, 'E', -1, 64))
	case reflect.Float32:
		valueFloat := value.Float()
		return []byte(strconv.FormatFloat(valueFloat, 'E', -1, 32))
	default:
		return nil
	}
}

func Hasher(data []byte) uint64 {
	var hash uint64
	for _, b := range data {
		hash = uint64(b) + (hash << 6) + (hash << 16) - hash
	}
	return hash
}
