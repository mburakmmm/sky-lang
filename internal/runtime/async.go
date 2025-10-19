package runtime

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// TaskState async task durumu
type TaskState int

const (
	TaskPending TaskState = iota
	TaskRunning
	TaskCompleted
	TaskFailed
	TaskCancelled
)

func (ts TaskState) String() string {
	switch ts {
	case TaskPending:
		return "pending"
	case TaskRunning:
		return "running"
	case TaskCompleted:
		return "completed"
	case TaskFailed:
		return "failed"
	case TaskCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Task async görev
type Task struct {
	id         uint64
	state      atomic.Value // TaskState
	fn         func(context.Context) (interface{}, error)
	result     interface{}
	err        error
	mu         sync.RWMutex
	done       chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	parent     *Task
	children   []*Task
	childrenMu sync.RWMutex
}

// Future async sonuç
type Future struct {
	task *Task
}

// NewTask yeni bir async task oluşturur
func NewTask(fn func(context.Context) (interface{}, error)) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		id:     atomic.AddUint64(&taskIDCounter, 1),
		fn:     fn,
		done:   make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}
	task.state.Store(TaskPending)
	return task
}

var taskIDCounter uint64

// ID task ID'sini döndürür
func (t *Task) ID() uint64 {
	return t.id
}

// State task durumunu döndürür
func (t *Task) State() TaskState {
	return t.state.Load().(TaskState)
}

// Result task sonucunu döndürür (blocking)
func (t *Task) Result() (interface{}, error) {
	<-t.done
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.result, t.err
}

// Await task tamamlanmasını bekler
func (t *Task) Await() (interface{}, error) {
	return t.Result()
}

// Cancel task'i iptal eder
func (t *Task) Cancel() {
	if t.State() == TaskPending || t.State() == TaskRunning {
		t.state.Store(TaskCancelled)
		t.cancel()
		close(t.done)
	}
}

// Then continuation ekler
func (t *Task) Then(fn func(interface{}) (interface{}, error)) *Task {
	newTask := NewTask(func(ctx context.Context) (interface{}, error) {
		result, err := t.Await()
		if err != nil {
			return nil, err
		}
		return fn(result)
	})

	t.childrenMu.Lock()
	t.children = append(t.children, newTask)
	newTask.parent = t
	t.childrenMu.Unlock()

	return newTask
}

// Catch error handler ekler
func (t *Task) Catch(fn func(error) interface{}) *Task {
	newTask := NewTask(func(ctx context.Context) (interface{}, error) {
		_, err := t.Await()
		if err != nil {
			return fn(err), nil
		}
		return t.result, nil
	})

	t.childrenMu.Lock()
	t.children = append(t.children, newTask)
	newTask.parent = t
	t.childrenMu.Unlock()

	return newTask
}

// EventLoop async event loop
type EventLoop struct {
	running     atomic.Bool
	tasks       chan *Task
	workers     int
	wg          sync.WaitGroup
	mu          sync.RWMutex
	activeTasks map[uint64]*Task
	stopCh      chan struct{}

	// Timers and intervals
	timers   []*Timer
	timersMu sync.RWMutex

	// Microtask queue (Promise callbacks, etc.)
	microtasks   []*Task
	microtasksMu sync.Mutex

	// Stats
	stats EventLoopStats
}

// EventLoopStats event loop istatistikleri
type EventLoopStats struct {
	TasksScheduled   uint64
	TasksCompleted   uint64
	TasksFailed      uint64
	TasksCancelled   uint64
	MicrotasksQueued uint64
	TimersScheduled  uint64
}

// NewEventLoop yeni bir event loop oluşturur
func NewEventLoop(workers int) *EventLoop {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	el := &EventLoop{
		tasks:       make(chan *Task, 1024),
		workers:     workers,
		activeTasks: make(map[uint64]*Task),
		stopCh:      make(chan struct{}),
		timers:      make([]*Timer, 0),
		microtasks:  make([]*Task, 0, 256),
	}

	return el
}

// Start event loop'u başlatır
func (el *EventLoop) Start() {
	if !el.running.CompareAndSwap(false, true) {
		return
	}

	// Worker goroutines
	for i := 0; i < el.workers; i++ {
		el.wg.Add(1)
		go el.worker(i)
	}

	// Timer goroutine
	el.wg.Add(1)
	go el.timerWorker()

	// Microtask processor
	el.wg.Add(1)
	go el.microtaskProcessor()
}

// Stop event loop'u durdurur
func (el *EventLoop) Stop() {
	if !el.running.CompareAndSwap(true, false) {
		return
	}

	close(el.stopCh)
	close(el.tasks)
	el.wg.Wait()
}

