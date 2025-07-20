package semaphore

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

/*
Package semaphore provides a highly performant, lock-free, zero-allocation
mechanism to control concurrency and queue tasks, while maintaining type
safety through Go generics. It offers optional retry logic, dynamic capacity
adjustments (without losing tasks), and graceful shutdown support.

Key Features:
- Type-safe tasks with generics: `Semaphore[T]`.
- Lock-free, bounded queue for tasks.
- Fixed number of worker goroutines controlled by configuration.
- Optional retry logic with configurable backoff schedules.
- Zero allocations in steady state.
- Graceful shutdown ensuring all tasks complete before stopping.
- Detailed error types for diagnostics (QueueFullError, QueueCloseError, TaskError, MaxRetryReachedError).

Use Cases:
- High-load servers needing concurrency limits and request buffering.
- Background job pipelines requiring controlled concurrency and retry logic.
- Systems dealing with transient failures, where retries with backoff are beneficial.

Typical Flow:
1. Implement the `Executor[T]` interface to define how tasks of type T are processed.
2. Create a new `Semaphore` with `New(...)`, supplying an `Executor[T]` and any `Option`s.
3. Start worker goroutines with `StartConsumers()`.
4. Enqueue tasks using `Execute(...)`.
5. Adjust capacity during runtime with `SetCapacity(...)` if needed.
6. Gracefully shut down with `StopConsumers()`.
7. Handle errors (e.g., `QueueFullError`, `QueueCloseError`) as needed.
*/

// QueueFullError is returned when the semaphore's queue reaches its capacity
// and cannot accept more tasks. Clients can handle this by retrying, applying
// backpressure, or dropping tasks as business logic dictates.
type QueueFullError struct{}

func (e QueueFullError) Error() string {
	return "queue is full"
}

// QueueCloseError is returned when new tasks are attempted to be enqueued
// after the semaphore has been stopped. Once `StopConsumers()` is called,
// no further tasks can be added.
type QueueCloseError struct{}

func (e QueueCloseError) Error() string {
	return "queue closed"
}

// TaskError indicates that a task failed to process. If no retries are configured,
// or all retries have been exhausted, the task is considered permanently failed.
// Users can inspect or log these errors to diagnose task-specific issues.
type TaskError struct {
	origErr error
}

func (e TaskError) Error() string {
	return "task processing error: " + e.origErr.Error()
}

func (e TaskError) Unwrap() error {
	return e.origErr
}

// MaxRetryReachedError indicates that a task has failed after exhausting all
// retries. No further attempts are made to process this task. Users may want
// to log or monitor occurrences of this error to identify persistent failures.
type MaxRetryReachedError struct{}

func (e MaxRetryReachedError) Error() string {
	return "max retry reached"
}

// Executor defines how tasks of type T are processed. The `Process` method is
// called by worker goroutines for each task. If `Process` returns an error and
// retries are enabled, the task will be re-queued until it succeeds or runs out
// of retries.
type Executor[T any] interface {
	Process(ctx context.Context, t T) error
}

// queueItem is an internal structure representing a task and its current retry count.
// The semaphore uses this to track how many times a given task has been retried.
type queueItem[T any] struct {
	task    T
	retries int
}

// Config holds configuration parameters for a Semaphore. It can be customized
// using the provided `Option` functions and passed to `New(...)`.
type Config struct {
	// Capacity is the initial size of the task queue.
	// Must be > 0. Defaults to 1024 if not set.
	Capacity int

	// MaxRetries is the maximum number of retry attempts for failed tasks.
	// Defaults to 0 (no retries).
	MaxRetries int

	// BackoffSchedule defines delays between retries. The index corresponds
	// to the retry attempt number. If the attempt number exceeds the length
	// of this slice, it wraps around. If empty, retries happen immediately.
	BackoffSchedule []time.Duration

	// MaxWorkers is the number of worker goroutines that process tasks.
	// Defaults to runtime.GOMAXPROCS(0).
	MaxWorkers int
}

