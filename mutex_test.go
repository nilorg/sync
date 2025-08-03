package sync

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRedisMutex(t *testing.T) {
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(pong)
	rs := NewRedisSync(rdb)
	m := rs.NewMutex("test_key")
	m.Lock()
	for i := 0; i < 10; i++ {
		err = m.Lock()
		fmt.Println("lock:", err)
		time.Sleep(time.Second * 2)
		if i%2 == 0 {
			fmt.Println("2的倍数解锁")
			err = m.Unlock()
			if err != nil {
				fmt.Println("unlock:", err)
			}
		}
	}
	err = m.Unlock()
	if err != nil {
		fmt.Println("unlock:", err)
	}
}

func BenchmarkRedisMutex(b *testing.B) {
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		b.Error(err)
		return
	}
	b.Log(pong)
	rs := NewRedisSync(rdb)
	b.ResetTimer()
	m := rs.NewMutex("test_benchmark")
	for i := 0; i < b.N; i++ {
		err = m.Lock()
		if err != nil {
			b.Errorf("Lock: %s", err)
		}
		err = m.Unlock()
		if err != nil {
			b.Errorf("Unlock: %s", err)
		}
	}
}

func TestRedisMutexLockWait(t *testing.T) {
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(pong)
	rs := NewRedisSync(rdb)

	// 测试成功获取锁
	t.Run("成功获取锁", func(t *testing.T) {
		m := rs.NewMutex("test_lock_wait_success")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := m.LockWait(ctx)
		if err != nil {
			t.Errorf("LockWait failed: %v", err)
		}

		err = m.Unlock()
		if err != nil {
			t.Errorf("Unlock failed: %v", err)
		}
	})

	// 测试等待锁释放后成功获取
	t.Run("等待锁释放后成功获取", func(t *testing.T) {
		m1 := rs.NewMutex("test_lock_wait_retry")
		m2 := rs.NewMutex("test_lock_wait_retry")

		// m1 先获取锁
		err := m1.Lock()
		if err != nil {
			t.Errorf("m1 Lock failed: %v", err)
		}

		// 启动一个 goroutine，2秒后释放 m1 的锁
		go func() {
			time.Sleep(2 * time.Second)
			err := m1.Unlock()
			if err != nil {
				t.Errorf("m1 Unlock failed: %v", err)
			}
		}()

		// m2 等待获取锁
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		err = m2.LockWait(ctx)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("m2 LockWait failed: %v", err)
		}

		// 应该等待了大约 2 秒
		if duration < 1*time.Second || duration > 3*time.Second {
			t.Errorf("Expected wait time around 2s, got %v", duration)
		}

		err = m2.Unlock()
		if err != nil {
			t.Errorf("m2 Unlock failed: %v", err)
		}
	})

	// 测试上下文超时
	t.Run("上下文超时", func(t *testing.T) {
		m1 := rs.NewMutex("test_lock_wait_timeout")
		m2 := rs.NewMutex("test_lock_wait_timeout")

		// m1 先获取锁且不释放
		err := m1.Lock()
		if err != nil {
			t.Errorf("m1 Lock failed: %v", err)
		}
		defer m1.Unlock() // 确保测试结束后释放锁

		// m2 尝试获取锁，但设置很短的超时时间
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		start := time.Now()
		err = m2.LockWait(ctx)
		duration := time.Since(start)

		if err == nil {
			t.Error("Expected timeout error, but got nil")
		}

		if err != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded, got %v", err)
		}

		// 应该等待了大约 500ms
		if duration < 400*time.Millisecond || duration > 600*time.Millisecond {
			t.Errorf("Expected wait time around 500ms, got %v", duration)
		}
	})

	// 测试上下文取消
	t.Run("上下文取消", func(t *testing.T) {
		m1 := rs.NewMutex("test_lock_wait_cancel")
		m2 := rs.NewMutex("test_lock_wait_cancel")

		// m1 先获取锁且不释放
		err := m1.Lock()
		if err != nil {
			t.Errorf("m1 Lock failed: %v", err)
		}
		defer m1.Unlock() // 确保测试结束后释放锁

		// 创建可取消的上下文
		ctx, cancel := context.WithCancel(context.Background())

		// 启动一个 goroutine，1秒后取消上下文
		go func() {
			time.Sleep(1 * time.Second)
			cancel()
		}()

		start := time.Now()
		err = m2.LockWait(ctx)
		duration := time.Since(start)

		if err == nil {
			t.Error("Expected cancel error, but got nil")
		}

		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}

		// 应该等待了大约 1 秒
		if duration < 900*time.Millisecond || duration > 1100*time.Millisecond {
			t.Errorf("Expected wait time around 1s, got %v", duration)
		}
	})
}

func BenchmarkRedisMutexLockWait(b *testing.B) {
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		b.Error(err)
		return
	}
	b.Log(pong)
	rs := NewRedisSync(rdb)
	b.ResetTimer()

	// 测试无竞争情况下的 LockWait 性能
	b.Run("无竞争", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := rs.NewMutex(fmt.Sprintf("test_benchmark_lockwait_%d", i))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			err = m.LockWait(ctx)
			if err != nil {
				b.Errorf("LockWait: %s", err)
			}
			err = m.Unlock()
			if err != nil {
				b.Errorf("Unlock: %s", err)
			}
			cancel()
		}
	})
}
