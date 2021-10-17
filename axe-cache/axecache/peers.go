//peers impl
//@author: baoqiang
//@time: 2021/10/17 22:54:35
package axecache

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// TODO fixme, replace me with pb req&resp
type PeerGetter interface {
	Get(in, out []byte) error
}
