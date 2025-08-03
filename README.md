# sync

distributed  sync lock

分布式同步锁（支持 Redis 实现）

## 特性
- 支持分布式互斥锁
- 支持阻塞等待加锁（LockWait）
- 支持自定义锁超时、重试间隔

## 安装
```shell
go get github.com/nilorg/sync
```

## 快速开始
```go
import (
    "context"
    "github.com/go-redis/redis/v8"
    "github.com/nilorg/sync"
)

func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    rs := sync.NewRedisSync(rdb)
    m := rs.NewMutex("my_key")
    // 普通加锁
    if err := m.Lock(); err != nil {
        // 处理加锁失败
    }
    // 解锁
    _ = m.Unlock()

    // 阻塞等待加锁
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := m.LockWait(ctx); err != nil {
        // 处理超时或取消
    }
    _ = m.Unlock()
}
```

## 接口说明

### Mutexer
```
type Mutexer interface {
    Lock() error
    Unlock() error
}
```

### WaitableMutexer
```
type WaitableMutexer interface {
    Mutexer
    LockWait(ctx context.Context) error // 阻塞直到加锁成功或上下文取消
}
```

## 测试
```shell
go test -v
```

## License
MIT
