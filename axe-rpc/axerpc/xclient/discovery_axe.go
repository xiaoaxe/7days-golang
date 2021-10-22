// discory using registry
//@author: baoqiang
//@time: 2021/10/22 23:33:00
package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type AxeRegistryDiscovery struct {
	*MultiServersDiscovery
	registry   string
	timeout    time.Duration
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second * 10

// new func
func NewAxeRegistryDiscovery(registryAddr string, timeout time.Duration) *AxeRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &AxeRegistryDiscovery{
		MultiServersDiscovery: NewMultiServersDiscovery(make([]string, 0)),
		registry:              registryAddr,
		timeout:               timeout,
	}
	return d
}

// impl
func (d *AxeRegistryDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *AxeRegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// no need for refresh
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}

	log.Println("rpc registry: refresh servers from registry: ", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("rpc registry refresh err: ", err)
		return err
	}

	servers := strings.Split(resp.Header.Get("X-Axerpc-Servers"), ",")
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}

	// update lastUpdateTime
	d.lastUpdate = time.Now()
	return nil
}

func (d *AxeRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	// get from servers
	return d.MultiServersDiscovery.Get(mode)
}

func (d *AxeRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	// get from servers
	return d.MultiServersDiscovery.GetAll()
}
