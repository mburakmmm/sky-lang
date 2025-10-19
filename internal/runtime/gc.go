package runtime

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// ObjectHeader her GC objesi için header
type ObjectHeader struct {
	marked   uint32        // Mark biti (atomic)
	size     uintptr       // Obje boyutu
	typeInfo *TypeInfo     // Tip bilgisi
	next     *ObjectHeader // Free list için
}

// TypeInfo tip metadata'sı
type TypeInfo struct {
	name        string
	size        uintptr
	pointerMask []byte // Hangi alanların pointer olduğu
}

// GCStats GC istatistiklerini tutar
type GCStats struct {
	Collections    int64           // Toplam collection sayısı
	PauseTimeNs    int64           // Toplam pause süresi (nanosaniye)
	HeapSize       int64           // Heap boyutu
	HeapUsed       int64           // Kullanılan heap
	LastGC         time.Time       // Son GC zamanı
	NextGC         int64           // Sonraki GC threshold
	NumGoroutine   int             // Aktif goroutine sayısı
	PauseDurations []time.Duration // Son pause süreleri
}

const (
	// GC thresholds
	initialHeapSize  = 4 * 1024 * 1024       // 4MB başlangıç
	gcTriggerRatio   = 2.0                   // %200 büyüme ile trigger
	maxPauseDuration = 10 * time.Millisecond // Max STW pause

	// Arena settings
	arenaSize = 64 * 1024 // 64KB arena

	// Color marks (tri-color marking)
	colorWhite = 0
	colorGray  = 1
	colorBlack = 2
)

// Arena memory arena (basic implementation for GC)
// Note: Enhanced ArenaAllocator available in arena.go
type Arena struct {
	start uintptr
	end   uintptr
	free  uintptr
	mu    sync.Mutex
}

// GC global garbage collector instance
var GC *GarbageCollector

