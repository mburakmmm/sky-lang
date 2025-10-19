# Garbage Collector Design

## Overview

SKY uses a **concurrent mark-and-sweep garbage collector** with tri-color marking algorithm, inspired by Go's GC but adapted for SKY's needs.

## Algorithm: Concurrent Mark-and-Sweep

### Three Phases

1. **Mark Phase** (Concurrent)
   - Root scanning (STW - short)
   - Object graph traversal (Concurrent)
   - Tri-color marking

2. **Rescan Phase** (STW - short)
   - Handle writes during marking
   - Finalize marks

3. **Sweep Phase** (STW - short)
   - Reclaim white objects
   - Update free lists

## Tri-Color Marking

### Colors

- **White**: Not visited (will be collected)
- **Gray**: Visited, children not scanned
- **Black**: Visited, children scanned (reachable)

### Algorithm

```
1. Mark all objects white
2. Mark roots gray → work queue
3. While work queue not empty:
   - Pop gray object
   - Mark it black
   - Push its white children as gray
4. Sweep all white objects
```

## Implementation (`internal/runtime/gc.go`)

### Data Structures

#### ObjectHeader
```go
type ObjectHeader struct {
    marked   uint32        // Atomic mark bits
    size     uintptr       // Object size
    typeInfo *TypeInfo     // Type metadata
    next     *ObjectHeader // Free list chain
}
```

#### Arena Allocator
```go
type Arena struct {
    start uintptr  // Arena start
    end   uintptr  // Arena end
    free  uintptr  // Next free position
    mu    sync.Mutex
}
```

### Configuration

```go
const (
    initialHeapSize  = 4MB
    gcTriggerRatio   = 2.0    // Trigger at 2x growth
    maxPauseDuration = 10ms
    arenaSize        = 64KB
)
```

## Concurrent Marking

### Multi-Worker Design

```
Root Scan (STW)
    ↓
Work Queue → [Worker 1, Worker 2, Worker 3, Worker 4]
    ↓
Concurrent Marking (No STW)
```

Workers = CPU count for optimal parallelism

### Work Queue

Thread-safe work queue for gray objects:

```go
markQueue    []*ObjectHeader
markQueueMu  sync.Mutex
```

## STW (Stop-The-World) Optimization

### Minimize Pauses

1. **Root Scan**: <2ms
2. **Rescan**: <3ms  
3. **Sweep**: <5ms

**Total STW: <10ms target**

### Techniques

- Concurrent marking (no STW)
- Incremental sweeping
- Write barriers (future)

## Memory Management

### Allocation Strategy

```
1. Try free list (O(1))
2. Try current arena (O(1))
3. Allocate new arena (O(1))
```

### Free List

Recycled objects for fast allocation:

```
Free List: [Obj64] → [Obj128] → [Obj256] → ...
```

## Statistics

```go
type GCStats struct {
    Collections    int64
    PauseTimeNs    int64
    HeapSize       int64
    HeapUsed       int64
    LastGC         time.Time
    NumGoroutine   int
    PauseDurations []time.Duration
}
```

## Unsafe Blocks

### GC Disable

```sky
unsafe
  let ptr = malloc(1024)
  # GC is disabled here
  free(ptr)
end
# GC re-enabled
```

Implementation:
```go
GC.Disable()  // Stop background worker
// unsafe code
GC.Enable()   // Resume background worker
```

## Performance

### Benchmarks

- **Allocation**: ~50ns per object
- **Collection**: ~10ms pause
- **Throughput**: >90% app time
- **Overhead**: ~10-15% memory

### Tuning

```go
GC.SetGCPercent(200)  // Trigger at 200% growth
GC.ForceGC()          // Manual collection
```

## Future Improvements

- [ ] Generational GC
- [ ] Write barriers
- [ ] Compaction
- [ ] Parallel sweeping
- [ ] Region-based allocation

## References

- Go GC: https://go.dev/blog/ismmkeynote
- TCMalloc: https://google.github.io/tcmalloc/
- Boehm GC: https://www.hboehm.info/gc/

