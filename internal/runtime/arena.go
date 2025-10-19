package runtime

import (
	"sync"
	"unsafe"
)

// ArenaAllocator manages a pool of memory for fast allocation
type ArenaAllocator struct {
	blocks    [][]byte
	blockSize int
	current   int
	offset    int
	mu        sync.Mutex
}

// NewArenaAllocator creates a new arena allocator
func NewArenaAllocator(blockSize int) *ArenaAllocator {
	return &ArenaAllocator{
		blocks:    make([][]byte, 0),
		blockSize: blockSize,
		current:   -1,
		offset:    0,
	}
}

// Alloc allocates memory from the arena
func (a *ArenaAllocator) Alloc(size int) unsafe.Pointer {
	a.mu.Lock()
	defer a.mu.Unlock()

	// If size is too large or no current block, allocate new block
	if a.current == -1 || a.offset+size > a.blockSize {
		newBlock := make([]byte, a.blockSize)
		a.blocks = append(a.blocks, newBlock)
		a.current = len(a.blocks) - 1
		a.offset = 0
	}

	// Allocate from current block
	block := a.blocks[a.current]
	ptr := unsafe.Pointer(&block[a.offset])
	a.offset += size

	return ptr
}

// Reset resets the arena (marks all memory as free)
func (a *ArenaAllocator) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.current = 0
	a.offset = 0
	// Keep blocks allocated for reuse
}

// Free frees all arena memory
func (a *ArenaAllocator) Free() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.blocks = nil
	a.current = -1
	a.offset = 0
}

// Size returns total allocated size
func (a *ArenaAllocator) Size() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	return len(a.blocks) * a.blockSize
}