// GarbageCollector concurrent mark-and-sweep GC
type GarbageCollector struct {
	enabled    atomic.Bool
	collecting atomic.Bool
	stats      GCStats
	statsMu    sync.RWMutex

	// Heap management
	arenas   []*Arena
	arenasMu sync.RWMutex

	// Object tracking
	allObjects []*ObjectHeader
	objectsMu  sync.RWMutex
	freeList   *ObjectHeader
	freeListMu sync.Mutex

	// Root set (stack roots, global variables)
	roots   []unsafe.Pointer
	rootsMu sync.RWMutex

	// Work queue for concurrent marking
	markQueue     []*ObjectHeader
	markQueueMu   sync.Mutex
	markWorkersWg sync.WaitGroup

	// GC trigger
	triggerSize int64

	// Background goroutine
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewGC yeni bir GC oluşturur
func NewGC() *GarbageCollector {
	gc := &GarbageCollector{
		arenas:      make([]*Arena, 0, 16),
		allObjects:  make([]*ObjectHeader, 0, 1024),
		roots:       make([]unsafe.Pointer, 0, 256),
		markQueue:   make([]*ObjectHeader, 0, 256),
		triggerSize: initialHeapSize,
		stopCh:      make(chan struct{}),
	}

	gc.enabled.Store(true)

	// İlk arena'yı oluştur
	gc.newArena()

	// Background GC goroutine başlat
	gc.wg.Add(1)
	go gc.backgroundWorker()

	return gc
}

// Enable GC'yi aktif eder
func (gc *GarbageCollector) Enable() {
	gc.enabled.Store(true)
}

// Disable GC'yi devre dışı bırakır (unsafe bloklar için)
func (gc *GarbageCollector) Disable() {
	gc.enabled.Store(false)
}

// IsEnabled GC'nin aktif olup olmadığını döndürür
func (gc *GarbageCollector) IsEnabled() bool {
	return gc.enabled.Load()
}

// Alloc bellek ayırır
func (gc *GarbageCollector) Alloc(size uintptr, typeInfo *TypeInfo) unsafe.Pointer {
	// Önce free list'ten bak
	gc.freeListMu.Lock()
	if gc.freeList != nil && gc.freeList.size >= size {
		obj := gc.freeList
		gc.freeList = obj.next
		gc.freeListMu.Unlock()

		// Reset header
		atomic.StoreUint32(&obj.marked, colorWhite)
		obj.typeInfo = typeInfo

		return unsafe.Pointer(uintptr(unsafe.Pointer(obj)) + unsafe.Sizeof(ObjectHeader{}))
	}
	gc.freeListMu.Unlock()

	// Yeni obje allocate et
	totalSize := unsafe.Sizeof(ObjectHeader{}) + size

	// Arena'dan allocate et
	gc.arenasMu.RLock()
	for _, arena := range gc.arenas {
		if ptr := arena.tryAlloc(totalSize); ptr != 0 {
			gc.arenasMu.RUnlock()

			// Object header oluştur
			headerPtr := unsafe.Pointer(ptr)
			header := (*ObjectHeader)(headerPtr)
			header.marked = colorWhite
			header.size = size
			header.typeInfo = typeInfo
			header.next = nil

			// Track object
			gc.objectsMu.Lock()
			gc.allObjects = append(gc.allObjects, header)
			atomic.AddInt64(&gc.stats.HeapUsed, int64(totalSize))
			gc.objectsMu.Unlock()

			// GC trigger kontrolü
			if atomic.LoadInt64(&gc.stats.HeapUsed) > gc.triggerSize {
				go gc.Collect() // Async trigger
			}

			return unsafe.Pointer(uintptr(unsafe.Pointer(header)) + unsafe.Sizeof(ObjectHeader{}))
		}
	}
	gc.arenasMu.RUnlock()

	// Yeni arena gerekiyor
	gc.arenasMu.Lock()
	arena := gc.newArena()
	gc.arenasMu.Unlock()

	ptr := arena.tryAlloc(totalSize)
	if ptr == 0 {
		panic("GC: failed to allocate memory")
	}

	headerPtr := unsafe.Pointer(ptr)
	header := (*ObjectHeader)(headerPtr)
	header.marked = colorWhite
	header.size = size
	header.typeInfo = typeInfo
	header.next = nil

	gc.objectsMu.Lock()
	gc.allObjects = append(gc.allObjects, header)
	atomic.AddInt64(&gc.stats.HeapUsed, int64(totalSize))
	gc.objectsMu.Unlock()

	return unsafe.Pointer(uintptr(unsafe.Pointer(header)) + unsafe.Sizeof(ObjectHeader{}))
}

// AddRoot root set'e pointer ekler
func (gc *GarbageCollector) AddRoot(ptr unsafe.Pointer) {
	gc.rootsMu.Lock()
	defer gc.rootsMu.Unlock()
	gc.roots = append(gc.roots, ptr)
}

// RemoveRoot root set'ten pointer çıkarır
func (gc *GarbageCollector) RemoveRoot(ptr unsafe.Pointer) {
	gc.rootsMu.Lock()
	defer gc.rootsMu.Unlock()

	for i, root := range gc.roots {
		if root == ptr {
			gc.roots = append(gc.roots[:i], gc.roots[i+1:]...)
			break
		}
	}
}

// Collect tam GC cycle çalıştırır (concurrent mark + sweep)
func (gc *GarbageCollector) Collect() {
	if !gc.enabled.Load() {
		return
	}

	// Zaten bir collection çalışıyor mu?
	if !gc.collecting.CompareAndSwap(false, true) {
		return
	}
	defer gc.collecting.Store(false)

	startTime := time.Now()

	// Phase 1: STW - Mark roots (short pause)
	stw := gc.stopTheWorld()
	gc.markRoots()
	gc.startTheWorld(stw)

	// Phase 2: Concurrent marking (no STW)
	gc.concurrentMark()

	// Phase 3: STW - Rescan and sweep (short pause)
	stw = gc.stopTheWorld()
	gc.rescan()
	gc.sweep()
	gc.startTheWorld(stw)

	// Update stats
	pauseDuration := time.Since(startTime)
	gc.statsMu.Lock()
	gc.stats.Collections++
	gc.stats.PauseTimeNs += pauseDuration.Nanoseconds()
	gc.stats.LastGC = time.Now()
	if len(gc.stats.PauseDurations) >= 256 {
		gc.stats.PauseDurations = gc.stats.PauseDurations[1:]
	}
	gc.stats.PauseDurations = append(gc.stats.PauseDurations, pauseDuration)
	gc.statsMu.Unlock()

	// Sonraki GC threshold'u ayarla
	heapUsed := atomic.LoadInt64(&gc.stats.HeapUsed)
	gc.triggerSize = int64(float64(heapUsed) * gcTriggerRatio)
}

// markRoots root set'i işaretle
func (gc *GarbageCollector) markRoots() {
	gc.rootsMu.RLock()
	defer gc.rootsMu.RUnlock()

	for _, root := range gc.roots {
		if root != nil {
			gc.markObject(root)
		}
	}
}

// markObject bir objeyi işaretle
func (gc *GarbageCollector) markObject(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}

	// Header'ı bul
	ptrAddr := uintptr(ptr)
	headerAddr := ptrAddr - unsafe.Sizeof(ObjectHeader{})
	headerPtr := unsafe.Pointer(headerAddr)
	header := (*ObjectHeader)(headerPtr)

	// Zaten işaretlendi mi?
	if !atomic.CompareAndSwapUint32(&header.marked, colorWhite, colorGray) {
		return
	}

	// Work queue'ya ekle
	gc.markQueueMu.Lock()
	gc.markQueue = append(gc.markQueue, header)
	gc.markQueueMu.Unlock()
}