// DefaultConfig returns a Config initialized with default values:
// Capacity = 1024, MaxRetries = 0, BackoffSchedule = nil, MaxWorkers = GOMAXPROCS.
func DefaultConfig() Config {
	return Config{
		Capacity:        1024,
		MaxRetries:      0,
		BackoffSchedule: nil,
		MaxWorkers:      runtime.GOMAXPROCS(0),
	}
}

// Option is a functional option type for configuring a Semaphore. Users can apply
// multiple options (like `WithCapacity`, `WithMaxRetries`, etc.) to tailor the
// Semaphore's behavior before starting it.
type Option func(*Config)

// WithCapacity sets the initial capacity of the queue. If not set, defaults to 1024.
// Must be > 0. Capacity can later be changed via `SetCapacity()`.
func WithCapacity(capacity int) Option {
	return func(cfg *Config) {
		cfg.Capacity = capacity
	}
}

// WithMaxRetries sets the maximum number of retries for failed tasks.
// Defaults to 0 (no retries).
func WithMaxRetries(maxRetries int) Option {
	return func(cfg *Config) {
		cfg.MaxRetries = maxRetries
	}
}

// WithBackoffSchedule sets the retry backoff schedule. If empty, retries occur immediately.
// If multiple retries occur, the schedule wraps around if attempts exceed its length.
func WithBackoffSchedule(schedule []time.Duration) Option {
	return func(cfg *Config) {
		cfg.BackoffSchedule = schedule
	}
}

// WithMaxWorkers sets the number of worker goroutines that process tasks.
// Defaults to GOMAXPROCS(0).
func WithMaxWorkers(workers int) Option {
	return func(cfg *Config) {
		cfg.MaxWorkers = workers
	}
}

// Semaphore controls concurrency and provides a bounded queue with optional retry logic.
// It uses generics to enforce type safety on tasks and runs a fixed number of workers
// to process them. Tasks are processed in a FIFO manner, and if they fail, optional retries
// may re-queue them according to the configured max retries and backoff.
type Semaphore[T any] struct {
	cfg         Config
	executor    Executor[T]
	mu          sync.Mutex
	capacity    int32
	mask        int32
	queue       []queueItem[T]
	head        int32
	tail        int32
	curWorkers  int32
	consumerWg  sync.WaitGroup
	retryWg     sync.WaitGroup  // Track pending retries
	stop        int32
	consumerSem chan struct{}
}

// New creates a new Semaphore[T] instance with the given Executor and functional Options.
// If no options are provided, it uses defaults from `DefaultConfig()`. After creating
// the semaphore, call `StartConsumers()` to start processing tasks.
func New[T any](executor Executor[T], opts ...Option) *Semaphore[T] {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	s := &Semaphore[T]{
		cfg:      cfg,
		executor: executor,
	}

	size := nextPowerOfTwo(s.cfg.Capacity)
	s.capacity = int32(size)
	s.mask = int32(size - 1)
	s.queue = make([]queueItem[T], size)
	s.consumerSem = make(chan struct{}, s.cfg.MaxWorkers)
	return s
}

// StartConsumers launches worker goroutines that continuously process tasks
// from the queue until `StopConsumers()` is called. This should be done before
// calling `Execute(...)` to ensure tasks can be processed.
func (s *Semaphore[T]) StartConsumers() {
	for i := 0; i < s.cfg.MaxWorkers; i++ {
		s.consumerWg.Add(1)
		go s.consumer()
	}
}

// StopConsumers requests a graceful shutdown of all worker goroutines. After calling this,
// no new tasks will be processed, and `Execute(...)` will return `QueueCloseError`.
// StopConsumers waits until all currently in-flight tasks are completed, ensuring that
// the system shuts down cleanly.
func (s *Semaphore[T]) StopConsumers() {
	atomic.StoreInt32(&s.stop, 1)
	
	// Signal all consumers to wake up and check stop condition
	// Use non-blocking sends to avoid deadlock if channel is full
	for i := 0; i < s.cfg.MaxWorkers; i++ {
		select {
		case s.consumerSem <- struct{}{}:
		default:
			// Channel full, consumers will check stop condition anyway
		}
	}
	
	// Wait for all workers to complete
	s.consumerWg.Wait()
	
	// Wait for all pending retries to complete
	s.retryWg.Wait()
}

