//http server&client
//@author: baoqiang
//@time: 2021/10/17 23:32:44
package axecache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	pb "github.com/xiaoaxe/7days-golang/axe-cache/axecache/axecachepb"
	"github.com/xiaoaxe/7days-golang/axe-cache/axecache/consistenthash"
)

const (
	defaultBasePath = "/_axecache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// run http server
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + req.URL.Path)
	}

	p.Log("Method=%v, Path=%v", req.Method, req.URL.Path)
	parts := strings.SplitN(req.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// get from group
	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		log.Printf("err get key: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// resp value
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		log.Printf("err marshal key: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get ok
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// add peers
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)

	// set httpGetter
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %v", peer)
		return p.httpGetters[peer], true
	}

	return nil, false
}

// HTTPPool impl PeerPicker
var _ PeerPicker = (*HTTPPool)(nil)

// http getter
type httpGetter struct {
	baseURL string
}

// get from brother node
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))

	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// parse response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", resp.Status)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

// httpGetter impl PeerGetter
var _ PeerGetter = (*httpGetter)(nil)