// concurrentMark concurrent marking phase
func (gc *GarbageCollector) concurrentMark() {
	// Mark worker'ları başlat
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		gc.markWorkersWg.Add(1)
		go gc.markWorker()
	}

	gc.markWorkersWg.Wait()
}

// markWorker marking işini yapan worker
func (gc *GarbageCollector) markWorker() {
	defer gc.markWorkersWg.Done()

	for {
		gc.markQueueMu.Lock()
		if len(gc.markQueue) == 0 {
			gc.markQueueMu.Unlock()
			return
		}

		obj := gc.markQueue[0]
		gc.markQueue = gc.markQueue[1:]
		gc.markQueueMu.Unlock()

		// Gray -> Black
		atomic.StoreUint32(&obj.marked, colorBlack)

		// Scan pointers
		if obj.typeInfo != nil && len(obj.typeInfo.pointerMask) > 0 {
			objAddr := uintptr(unsafe.Pointer(obj))
			objPtrAddr := objAddr + unsafe.Sizeof(ObjectHeader{})
			for i, hasPtrGC := range obj.typeInfo.pointerMask {
				if hasPtrGC != 0 {
					fieldAddr := objPtrAddr + uintptr(i)*unsafe.Sizeof(unsafe.Pointer(nil))
					fieldPtr := *(*unsafe.Pointer)(unsafe.Pointer(fieldAddr))
					if fieldPtr != nil {
						gc.markObject(fieldPtr)
					}
				}
			}
		}
	}
}

// rescan yazma bariyerlerinden kaçan objeleri tekrar tara
func (gc *GarbageCollector) rescan() {
	// Basit implementasyon: tüm gray objeleri tekrar işaretle
	// Production'da write barrier kullanılır
	gc.objectsMu.RLock()
	defer gc.objectsMu.RUnlock()

	for _, obj := range gc.allObjects {
		if atomic.LoadUint32(&obj.marked) == colorGray {
			atomic.StoreUint32(&obj.marked, colorBlack)
		}
	}
}