// Schedule task'i event loop'a ekler
func (el *EventLoop) Schedule(task *Task) error {
	if !el.running.Load() {
		return fmt.Errorf("event loop not running")
	}

	atomic.AddUint64(&el.stats.TasksScheduled, 1)

	el.mu.Lock()
	el.activeTasks[task.ID()] = task
	el.mu.Unlock()

	select {
	case el.tasks <- task:
		return nil
	case <-el.stopCh:
		return fmt.Errorf("event loop stopped")
	}
}

// ScheduleMicrotask microtask ekler (öncelikli)
func (el *EventLoop) ScheduleMicrotask(task *Task) {
	el.microtasksMu.Lock()
	el.microtasks = append(el.microtasks, task)
	atomic.AddUint64(&el.stats.MicrotasksQueued, 1)
	el.microtasksMu.Unlock()
}

// worker event loop worker
func (el *EventLoop) worker(id int) {
	defer el.wg.Done()

	for {
		select {
		case task, ok := <-el.tasks:
			if !ok {
				return
			}
			el.runTask(task)

		case <-el.stopCh:
			return
		}
	}
}

// runTask task'i çalıştırır
func (el *EventLoop) runTask(task *Task) {
	task.state.Store(TaskRunning)

	defer func() {
		if r := recover(); r != nil {
			task.mu.Lock()
			task.err = fmt.Errorf("panic: %v", r)
			task.state.Store(TaskFailed)
			task.mu.Unlock()
			atomic.AddUint64(&el.stats.TasksFailed, 1)
		}

		close(task.done)

		el.mu.Lock()
		delete(el.activeTasks, task.ID())
		el.mu.Unlock()

		// Schedule children
		task.childrenMu.RLock()
		for _, child := range task.children {
			el.Schedule(child)
		}
		task.childrenMu.RUnlock()
	}()

	result, err := task.fn(task.ctx)

	task.mu.Lock()
	task.result = result
	task.err = err
	if err != nil {
		task.state.Store(TaskFailed)
		atomic.AddUint64(&el.stats.TasksFailed, 1)
	} else if task.State() == TaskCancelled {
		atomic.AddUint64(&el.stats.TasksCancelled, 1)
	} else {
		task.state.Store(TaskCompleted)
		atomic.AddUint64(&el.stats.TasksCompleted, 1)
	}
	task.mu.Unlock()
}

// microtaskProcessor microtask'ları işler
func (el *EventLoop) microtaskProcessor() {
	defer el.wg.Done()

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			el.processMicrotasks()
		case <-el.stopCh:
			return
		}
	}
}

// processMicrotasks bekleyen microtask'ları işler
func (el *EventLoop) processMicrotasks() {
	el.microtasksMu.Lock()
	if len(el.microtasks) == 0 {
		el.microtasksMu.Unlock()
		return
	}

	tasks := el.microtasks
	el.microtasks = make([]*Task, 0, 256)
	el.microtasksMu.Unlock()

	for _, task := range tasks {
		el.runTask(task)
	}
}

// Timer zamanlayıcı
type Timer struct {
	id        uint64
	task      *Task
	delay     time.Duration
	interval  bool
	nextRun   time.Time
	cancelled atomic.Bool
}

var timerIDCounter uint64

// SetTimeout belirtilen süre sonra task çalıştırır
func (el *EventLoop) SetTimeout(fn func(context.Context) (interface{}, error), delay time.Duration) *Timer {
	task := NewTask(fn)
	timer := &Timer{
		id:       atomic.AddUint64(&timerIDCounter, 1),
		task:     task,
		delay:    delay,
		interval: false,
		nextRun:  time.Now().Add(delay),
	}

	el.timersMu.Lock()
	el.timers = append(el.timers, timer)
	atomic.AddUint64(&el.stats.TimersScheduled, 1)
	el.timersMu.Unlock()

	return timer
}

// SetInterval belirtilen aralıklarla task çalıştırır
func (el *EventLoop) SetInterval(fn func(context.Context) (interface{}, error), interval time.Duration) *Timer {
	task := NewTask(fn)
	timer := &Timer{
		id:       atomic.AddUint64(&timerIDCounter, 1),
		task:     task,
		delay:    interval,
		interval: true,
		nextRun:  time.Now().Add(interval),
	}

	el.timersMu.Lock()
	el.timers = append(el.timers, timer)
	atomic.AddUint64(&el.stats.TimersScheduled, 1)
	el.timersMu.Unlock()

	return timer
}

// ClearTimer timer'ı iptal eder
func (t *Timer) Cancel() {
	t.cancelled.Store(true)
	if t.task != nil {
		t.task.Cancel()
	}
}

// timerWorker timer'ları işler
func (el *EventLoop) timerWorker() {
	defer el.wg.Done()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			el.processTimers()
		case <-el.stopCh:
			return
		}
	}
}

