//peers impl
//@author: baoqiang
//@time: 2021/10/17 22:54:35
package axecache

import pb "github.com/xiaoaxe/7days-golang/axe-cache/axecache/axecachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
