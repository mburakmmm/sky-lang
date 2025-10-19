package runtime

import (
	"testing"
	"time"
	"unsafe"
)

func TestGCAlloc(t *testing.T) {
	gc := NewGC()
	defer gc.Stop()

	typeInfo := &TypeInfo{
		name:        "TestObject",
		size:        64,
		pointerMask: []byte{},
	}

	ptr := gc.Alloc(64, typeInfo)
	if ptr == nil {
		t.Fatal("Alloc returned nil")
	}
}

func TestGCMultipleAllocs(t *testing.T) {
	gc := NewGC()
	defer gc.Stop()

	typeInfo := &TypeInfo{
		name:        "TestObject",
		size:        32,
		pointerMask: []byte{},
	}

	// Allocate 100 objects
	pointers := make([]unsafe.Pointer, 100)
	for i := 0; i < 100; i++ {
		pointers[i] = gc.Alloc(32, typeInfo)
		if pointers[i] == nil {
			t.Fatalf("Alloc %d failed", i)
		}
	}

	// All pointers should be unique
	seen := make(map[unsafe.Pointer]bool)
	for _, ptr := range pointers {
		if seen[ptr] {
			t.Error("duplicate pointer allocated")
		}
		seen[ptr] = true
	}
}

func TestGCCollect(t *testing.T) {
	gc := NewGC()
	defer gc.Stop()

	typeInfo := &TypeInfo{
		name:        "TestObject",
		size:        64,
		pointerMask: []byte{},
	}

	// Allocate and add as root
	ptr := gc.Alloc(64, typeInfo)
	gc.AddRoot(ptr)

	initialHeap := gc.Stats().HeapUsed

	// Allocate more (not rooted)
	for i := 0; i < 100; i++ {
		gc.Alloc(64, typeInfo)
	}

	// Force GC
	gc.ForceGC()
	time.Sleep(50 * time.Millisecond) // Wait for concurrent marking

	// Heap should have decreased (unreachable objects collected)
	stats := gc.Stats()
	if stats.Collections == 0 {
		t.Error("GC should have run at least once")
	}

	t.Logf("Initial heap: %d, After GC: %d, Collections: %d",
		initialHeap, stats.HeapUsed, stats.Collections)
}

func TestGCEnableDisable(t *testing.T) {
	gc := NewGC()
	defer gc.Stop()

	if !gc.IsEnabled() {
		t.Error("GC should be enabled by default")
	}

	gc.Disable()
	if gc.IsEnabled() {
		t.Error("GC should be disabled")
	}

	gc.Enable()
	if !gc.IsEnabled() {
		t.Error("GC should be enabled")
	}
}

func TestGCStats(t *testing.T) {
	gc := NewGC()
	defer gc.Stop()

	stats := gc.Stats()

	if stats.HeapSize <= 0 {
		t.Error("HeapSize should be positive")
	}

	// Allocate something
	typeInfo := &TypeInfo{name: "Test", size: 100}
	gc.Alloc(100, typeInfo)

	newStats := gc.Stats()
	if newStats.HeapUsed <= stats.HeapUsed {
		t.Error("HeapUsed should have increased")
	}
}

func TestGCRootManagement(t *testing.T) {
	gc := NewGC()
	defer gc.Stop()

	typeInfo := &TypeInfo{name: "Root", size: 64}
	ptr := gc.Alloc(64, typeInfo)

	// Add root
	gc.AddRoot(ptr)

	// Remove root
	gc.RemoveRoot(ptr)

	// Should not crash
}

func TestArenaAlloc(t *testing.T) {
	arena := &Arena{
		start: 1000,
		end:   2000,
		free:  1000,
	}

	ptr := arena.tryAlloc(64)
	if ptr == 0 {
		t.Fatal("arena alloc failed")
	}

	if ptr != 1000 {
		t.Errorf("expected ptr at 1000, got %d", ptr)
	}

	if arena.free != 1064 {
		t.Errorf("expected free at 1064, got %d", arena.free)
	}
}

func TestArenaFull(t *testing.T) {
	arena := &Arena{
		start: 1000,
		end:   1100,
		free:  1000,
	}

	// Should succeed
	ptr1 := arena.tryAlloc(50)
	if ptr1 == 0 {
		t.Error("first alloc should succeed")
	}

	// Should succeed (aligned)
	ptr2 := arena.tryAlloc(40)
	if ptr2 == 0 {
		t.Error("second alloc should succeed")
	}

	// Should fail (arena full)
	ptr3 := arena.tryAlloc(50)
	if ptr3 != 0 {
		t.Error("third alloc should fail (arena full)")
	}
}