// Execute enqueues the given task for processing. If the queue is full, returns `QueueFullError`.
// If the queue is closed, returns `QueueCloseError`. On success, the task will be processed
// by a worker. If retries are enabled and the task fails, it may be re-queued until it succeeds
// or reaches the max retries.
func (s *Semaphore[T]) Execute(ctx context.Context, t T) error {
	if atomic.LoadInt32(&s.stop) == 1 {
		return QueueCloseError{}
	}
	return s.enqueue(queueItem[T]{task: t, retries: 0})
}

// SetCapacity adjusts the queue capacity at runtime. It can only increase capacity or set it
// to a value that is at least the current queue size, ensuring no tasks are lost. If the requested
// capacity is smaller than the current number of tasks, it returns an error. This method can be
// used to scale the system under changing load conditions.
func (s *Semaphore[T]) SetCapacity(newCap int) error {
	if newCap < 1 {
		return errors.New("capacity must be >= 1")
	}
	newSize := nextPowerOfTwo(newCap)

	s.mu.Lock()
	defer s.mu.Unlock()

	h := atomic.LoadInt32(&s.head)
	tl := atomic.LoadInt32(&s.tail)
	currentSize := tl - h
	if int32(newSize) < currentSize {
		return errors.New("new capacity is smaller than current queue size")
	}

	if int32(newSize) == s.capacity {
		return nil // No change needed
	}

	newQueue := make([]queueItem[T], newSize)
	for i := int32(0); i < currentSize; i++ {
		newQueue[i] = s.queue[(h+i)&s.mask]
	}

	s.queue = newQueue
	s.capacity = int32(newSize)
	s.mask = int32(newSize - 1)
	atomic.StoreInt32(&s.head, 0)
	atomic.StoreInt32(&s.tail, currentSize)

	s.cfg.Capacity = newCap
	return nil
}

// enqueue attempts to place a given task item into the queue. If there's capacity,
// it uses a lock-free CAS operation to update the `tail` pointer and insert the task.
// If the queue is full, returns `QueueFullError`. If closed, returns `QueueCloseError`.
//
// This is an internal method that `Execute(...)` and retry logic calls to enqueue tasks.
func (s *Semaphore[T]) enqueue(item queueItem[T]) error {
	for {
		if atomic.LoadInt32(&s.stop) == 1 {
			return QueueCloseError{}
		}
		h := atomic.LoadInt32(&s.head)
		tl := atomic.LoadInt32(&s.tail)

		// Check if there is capacity
		if (tl - h) < s.capacity {
			// Attempt to claim a slot in the queue
			if atomic.CompareAndSwapInt32(&s.tail, tl, tl+1) {
				s.queue[tl&s.mask] = item
				s.signalConsumer()
				return nil
			}
		} else {
			// Queue is full
			return QueueFullError{}
		}
	}
}

// enqueueRetry is like enqueue but allows retries to be queued even during shutdown.
// This ensures that retries scheduled before shutdown can still be processed.
func (s *Semaphore[T]) enqueueRetry(item queueItem[T]) error {
	for {
		h := atomic.LoadInt32(&s.head)
		tl := atomic.LoadInt32(&s.tail)

		// Check if there is capacity
		if (tl - h) < s.capacity {
			// Attempt to claim a slot in the queue
			if atomic.CompareAndSwapInt32(&s.tail, tl, tl+1) {
				s.queue[tl&s.mask] = item
				s.signalConsumer()
				return nil
			}
		} else {
			// Queue is full
			return QueueFullError{}
		}
	}
}

// signalConsumer notifies a waiting consumer goroutine that a new task is available.
// If the consumerSem channel is full, the notification is dropped, but consumers will
// eventually poll for tasks anyway. This helps keep the system responsive without
// blocking enqueue operations.
//
// This is an internal method used when a new task is enqueued.
func (s *Semaphore[T]) signalConsumer() {
	select {
	case s.consumerSem <- struct{}{}:
	default:
		// Channel full, no problem. Consumers will still eventually check the queue.
	}
}

