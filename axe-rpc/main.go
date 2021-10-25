// main func
//@author: baoqiang
//@time: 2021/10/24 23:42:20
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/xiaoaxe/7days-golang/axe-rpc/axerpc"
	"github.com/xiaoaxe/7days-golang/axe-rpc/axerpc/registry"
	"github.com/xiaoaxe/7days-golang/axe-rpc/axerpc/xclient"
)

// server impl
type Foo int

type Args struct {
	Num1 int
	Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

// register run
func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	// listen http
	registry.HandleHTTP()
	wg.Done()

	_ = http.Serve(l, nil)
}

// server run
func startServer(registryAddr string, wg *sync.WaitGroup) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := axerpc.NewServer()

	// add server impl method
	_ = server.Register(&foo)

	// add self to registry
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)

	// server started
	wg.Done()

	// waiting for connect
	server.Accept(l)
}

// client run
func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var (
		reply int
		err   error
	)
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}

	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		// call ok
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(register string) {
	d := xclient.NewAxeRegistryDiscovery(register, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() {
		_ = xc.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i + 1})
		}(i)
	}

	// wait for complete
	wg.Wait()
}

func broadcast(register string) {
	d := xclient.NewAxeRegistryDiscovery(register, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() {
		_ = xc.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i + 1})
			// expect 2-5 timeout
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i + 1})
		}(i)
	}

	// wait for complete
	wg.Wait()
}

// debug: http://localhost:52988/debug/axerpc/
func main() {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_axerpc_/registry"

	var wg sync.WaitGroup

	// registry
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	// server
	time.Sleep(time.Second)
	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	// client
	time.Sleep(time.Second)
	call(registryAddr)
	broadcast(registryAddr)

	// viewing debug
	// time.Sleep(time.Minute)
}
