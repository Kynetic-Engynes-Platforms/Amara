package sdk

import (
	"fmt"
	"sync"
	"time"
)

type Node struct {
	URL         string
	IsHealthy   bool
	LastFailure time.Time
}

type nodeManager struct {
	nodes          []*Node
	mu             sync.RWMutex
	idx            int
	healthWaitTime time.Duration
}

// Constructor returns the NodeManager interface
func NewNodeManager(urls []string, healthWaitTime time.Duration) (NodeManager, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("at least one node URL must be provided")
	}

	nodes := make([]*Node, len(urls))
	for i, u := range urls {
		nodes[i] = &Node{URL: u, IsHealthy: true}
	}

	return &nodeManager{
		nodes:          nodes,
		healthWaitTime: healthWaitTime,
	}, nil
}

// GetNextNode implements round-robin node selection with health state checks and recovery windows.
func (nm *nodeManager) GetNextNode() string {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	now := time.Now()
	total := len(nm.nodes)

	// First pass: locate the next healthy node (or one whose cooldown has expired)
	for i := 0; i < total; i++ {
		nm.idx = (nm.idx + 1) % total
		node := nm.nodes[nm.idx]

		if !node.IsHealthy && now.Sub(node.LastFailure) > nm.healthWaitTime {
			node.IsHealthy = true
		}

		if node.IsHealthy {
			return node.URL
		}
	}

	// Fallback: If all nodes are marked unhealthy, return current round-robin candidate
	nm.idx = (nm.idx + 1) % total
	return nm.nodes[nm.idx].URL
}

func (nm *nodeManager) MarkUnhealthy(rawURL string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	for _, n := range nm.nodes {
		if n.URL == rawURL {
			n.IsHealthy = false
			n.LastFailure = time.Now()
			break
		}
	}
}

func (nm *nodeManager) MarkHealthy(rawURL string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	for _, n := range nm.nodes {
		if n.URL == rawURL {
			n.IsHealthy = true
			break
		}
	}
}