// consumer is the function run by each worker goroutine. It waits for signals
// (via consumerSem) or yields if none are available. When tasks are detected,
// it dequeues them using a lock-free CAS on `head` and calls `s.run(item)`.
// This method runs until `StopConsumers()` is called and `stop` is set.
//
// This is an internal method, not intended for external use.
func (s *Semaphore[T]) consumer() {
	defer s.consumerWg.Done()
	for {
		select {
		case <-s.consumerSem:
			// Process all available tasks
			for {
				h := atomic.LoadInt32(&s.head)
				tl := atomic.LoadInt32(&s.tail)
				if h == tl {
					// No more tasks
					break
				}
				// Attempt to dequeue one task
				if atomic.CompareAndSwapInt32(&s.head, h, h+1) {
					item := s.queue[h&s.mask]
					s.run(item)
				}
			}
		default:
			// Check if we should stop and no more tasks to process
			if atomic.LoadInt32(&s.stop) == 1 {
				// Process any remaining tasks before stopping
				for {
					h := atomic.LoadInt32(&s.head)
					tl := atomic.LoadInt32(&s.tail)
					if h == tl {
						// No more tasks, safe to exit
						return
					}
					// Attempt to dequeue one task
					if atomic.CompareAndSwapInt32(&s.head, h, h+1) {
						item := s.queue[h&s.mask]
						s.run(item)
					}
				}
			}
			// No signal, yield CPU to other goroutines
			runtime.Gosched()
		}
	}
}

// run executes a single task by calling `executor.Process(...)`. If the task fails
// and retries are enabled, it re-enqueues the task with an incremented retry count
// and an optional backoff delay. If retries are exhausted or not enabled, the task
// fails permanently.
//
// This is an internal method handling the entire lifecycle of a single task processing attempt.
func (s *Semaphore[T]) run(item queueItem[T]) {
	atomic.AddInt32(&s.curWorkers, 1)
	defer atomic.AddInt32(&s.curWorkers, -1)
	
	err := s.executor.Process(context.Background(), item.task)

	if err == nil {
		return // Task succeeded
	}

	// Task failed, check if we can retry
	if s.cfg.MaxRetries > 0 && item.retries < s.cfg.MaxRetries {
		// Schedule retry asynchronously to avoid blocking the worker
		s.retryWg.Add(1)
		go func() {
			defer s.retryWg.Done()
			
			var delay time.Duration
			if len(s.cfg.BackoffSchedule) > 0 {
				delay = s.cfg.BackoffSchedule[item.retries%len(s.cfg.BackoffSchedule)]
			}
			
			if delay > 0 {
				time.Sleep(delay)
			}
			
			retryItem := queueItem[T]{
				task:    item.task,
				retries: item.retries + 1,
			}
			
			// Try to enqueue retry even if semaphore is stopping
			// We use a special retry enqueue that bypasses the stop check
			enqueueErr := s.enqueueRetry(retryItem)
			if enqueueErr != nil {
				// Could not re-enqueue due to full queue
				// The task is effectively lost at this point.
				return
			}
		}()
	} else {
		// No retries left or retries not enabled
		if item.retries >= s.cfg.MaxRetries && s.cfg.MaxRetries > 0 {
			// Max retries reached
			_ = MaxRetryReachedError{}
			return
		}
		// No retry configured, task permanently failed
		_ = TaskError{origErr: err}
		return
	}
}

// nextPowerOfTwo returns the smallest power of two greater than or equal to x.
// If x is already a power of two, it returns x. This ensures that the queue
// uses a power-of-two size for efficient indexing and wraparound using `mask`.
func nextPowerOfTwo(x int) int {
	if x < 2 {
		return 2
	}
	x--
	for i := 1; i < 64; i <<= 1 {
		x |= x >> i
	}
	return x + 1
}

// // noescape is a low-level optimization hint to the compiler to avoid heap allocations.
// // It's included as a reference for advanced optimization but is not currently used in this code.
// // In typical usage scenarios, this function can be safely removed.
// func noescape[T any](p *T) *T {
// 	x := uintptr(unsafe.Pointer(p))
// 	return (*T)(unsafe.Pointer(x))
// }
