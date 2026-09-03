package nodemanager

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Kynetic-Engynes-Platforms/amara/pkg/impls/ops"
	"github.com/ccoveille/go-safecast/v2"
)

type Node struct {
	URL         string
	IsHealthy   atomic.Bool
	LastFailure atomic.Int64 // Stores UnixNano
}

type nodeManager struct {
	nodes          []*Node
	idx            atomic.Uint32
	healthWaitTime time.Duration
}

func NewNodeManager(urls []string, healthWaitTime time.Duration) (ops.NodeManager, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("at least one node URL must be provided")
	}

	nodes := make([]*Node, len(urls))
	for i, u := range urls {
		n := &Node{URL: u}
		n.IsHealthy.Store(true)
		nodes[i] = n
	}

	return &nodeManager{
		nodes:          nodes,
		healthWaitTime: healthWaitTime,
	}, nil
}

func (nm *nodeManager) GetNextNode() string {
	now := time.Now().UnixNano()
	total, _ := safecast.Convert[uint32](len(nm.nodes))

	for range total {
		// Atomically increment and get the current index
		currentIndex := nm.idx.Add(1) % total
		node := nm.nodes[currentIndex]

		if node.IsHealthy.Load() {
			return node.URL
		}

		// Check if health timeout has expired
		lastFailure := node.LastFailure.Load()
		if time.Duration(now-lastFailure) > nm.healthWaitTime {
			node.IsHealthy.Store(true)
			return node.URL
		}
	}

	// Fallback if all are unhealthy
	return nm.nodes[nm.idx.Add(1)%total].URL
}

func (nm *nodeManager) MarkUnhealthy(rawURL string) {
	for _, n := range nm.nodes {
		if n.URL == rawURL {
			n.IsHealthy.Store(false)
			n.LastFailure.Store(time.Now().UnixNano())
			break
		}
	}
}

func (nm *nodeManager) MarkHealthy(rawURL string) {
	for _, n := range nm.nodes {
		if n.URL == rawURL {
			n.IsHealthy.Store(true)
			break
		}
	}
}
