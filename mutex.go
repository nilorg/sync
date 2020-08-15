package sync

import "errors"

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
