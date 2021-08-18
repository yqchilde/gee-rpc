package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect     SelectMode = iota // 随机选择
	RoundRobinSelect                   // 基于round robin的轮询选择
)

type Discovery interface {
	// Refresh 从注册中心更新服务列表
	Refresh() error

	// Update 手动更新服务列表
	Update(servers []string) error

	// Get 根据负载均衡策略，选择一个服务实例
	Get(mode SelectMode) (string, error)

	// GetAll 返回所有的服务实例
	GetAll() ([]string, error)
}

// MultiServersDiscovery 是对没有注册中心的多服务器发现
// 用户需提供明确可寻址的服务器地址
type MultiServersDiscovery struct {
	r       *rand.Rand // 生成随机数
	mu      sync.Mutex // protect following
	servers []string   // 存放多个server
	index   int        // 记录robin算法的选择位置
}

// NewMultiServerDiscovery ...
func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		servers: servers,
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MultiServersDiscovery)(nil)

// Refresh doesn't make sense for MultiServersDiscovery, so ignore it
func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

// Update the servers of discovery dynamically if needed
func (d *MultiServersDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

// Get a server according to me
func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		// servers could be updated, so mode n to ensure safety
		s := d.servers[d.index%n]
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

// GetAll all servers in discovery
func (d *MultiServersDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// return a copy of d.servers
	servers := make([]string, len(d.servers), len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
