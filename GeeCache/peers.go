package geecache

// PeerPicker is the interface that must be implemented to locate
// the peer than owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // 根据传入的key选择相应节点PeerGetter
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(group string, key string) ([]byte, error) // 从对应group中查找缓存值
}
