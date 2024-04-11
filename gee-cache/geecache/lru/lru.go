package lru

import "container/list"

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List               // 双向链表
	cache    map[string]*list.Element // 键是字符串，值是双向链表中对应节点的指针
	// 回调函数
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// 方便实例化Cache
// New is the constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 查找功能
// Get look ups a key's value
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) // 如果ele存在，则将该节点移动到队尾
		//(双向链表作为队列，队首队尾是相对的，此处约定front为队尾)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 删除，淘汰最近最少访问的节点(队首)
// RemoveOldest removes the oldest item
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele) // 取队首节点，从链表中删除
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)                                // 从字典中c.cache删除该节点的映射关系
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 更新所用内存
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) // 若回调函数不为nil，则调用回调函数
		}
	}
}

// 新增/修改
// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // 如果key存在，则更新对应节点的值，并将该节点移到队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else { // 不存在则新增节点，并在字典中添加key和节点的映射关系
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 更新c.nbytes，如果超过了设定的最大值c.maxBytes，则移除最少访问的节点
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 获取添加了多少条数据，便于测试
func (c *Cache) Len() int {
	return c.ll.Len()
}