// sweep işaretlenmemiş objeleri temizle
func (gc *GarbageCollector) sweep() {
	gc.objectsMu.Lock()
	defer gc.objectsMu.Unlock()

	newObjects := make([]*ObjectHeader, 0, len(gc.allObjects))
	freedBytes := int64(0)

	for _, obj := range gc.allObjects {
		if atomic.LoadUint32(&obj.marked) == colorWhite {
			// White = unreachable, free it
			gc.freeListMu.Lock()
			obj.next = gc.freeList
			gc.freeList = obj
			gc.freeListMu.Unlock()

			freedBytes += int64(obj.size + uintptr(unsafe.Sizeof(ObjectHeader{})))
		} else {
			// Black = reachable, keep it and reset to white
			atomic.StoreUint32(&obj.marked, colorWhite)
			newObjects = append(newObjects, obj)
		}
	}

	gc.allObjects = newObjects
	atomic.AddInt64(&gc.stats.HeapUsed, -freedBytes)
}

// stopTheWorld tüm mutator'ları durdur
func (gc *GarbageCollector) stopTheWorld() time.Time {
	// Go runtime'ın STW'sini kullan
	runtime.GC() // Bu gerçekte çok basitleştirilmiş
	return time.Now()
}

// startTheWorld mutator'ları tekrar başlat
func (gc *GarbageCollector) startTheWorld(stw time.Time) {
	pauseDuration := time.Since(stw)
	if pauseDuration > maxPauseDuration {
		// Log warning: pause too long
	}
}

// backgroundWorker arka plan GC worker'ı
func (gc *GarbageCollector) backgroundWorker() {
	defer gc.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Periyodik olarak heap kullanımını kontrol et
			if atomic.LoadInt64(&gc.stats.HeapUsed) > gc.triggerSize {
				gc.Collect()
			}
		case <-gc.stopCh:
			return
		}
	}
}

// newArena yeni bir arena oluşturur
func (gc *GarbageCollector) newArena() *Arena {
	// Basit implementasyon: Go heap'inden allocate et
	// Production'da mmap kullanılır
	buf := make([]byte, arenaSize)
	arena := &Arena{
		start: uintptr(unsafe.Pointer(&buf[0])),
		end:   uintptr(unsafe.Pointer(&buf[0])) + arenaSize,
		free:  uintptr(unsafe.Pointer(&buf[0])),
	}

	gc.arenas = append(gc.arenas, arena)
	atomic.AddInt64(&gc.stats.HeapSize, arenaSize)

	return arena
}

// tryAlloc arena'dan allocate etmeyi dene
func (a *Arena) tryAlloc(size uintptr) uintptr {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Align to 8 bytes
	size = (size + 7) &^ 7

	if a.free+size > a.end {
		return 0 // Arena full
	}

	ptr := a.free
	a.free += size
	return ptr
}

// Stats GC istatistiklerini döndürür
func (gc *GarbageCollector) Stats() GCStats {
	gc.statsMu.RLock()
	defer gc.statsMu.RUnlock()

	stats := gc.stats
	stats.HeapSize = atomic.LoadInt64(&gc.stats.HeapSize)
	stats.HeapUsed = atomic.LoadInt64(&gc.stats.HeapUsed)
	stats.NumGoroutine = runtime.NumGoroutine()

	return stats
}

// ForceGC zorla GC çalıştırır
func (gc *GarbageCollector) ForceGC() {
	gc.Collect()
}

// SetGCPercent GC trigger yüzdesini ayarlar
func (gc *GarbageCollector) SetGCPercent(percent int) {
	// GC trigger ratio'yu ayarla
	if percent > 0 {
		gc.triggerSize = int64(float64(atomic.LoadInt64(&gc.stats.HeapUsed)) * float64(percent) / 100.0)
	}
}

// Stop GC'yi durdurur
func (gc *GarbageCollector) Stop() {
	close(gc.stopCh)
	gc.wg.Wait()
}

func init() {
	GC = NewGC()
}
