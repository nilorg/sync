package sync

import "time"

// Options ...
type Options struct {
	KeyPrefix   string
	LockTimeout time.Duration
	WaitRetry   time.Duration
}

// Option 为可选参数赋值的函数
type Option func(*Options)

// newOptions 创建可选参数
func newOptions(opts ...Option) *Options {
	opt := &Options{
		KeyPrefix:   "synclock:",
		LockTimeout: 20 * time.Second,
		WaitRetry:   6 * time.Second,
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

// KeyPrefix ...
func KeyPrefix(keyPrefix string) Option {
	return func(o *Options) {
		o.KeyPrefix = keyPrefix
	}
}

// LockTimeout ...
func LockTimeout(lockTimeout time.Duration) Option {
	return func(o *Options) {
		o.LockTimeout = lockTimeout
	}
}

// WaitRetry ...
func WaitRetry(waitRetry time.Duration) Option {
	return func(o *Options) {
		o.WaitRetry = waitRetry
	}
}
