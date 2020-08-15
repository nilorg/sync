package sync

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
)

// func TestRedisMutex(t *testing.T) {
// 	var ctx = context.Background()
// 	rdb := redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "", // no password set
// 		DB:       0,  // use default DB
// 	})
// 	pong, err := rdb.Ping(ctx).Result()
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	t.Log(pong)
// 	rs := NewRedisSync(rdb)
// 	m := rs.NewMutex("test_key")
// 	m.Lock()
// 	for i := 0; i < 10; i++ {
// 		err = m.Lock()
// 		fmt.Println("lock:", err)
// 		time.Sleep(time.Second * 2)
// 		if i%2 == 0 {
// 			fmt.Println("2的倍数解锁")
// 			err = m.Unlock()
// 			if err != nil {
// 				fmt.Println("unlock:", err)
// 			}
// 		}
// 	}
// 	err = m.Unlock()
// 	if err != nil {
// 		fmt.Println("unlock:", err)
// 	}
// }

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
