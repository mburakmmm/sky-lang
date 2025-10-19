package runtime

import (
	"context"
	"sync"
	"time"
)

// CancellationToken represents a cancellation token
type CancellationToken struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

// NewCancellationToken creates a new cancellation token
func NewCancellationToken() *CancellationToken {
	ctx, cancel := context.WithCancel(context.Background())
	return &CancellationToken{
		ctx:    ctx,
		cancel: cancel,
	}
}

// NewCancellationTokenWithTimeout creates a token with timeout
func NewCancellationTokenWithTimeout(timeout time.Duration) *CancellationToken {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return &CancellationToken{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Cancel cancels the token
func (ct *CancellationToken) Cancel() {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.cancel()
}

// IsCancelled checks if the token is cancelled
func (ct *CancellationToken) IsCancelled() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	select {
	case <-ct.ctx.Done():
		return true
	default:
		return false
	}
}

// Wait waits for cancellation
func (ct *CancellationToken) Wait() <-chan struct{} {
	return ct.ctx.Done()
}

// TaskTree manages hierarchical task cancellation
type TaskTree struct {
	root    *TaskNode
	mu      sync.RWMutex
	nodeMap map[string]*TaskNode
}

// TaskNode represents a node in the task tree
type TaskNode struct {
	id       string
	token    *CancellationToken
	children []*TaskNode
	parent   *TaskNode
}

// NewTaskTree creates a new task tree
func NewTaskTree() *TaskTree {
	rootToken := NewCancellationToken()
	root := &TaskNode{
		id:       "root",
		token:    rootToken,
		children: []*TaskNode{},
		parent:   nil,
	}

	return &TaskTree{
		root:    root,
		nodeMap: map[string]*TaskNode{"root": root},
	}
}

// AddTask adds a task to the tree
func (tt *TaskTree) AddTask(id, parentID string) *CancellationToken {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	parent, exists := tt.nodeMap[parentID]
	if !exists {
		parent = tt.root
	}

	token := NewCancellationToken()
	node := &TaskNode{
		id:       id,
		token:    token,
		children: []*TaskNode{},
		parent:   parent,
	}

	parent.children = append(parent.children, node)
	tt.nodeMap[id] = node

	return token
}

// CancelTask cancels a task and all its children
func (tt *TaskTree) CancelTask(id string) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	node, exists := tt.nodeMap[id]
	if !exists {
		return
	}

	tt.cancelNode(node)
}

func (tt *TaskTree) cancelNode(node *TaskNode) {
	node.token.Cancel()

	for _, child := range node.children {
		tt.cancelNode(child)
	}
}

// GetToken gets the cancellation token for a task
func (tt *TaskTree) GetToken(id string) *CancellationToken {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	if node, exists := tt.nodeMap[id]; exists {
		return node.token
	}
	return nil
}
