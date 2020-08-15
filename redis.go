package sync

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisSync ...
type RedisSync struct {
	Opts  *Options
	Redis *redis.Client
}

// NewRedisSync ...
func NewRedisSync(redis *redis.Client, opts ...Option) *RedisSync {
	return &RedisSync{
		Redis: redis,
		Opts:  newOptions(opts...),
	}
}

// NewMutex 创建互斥锁
func (r *RedisSync) NewMutex(key string, opts ...Option) Mutexer {
	optx := *r.Opts
	for _, o := range opts {
		o(&optx)
	}
	rm := &redisMutex{
		opts:  &optx,
		redis: r.Redis,
		key:   key,
	}
	return rm
}

// redisMutex redis互斥锁
type redisMutex struct {
	opts       *Options
	redis      *redis.Client
	key        string
	lockCtx    context.Context
	lockCancel context.CancelFunc
}

// Lock 加锁
func (rm *redisMutex) Lock() (err error) {
	var flag bool
	lockName := rm.lockName()
	flag, err = rm.redis.SetNX(context.Background(), lockName, 1, rm.opts.LockTimeout).Result()
	if err != nil {
		return
	}
	if !flag {
		err = ErrLockFailed
		return
	}
	rm.lockCtx, rm.lockCancel = context.WithCancel(context.Background())
	go func() {
		t := time.NewTimer(rm.opts.WaitRetry)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				_ = rm.redis.Expire(context.Background(), lockName, rm.opts.WaitRetry)
				t.Reset(rm.opts.WaitRetry)
				// TODO: 如果延迟失败
				// fmt.Println("触发延迟key条件")
			case <-rm.lockCtx.Done():
				// fmt.Println("解锁触发上下文结束事件，结束整个for循环")
				return
			}
		}
	}()
	return
}

const (
	luaRelease = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
)

// Unlock 解锁
func (rm *redisMutex) Unlock() (err error) {
	var flag bool
	lockName := rm.lockName()
	flag, err = rm.redis.Eval(context.Background(), luaRelease, []string{lockName}, 1).Bool()
	if err != nil {
		return
	}
	if !flag {
		err = ErrUnlockFailed
		return
	}
	if rm.lockCancel != nil {
		rm.lockCancel()
	}
	return
}

func (rm *redisMutex) lockName() string {
	return rm.opts.KeyPrefix + rm.key
}
