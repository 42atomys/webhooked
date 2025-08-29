# semaphore - High-Performance Concurrency Control with Queued Tasks

**Why this package?**  
In high-load, concurrent environments, you need efficient control over task processing. You must ensure that you never exceed system limits, avoid dropping tasks unnecessarily, handle transient failures gracefully with retries, and adjust capacity as traffic changes. This semaphore package provides a zero-allocation, lock-free, bounded queue for tasks, with fixed worker pools and optional retry/backoff logic, all while maintaining type safety with Go generics.

## Key Features

- **Type-Safe Tasks:**  
  Use Go generics to enforce type safety for queued tasks.

- **Lock-Free, Bounded Queue:**  
  A wait-free, ring-buffer queue ensures minimal overhead and zero allocations in steady state.

- **Fixed Worker Pool:**  
  A specified number of goroutines (workers) process tasks concurrently, respecting concurrency limits.

- **Optional Retry & Backoff:**  
  Retry failed tasks seamlessly with a configurable schedule. If a task fails, it can be retried until success or until max retries are reached, with configurable backoff delays.

- **Dynamic Capacity Adjustments:**  
  Increase the queue capacity at runtime without losing tasks. Adapt to changing load profiles safely.

- **Graceful Shutdown:**  
  Stop workers gracefully, ensuring all queued tasks finish processing before shutdown. After stopping, attempts to add tasks return a clear `QueueCloseError`.

- **Clear Error Handling:**  
  Distinguish between a full queue (`QueueFullError`), a closed queue (`QueueCloseError`), task failures (`TaskError`), and max retries exhausted (`MaxRetryReachedError`).

## When to Use

- **High-Load APIs and Services:**  
  Handle thousands of requests per second while controlling concurrency to avoid overwhelming resources.

- **Background Job Processing:**  
  Efficiently run tasks like data transformations, notifications, and batch jobs in a controlled manner.

- **Retrying Unreliable Operations:**  
  For tasks that might fail due to transient issues (network glitches, temporary service overload), seamlessly re-queue and retry with backoff.

## Quick Start

1. **Implement `Executor[T]`:**
   ```go
   type MyExecutor struct{}
   
   func (e *MyExecutor) Process(ctx context.Context, t string) error {
       // Process the string task here.
       return nil
   }
   ```

2. **Create a `Semaphore`:**
   ```go
   s := semaphore.New(&MyExecutor{},
       semaphore.WithCapacity(2048),
       semaphore.WithMaxRetries(3),
       semaphore.WithBackoffSchedule([]time.Duration{time.Millisecond, 2*time.Millisecond}),
       semaphore.WithMaxWorkers(4),
   )
   ```

3. **Start Workers:**
   ```go
   s.StartConsumers()
   ```

4. **Enqueue Tasks:**
   ```go
   err := s.Execute(context.Background(), "my-task")
   if err != nil {
       if errors.Is(err, semaphore.QueueFullError{}) {
           // Handle queue full scenario (e.g. backpressure or dropping tasks)
       }
   }
   ```

5. **Adjust Capacity at Runtime (if needed):**
   ```go
   err = s.SetCapacity(4096) // Increase capacity
   if err != nil {
       // Handle error (e.g. requested capacity too small)
   }
   ```

6. **Gracefully Stop:**
   ```go
   s.StopConsumers()
   // Now s is closed, Execute() will return QueueCloseError
   ```

## Error Types

- **QueueFullError:**  
  Returned when the queue is at capacity. Decide whether to retry enqueue later, drop the task, or return an error to the caller.

- **QueueCloseError:**  
  Returned when `Execute()` is called after `StopConsumers()` has been invoked. This ensures no tasks are accepted post-shutdown.

- **TaskError:**  
  Indicates the task failed and won't be retried further (either because retries aren't enabled or max retries have been reached).

- **MaxRetryReachedError:**  
  Indicates the task failed after all retries were attempted. Useful for logging or alerting on persistent failures.

## Configuration Options

- **WithCapacity(int):** Initial queue capacity (default: 1024)
- **WithMaxRetries(int):** Maximum retries per failed task (default: 0, no retries)
- **WithBackoffSchedule([]time.Duration):** Backoff delays for retries, cycling through if attempts exceed the schedule length
- **WithMaxWorkers(int):** Number of worker goroutines (default: GOMAXPROCS(0))

## Internals

- **Zero-Allocation & Lock-Free:**  
  The internal ring-buffer and CAS operations ensure tasks move efficiently from enqueue to dequeue without locks or extra allocations in a steady state.

- **Runtime Adjustments:**  
  `SetCapacity()` can increase the queue size without losing tasks. It cannot reduce capacity below the current queue size.

- **Retry Mechanics:**  
  Failed tasks are re-enqueued with incremented retry counts. If `MaxRetries` is reached, the task fails permanently.

## Example Use Cases

- **API Rate Limiting & Buffering:**  
  Prevent sudden spikes from overwhelming servers. If tasks exceed capacity, return `QueueFullError`.

- **Asynchronous Processing Pipelines:**  
  Offload CPU-heavy tasks to a controlled number of goroutines, queuing incoming tasks and retrying failures transparently.

- **Service-Oriented Architectures:**  
  Handle intermittent downstream failures by retrying requests with backoff, preventing cascading failures.

## Conclusion

This semaphore library offers a straightforward, efficient, and type-safe way to control concurrency, queue tasks, and handle transient failures through retries. By combining a lock-free structure, clear error semantics, flexible configuration, and graceful shutdown support, it helps you build robust, high-performance systems under heavy load.
