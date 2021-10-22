// registry
//@author: baoqiang
//@time: 2021/10/22 22:48:20
package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type AxeRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_axerpc_/registry"
	defaultTimeout = time.Minute * 5
)

// new registry
func New(timeout time.Duration) *AxeRegistry {
	return &AxeRegistry{
		servers: map[string]*ServerItem{},
		timeout: timeout,
	}
}

var DefaultAxeRegistry = New(defaultTimeout)

// set & get
func (r *AxeRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		s.start = time.Now()
	}
}

func (r *AxeRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var alives []string

	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alives = append(alives, addr)
		} else {
			delete(r.servers, addr)
		}
	}

	// sort addrs
	sort.Strings(alives)
	return alives
}

// http service
func (r *AxeRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-Axerpc-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		addr := req.Header.Get("X-Axerpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *AxeRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path: ", registryPath)
}

func HandleHTTP() {
	DefaultAxeRegistry.HandleHTTP(defaultPath)
}

// is the registry alive?
func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry: ", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-Axerpc-Server", addr)

	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err: ", err)
		return err
	}
	return nil
}
