package runtime

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Scheduler async task scheduler
type Scheduler struct {
	eventLoop   *EventLoop
	taskQueue   *PriorityQueue
	mu          sync.Mutex
	cond        *sync.Cond
	running     atomic.Bool
	stopCh      chan struct{}
	wg          sync.WaitGroup
	workerCount int
}

// ScheduledTask zamanlanmış task
type ScheduledTask struct {
	task     *Task
	priority int
	deadline time.Time
	index    int // heap index
}

// PriorityQueue öncelik kuyruğu
type PriorityQueue []*ScheduledTask

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Önce deadline'a göre, sonra priority'ye göre
	if pq[i].deadline.Equal(pq[j].deadline) {
		return pq[i].priority > pq[j].priority
	}
	return pq[i].deadline.Before(pq[j].deadline)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*ScheduledTask)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// NewScheduler yeni bir scheduler oluşturur
func NewScheduler(workers int) *Scheduler {
	if workers <= 0 {
		workers = 4
	}

	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	s := &Scheduler{
		eventLoop:   NewEventLoop(workers),
		taskQueue:   &pq,
		workerCount: workers,
		stopCh:      make(chan struct{}),
	}
	s.cond = sync.NewCond(&s.mu)

	return s
}

// Start scheduler'ı başlatır
func (s *Scheduler) Start() {
	if !s.running.CompareAndSwap(false, true) {
		return
	}

	s.eventLoop.Start()

	// Scheduler worker
	s.wg.Add(1)
	go s.schedulerWorker()
}

// Stop scheduler'ı durdurur
func (s *Scheduler) Stop() {
	if !s.running.CompareAndSwap(true, false) {
		return
	}

	close(s.stopCh)
	s.cond.Broadcast()
	s.wg.Wait()
	s.eventLoop.Stop()
}

// ScheduleTask task'i zamanla
func (s *Scheduler) ScheduleTask(task *Task, priority int, delay time.Duration) {
	deadline := time.Now().Add(delay)

	st := &ScheduledTask{
		task:     task,
		priority: priority,
		deadline: deadline,
	}

	s.mu.Lock()
	heap.Push(s.taskQueue, st)
	s.mu.Unlock()

	s.cond.Signal()
}

// ScheduleImmediate task'i hemen zamanla
func (s *Scheduler) ScheduleImmediate(task *Task) {
	s.ScheduleTask(task, 10, 0)
}

// ScheduleDelayed task'i gecikmeli zamanla
func (s *Scheduler) ScheduleDelayed(task *Task, delay time.Duration) {
	s.ScheduleTask(task, 5, delay)
}

// schedulerWorker ana scheduler loop
func (s *Scheduler) schedulerWorker() {
	defer s.wg.Done()

	for {
		s.mu.Lock()

		for s.taskQueue.Len() == 0 {
			// Queue boş, wait
			s.cond.Wait()

			if !s.running.Load() {
				s.mu.Unlock()
				return
			}
		}

		// Peek en yüksek öncelikli task
		st := (*s.taskQueue)[0]
		now := time.Now()

		if now.Before(st.deadline) {
			// Henüz zamanı gelmedi, wait
			s.mu.Unlock()

			sleepDuration := st.deadline.Sub(now)
			timer := time.NewTimer(sleepDuration)

			select {
			case <-timer.C:
				continue
			case <-s.stopCh:
				timer.Stop()
				return
			}
		}

		// Pop task
		heap.Pop(s.taskQueue)
		s.mu.Unlock()

		// Schedule on event loop
		s.eventLoop.Schedule(st.task)
	}
}

// Yield coroutine yield implementasyonu
type Yielder struct {
	values []interface{}
	mu     sync.Mutex
	done   bool
}

// NewYielder yeni bir yielder oluşturur
func NewYielder() *Yielder {
	return &Yielder{
		values: make([]interface{}, 0),
	}
}

// Yield değer üretir
func (y *Yielder) Yield(value interface{}) {
	y.mu.Lock()
	y.values = append(y.values, value)
	y.mu.Unlock()
}

// Next bir sonraki değeri alır
func (y *Yielder) Next() (interface{}, bool) {
	y.mu.Lock()
	defer y.mu.Unlock()

	if len(y.values) == 0 {
		return nil, false
	}

	val := y.values[0]
	y.values = y.values[1:]
	return val, true
}

// Done tamamlandı işaretle
func (y *Yielder) Done() {
	y.mu.Lock()
	y.done = true
	y.mu.Unlock()
}

// IsDone tamamlandı mı kontrol eder
func (y *Yielder) IsDone() bool {
	y.mu.Lock()
	defer y.mu.Unlock()
	return y.done && len(y.values) == 0
}

// Coroutine coroutine implementasyonu
type Coroutine struct {
	fn      func(*Yielder, context.Context) error
	yielder *Yielder
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	err     error
	mu      sync.RWMutex
}

// NewCoroutine yeni bir coroutine oluşturur
func NewCoroutine(fn func(*Yielder, context.Context) error) *Coroutine {
	ctx, cancel := context.WithCancel(context.Background())

	co := &Coroutine{
		fn:      fn,
		yielder: NewYielder(),
		ctx:     ctx,
		cancel:  cancel,
		done:    make(chan struct{}),
	}

	// Coroutine'i başlat
	go func() {
		defer close(co.done)
		defer co.yielder.Done()

		co.mu.Lock()
		err := co.fn(co.yielder, co.ctx)
		co.err = err
		co.mu.Unlock()
	}()

	return co
}

// Next bir sonraki yielded değeri alır
func (co *Coroutine) Next() (interface{}, bool) {
	return co.yielder.Next()
}

// Cancel coroutine'i iptal eder
func (co *Coroutine) Cancel() {
	co.cancel()
}

// Wait coroutine'in tamamlanmasını bekler
func (co *Coroutine) Wait() error {
	<-co.done
	co.mu.RLock()
	defer co.mu.RUnlock()
	return co.err
}

// Note: Channel, Select, SelectCase moved to separate files:
// - channel.go: Full Channel implementation
// - select.go: Select multiplexer
// See those files for the complete implementations
