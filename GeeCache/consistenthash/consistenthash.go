package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	hash     Hash
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环，Sorted
	hashMap  map[int]string // key为虚拟节点哈希值，value为真实节点名称
}

// 采用依赖注入的方式，Hash函数允许自定义
// 允许自定义虚拟节点倍数与Hash函数
// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ { // 每一个真实节点key，对应创建m.replicas个虚拟节点
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) // 虚拟节点名称为 strconv.Itoa(i) + key，即通过添加编号的方式区分不同虚拟节点
			m.keys = append(m.keys, hash)                      // 使用m.hash()计算虚拟节点的哈希值，使用append(m.keys, hash)添加到环上
			m.hashMap[hash] = key                              // 在hashMap中增加虚拟节点与真实节点的映射关系
		}
	}
	sort.Ints(m.keys) // 将环上的哈希值排序
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
