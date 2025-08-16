//go:build unit

package semaphore_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/42atomys/webhooked/semaphore"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testExecutor struct {
	processFunc func(ctx context.Context, t int) error
}

func (e *testExecutor) Process(ctx context.Context, t int) error {
	return e.processFunc(ctx, t)
}

func TestBasicFunctionality(t *testing.T) {
	execCalled := int32(0)
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			atomic.AddInt32(&execCalled, 1)
			return nil
		},
	}
	s := semaphore.New(exec) // Default settings
	s.StartConsumers()

	// Enqueue tasks
	for i := range 10 {
		err := s.Execute(context.Background(), i)
		require.NoError(t, err)
	}

	s.StopConsumers()

	assert.Equal(t, int32(10), atomic.LoadInt32(&execCalled))
}

func TestQueueFullError(t *testing.T) {
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			return nil
		},
	}
	// Use capacity 1 - don't start consumers so queue fills up
	// Note: capacity 1 becomes actual size 2 due to power-of-two rounding
	s := semaphore.New(exec, semaphore.WithCapacity(1))

	// Fill the queue (actual capacity is 2 due to power-of-two)
	err1 := s.Execute(context.Background(), 1)
	require.NoError(t, err1)

	err2 := s.Execute(context.Background(), 2)
	require.NoError(t, err2)

	// This one should fail because queue is full (capacity 2, 2 tasks queued)
	err3 := s.Execute(context.Background(), 3)
	require.Error(t, err3)
	assert.IsType(t, semaphore.QueueFullError{}, err3)
}

func TestQueueCloseError(t *testing.T) {
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error { return nil },
	}
	s := semaphore.New(exec, semaphore.WithCapacity(1))
	s.StartConsumers()

	err := s.Execute(context.Background(), 1)
	require.NoError(t, err)

	s.StopConsumers()

	// After stop, queue closed
	err = s.Execute(context.Background(), 2)
	require.Error(t, err)
	assert.IsType(t, semaphore.QueueCloseError{}, err)
}

func TestTaskErrorNoRetry(t *testing.T) {
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			return errors.New("fail")
		},
	}
	s := semaphore.New(exec, semaphore.WithCapacity(10))
	s.StartConsumers()

	err := s.Execute(context.Background(), 1)
	require.NoError(t, err)

	s.StopConsumers()

	// Task fails, no retry, just ensure no panic and done
}

func TestTaskErrorWithRetry(t *testing.T) {
	var attempts int32
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			// Fails first 2 attempts, succeeds on 3rd
			count := atomic.AddInt32(&attempts, 1)
			if count < 3 {
				return errors.New("fail")
			}
			return nil
		},
	}

	s := semaphore.New(exec,
		semaphore.WithMaxRetries(5),
		semaphore.WithBackoffSchedule([]time.Duration{time.Millisecond, time.Millisecond * 2}),
	)
	s.StartConsumers()

	err := s.Execute(context.Background(), 1)
	require.NoError(t, err)

	// Give some time for retries to be processed
	time.Sleep(10 * time.Millisecond)

	s.StopConsumers()

	// We expect 3 attempts total
	assert.Equal(t, int32(3), attempts)
}

func TestMaxRetryReached(t *testing.T) {
	var attempts int32
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			atomic.AddInt32(&attempts, 1)
			return errors.New("always fail")
		},
	}

	s := semaphore.New(exec,
		semaphore.WithMaxRetries(2), // 3 attempts total (initial + 2 retries)
		semaphore.WithBackoffSchedule([]time.Duration{time.Millisecond}),
	)
	s.StartConsumers()

	err := s.Execute(context.Background(), 10)
	require.NoError(t, err)

	// Give some time for retries to be processed
	time.Sleep(100 * time.Millisecond)

	s.StopConsumers()

	assert.Equal(t, int32(3), attempts)
	// Check it doesn't panic, max retries reached gracefully
}

func TestIncreaseCapacity(t *testing.T) {
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			time.Sleep(time.Millisecond * 10)
			return nil
		},
	}
	s := semaphore.New(exec, semaphore.WithCapacity(4))
	s.StartConsumers()

	// Fill the queue
	for i := 0; i < 4; i++ {
		err := s.Execute(context.Background(), i)
		require.NoError(t, err)
	}

	// Increase capacity to accommodate more tasks
	err := s.SetCapacity(8)
	require.NoError(t, err)

	// Now we can add more without error, even if not processed yet
	err = s.Execute(context.Background(), 99)
	require.NoError(t, err)

	s.StopConsumers()
}

func TestSetCapacitySmallerThanCurrentSize(t *testing.T) {
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			return nil
		},
	}
	// Use capacity 4 but don't start consumers so tasks queue up
	s := semaphore.New(exec, semaphore.WithCapacity(4))

	// Fill the queue without consumers running
	for i := 0; i < 4; i++ {
		err := s.Execute(context.Background(), i)
		require.NoError(t, err)
	}

	// Try to reduce capacity to 2 while 4 are queued
	err := s.SetCapacity(2)
	require.Error(t, err)
	assert.Equal(t, "new capacity is smaller than current queue size", err.Error())
}

