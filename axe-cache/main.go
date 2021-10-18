//main func
//@author: baoqiang
//@time: 2021/10/17 22:50:06
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/xiaoaxe/7days-golang/axe-cache/axecache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *axecache.Group {
	return axecache.NewGroup("scores", 2<<10, axecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Printf("[SlowDB] search key: %v", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not found", key)
		},
	))
}

func startCacheServer(addr string, addrs []string, g *axecache.Group) {
	peers := axecache.NewHTTPPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Printf("axecache is running at %v\n", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers)) // trim "http://"
}

func startAPIServer(apiAddr string, g *axecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			key := req.URL.Query().Get("key")
			view, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// write value
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		},
	))

	log.Printf("fontend server is running at %v\n", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil)) // trim "http://"
}

// ./c.out --api true
// curl http://localhost:8001/_axecache/scores/Tom
// curl http://localhost:9999/api?key=Tom
func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "AxeCache server port")
	flag.BoolVar(&api, "api", false, "start a api server?")
	flag.Parse()

	// addrs
	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, addr := range addrMap {
		addrs = append(addrs, addr)
	}

	// start server
	g := createGroup()
	if api {
		go startAPIServer(apiAddr, g)
	}
	startCacheServer(addrMap[port], addrs, g)
}
