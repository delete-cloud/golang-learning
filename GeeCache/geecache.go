package geecache

import (
	"fmt"
	"log"
	"sync"
)

// 接口
// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// 函数类型
// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// 接口型函数的实现
// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 实例化Group
// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,   // 唯一名称
		getter:    getter, // 缓存未命中时获取源数据的回调
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// 获取特定名称的Group
// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock() // RLock() 只读锁，此处不涉及任何冲突变量的写操作
	g := groups[name]
	mu.RUnlock()
	return g
}

// 核心方法Get
// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// 从mainCache中查找缓存，如果存在则返回缓存值
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	// 缓存不存在，调用load，load调用getLocally
	// getLocally调用用户回调函数g.getter.Get()获取源数据
	// 并将源数据添加到缓存mainCache中(通过populateCache方法)
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
