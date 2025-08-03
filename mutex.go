package sync

import (
	"context"
	"errors"
)

var (
	// ErrLockFailed 加锁失败
	ErrLockFailed = errors.New("lock failed")
	// ErrUnlockFailed 解锁失败
	ErrUnlockFailed = errors.New("unlock failed")
)

// Mutexer  互斥锁
type Mutexer interface {
	Lock() (err error)
	Unlock() (err error)
}

// WaitableMutexer 可等待的互斥锁，加锁失败时不直接返回错误而是等待加锁成功
type WaitableMutexer interface {
	Mutexer
	// LockWait 等待加锁成功，会一直重试直到加锁成功或者上下文被取消
	LockWait(ctx context.Context) (err error)
}
