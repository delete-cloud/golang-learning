package singleflight

import "sync"

type call struct { // call代表正在进行中或已经结束的请求
	wg  sync.WaitGroup // 锁，避免重入
	val interface{}
	err error
}

type Group struct { // 管理不同key的请求(call)
	mu sync.Mutex // protects m
	m  map[string]*call
}

// Do方法 针对相同的key，无论Do被调用多少次，函数fn都只被调用一次，等待fn调用结束后，返回返回值或错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock() // g.mu为保护Group成员变量m不被并发读写而加的锁
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // 如果请求正在进行中，则等待 阻塞，直到锁被释放
		return c.val, c.err // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)  // 发起请求前加锁，锁加1
	g.m[key] = c // 添加到g.m，表明key已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() // 调用fn，发起请求
	c.wg.Done()         // 请求结束，锁减1

	g.mu.Lock()
	delete(g.m, key) // 更新g.m
	g.mu.Unlock()

	return c.val, c.err // 返回结果
}
