// xclient
//@author: baoqiang
//@time: 2021/10/22 23:11:17
package xclient

import (
	"context"
	"io"
	"reflect"
	"sync"

	. "github.com/xiaoaxe/7days-golang/axe-rpc/axerpc"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *Option
	mu      sync.Mutex
	clients map[string]*Client
}

// new
var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		mu:      sync.Mutex{},
		clients: map[string]*Client{},
	}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()

	for key, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}

func (xc *XClient) dial(rpcAddr string) (*Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	c, ok := xc.clients[rpcAddr]
	// exists but not available
	if ok && !c.IsAvailable() {
		_ = c.Close()
		delete(xc.clients, rpcAddr)
		c = nil
	}

	if c == nil {
		var err error
		c, err = XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		// set map val
		xc.clients[rpcAddr] = c
	}
	return c, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	c, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	// real call with client
	return c.Call(ctx, serviceMethod, args, reply)
}

// exported Call & Broadcast func
// get rpcAddr from discovery
func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		e         error
		replyDone = reply == nil // nil reply do not need set val
	)
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			// real call
			err := xc.call(addr, ctx, serviceMethod, args, clonedReply)

			// set value
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel() // if any failed, then cancel ALL
			}

			// set reply once only if reply is not nil
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()

		}(rpcAddr)
	}

	// wait for complete
	wg.Wait()
	return e
}
