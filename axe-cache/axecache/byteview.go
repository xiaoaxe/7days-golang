//the readonly bytes
//@author: baoqiang
//@time: 2021/10/17 22:54:01
package axecache

type ByteView struct {
	b []byte
}

// impl Value interface
func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