// processTimers çalışması gereken timer'ları işler
func (el *EventLoop) processTimers() {
	now := time.Now()

	el.timersMu.Lock()
	defer el.timersMu.Unlock()

	activeTimers := make([]*Timer, 0, len(el.timers))

	for _, timer := range el.timers {
		if timer.cancelled.Load() {
			continue
		}

		if now.After(timer.nextRun) || now.Equal(timer.nextRun) {
			// Schedule task
			el.Schedule(timer.task)

			if timer.interval {
				// Reschedule for next interval
				timer.nextRun = now.Add(timer.delay)
				timer.task = NewTask(timer.task.fn)
				activeTimers = append(activeTimers, timer)
			}
		} else {
			activeTimers = append(activeTimers, timer)
		}
	}

	el.timers = activeTimers
}

// Stats event loop istatistiklerini döndürür
func (el *EventLoop) Stats() EventLoopStats {
	return EventLoopStats{
		TasksScheduled:   atomic.LoadUint64(&el.stats.TasksScheduled),
		TasksCompleted:   atomic.LoadUint64(&el.stats.TasksCompleted),
		TasksFailed:      atomic.LoadUint64(&el.stats.TasksFailed),
		TasksCancelled:   atomic.LoadUint64(&el.stats.TasksCancelled),
		MicrotasksQueued: atomic.LoadUint64(&el.stats.MicrotasksQueued),
		TimersScheduled:  atomic.LoadUint64(&el.stats.TimersScheduled),
	}
}

// Promise JavaScript Promise benzeri implementasyon
type Promise struct {
	task      *Task
	eventLoop *EventLoop
}

// NewPromise yeni bir Promise oluşturur
func NewPromise(el *EventLoop, fn func(resolve func(interface{}), reject func(error))) *Promise {
	task := NewTask(func(ctx context.Context) (interface{}, error) {
		var result interface{}
		var err error
		done := make(chan struct{})

		resolve := func(val interface{}) {
			result = val
			close(done)
		}

		reject := func(e error) {
			err = e
			close(done)
		}

		go fn(resolve, reject)

		select {
		case <-done:
			return result, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})

	el.Schedule(task)

	return &Promise{
		task:      task,
		eventLoop: el,
	}
}

// Then Promise continuation
func (p *Promise) Then(fn func(interface{}) (interface{}, error)) *Promise {
	newTask := p.task.Then(fn)
	p.eventLoop.Schedule(newTask)
	return &Promise{
		task:      newTask,
		eventLoop: p.eventLoop,
	}
}

// Catch error handler
func (p *Promise) Catch(fn func(error) interface{}) *Promise {
	newTask := p.task.Catch(fn)
	p.eventLoop.Schedule(newTask)
	return &Promise{
		task:      newTask,
		eventLoop: p.eventLoop,
	}
}

// Await Promise sonucunu bekler
func (p *Promise) Await() (interface{}, error) {
	return p.task.Await()
}

// All tüm promise'lerin tamamlanmasını bekler
func All(el *EventLoop, promises ...*Promise) *Promise {
	task := NewTask(func(ctx context.Context) (interface{}, error) {
		results := make([]interface{}, len(promises))

		for i, p := range promises {
			result, err := p.Await()
			if err != nil {
				return nil, err
			}
			results[i] = result
		}

		return results, nil
	})

	el.Schedule(task)

	return &Promise{
		task:      task,
		eventLoop: el,
	}
}

// Race ilk tamamlanan promise'i döndürür
func Race(el *EventLoop, promises ...*Promise) *Promise {
	task := NewTask(func(ctx context.Context) (interface{}, error) {
		resultCh := make(chan interface{}, len(promises))
		errCh := make(chan error, len(promises))

		for _, p := range promises {
			go func(promise *Promise) {
				result, err := promise.Await()
				if err != nil {
					errCh <- err
				} else {
					resultCh <- result
				}
			}(p)
		}

		select {
		case result := <-resultCh:
			return result, nil
		case err := <-errCh:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})

	el.Schedule(task)

	return &Promise{
		task:      task,
		eventLoop: el,
	}
}

// Global event loop
var globalEventLoop *EventLoop
var globalEventLoopOnce sync.Once

// GetGlobalEventLoop global event loop'u döndürür
func GetGlobalEventLoop() *EventLoop {
	globalEventLoopOnce.Do(func() {
		globalEventLoop = NewEventLoop(runtime.NumCPU())
		globalEventLoop.Start()
	})
	return globalEventLoop
}

// Async yardımcı fonksiyon
func Async(fn func(context.Context) (interface{}, error)) *Future {
	task := NewTask(fn)
	GetGlobalEventLoop().Schedule(task)
	return &Future{task: task}
}

// Await Future sonucunu bekler
func (f *Future) Await() (interface{}, error) {
	return f.task.Await()
}