func TestWithMaxWorkers(t *testing.T) {
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			time.Sleep(time.Millisecond)
			return nil
		},
	}

	s := semaphore.New(exec, semaphore.WithCapacity(20), semaphore.WithMaxWorkers(2))
	s.StartConsumers()

	start := time.Now()
	for i := 0; i < 10; i++ {
		require.NoError(t, s.Execute(context.Background(), i))
	}

	s.StopConsumers()
	elapsed := time.Since(start)

	// With only 2 workers, 10 tasks taking ~1ms each should take at least ~5ms (2 tasks at a time).
	assert.True(t, elapsed.Milliseconds() >= 5)
}

func TestBackoffScheduleWrapping(t *testing.T) {
	var attempts int32
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			atomic.AddInt32(&attempts, 1)
			return errors.New("fail")
		},
	}
	// Max retries 5 but schedule length 2, it should wrap around
	s := semaphore.New(exec,
		semaphore.WithMaxRetries(5),
		semaphore.WithBackoffSchedule([]time.Duration{time.Millisecond, time.Millisecond * 2}),
	)
	s.StartConsumers()
	require.NoError(t, s.Execute(context.Background(), 123))

	// Give some time for retries to be processed
	time.Sleep(20 * time.Millisecond)

	s.StopConsumers()

	// initial + 5 retries = 6 attempts total
	assert.Equal(t, int32(6), attempts)
}

func TestExecuteAfterStop(t *testing.T) {
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			return nil
		},
	}
	s := semaphore.New(exec)
	s.StartConsumers()
	s.StopConsumers()

	// After stop, should return QueueCloseError
	err := s.Execute(context.Background(), 1)
	require.Error(t, err)
	assert.IsType(t, semaphore.QueueCloseError{}, err)
}

// Benchmarks

func BenchmarkSemaphore_Execute(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			return nil
		},
	}
	s := semaphore.New(exec, semaphore.WithCapacity(10000))
	s.StartConsumers()
	defer s.StopConsumers()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.Execute(context.Background(), i)
		require.NoError(b, err)
	}
}

func BenchmarkSemaphore_Execute_WithWorkers(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			return nil
		},
	}
	s := semaphore.New(exec,
		semaphore.WithCapacity(10000),
		semaphore.WithMaxWorkers(10))
	s.StartConsumers()
	defer s.StopConsumers()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.Execute(context.Background(), i)
		require.NoError(b, err)
	}
}

func BenchmarkSemaphore_Execute_Concurrent(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			return nil
		},
	}
	s := semaphore.New(exec,
		semaphore.WithCapacity(10000),
		semaphore.WithMaxWorkers(20))
	s.StartConsumers()
	defer s.StopConsumers()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			err := s.Execute(context.Background(), i)
			require.NoError(b, err)
			i++
		}
	})
}

func BenchmarkSemaphore_WithRetries(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	failCount := int32(0)
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			// Fail 50% of the time
			if atomic.AddInt32(&failCount, 1)%2 == 0 {
				return errors.New("simulated failure")
			}
			return nil
		},
	}
	s := semaphore.New(exec,
		semaphore.WithCapacity(10000),
		semaphore.WithMaxRetries(2),
		semaphore.WithBackoffSchedule([]time.Duration{time.Microsecond}))
	s.StartConsumers()
	defer s.StopConsumers()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := s.Execute(context.Background(), i)
		require.NoError(b, err)
	}
}

func BenchmarkSemaphore_SetCapacity(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			return nil
		},
	}
	s := semaphore.New(exec, semaphore.WithCapacity(100))
	s.StartConsumers()
	defer s.StopConsumers()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newCapacity := int32(100 + (i % 100))
		err := s.SetCapacity(newCapacity)
		require.NoError(b, err)
	}
}

func BenchmarkSemaphore_ProcessingSpeed(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	processed := int32(0)
	exec := &testExecutor{
		processFunc: func(ctx context.Context, t int) error {
			atomic.AddInt32(&processed, 1)
			return nil
		},
	}
	s := semaphore.New(exec,
		semaphore.WithCapacity(int32(b.N)),
		semaphore.WithMaxWorkers(10))
	s.StartConsumers()

	// Fill the queue
	for i := 0; i < b.N; i++ {
		err := s.Execute(context.Background(), i)
		require.NoError(b, err)
	}

	// Wait for all to be processed
	start := time.Now()
	for atomic.LoadInt32(&processed) < int32(b.N) {
		time.Sleep(time.Millisecond)
	}
	elapsed := time.Since(start)

	s.StopConsumers()

	b.ReportMetric(float64(b.N)/elapsed.Seconds(), "tasks/sec")
}
