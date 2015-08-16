package polling

import "sync/atomic"

type tryLocker struct {
	locker int32
}

func (l *tryLocker) TryLock() bool {
	return atomic.CompareAndSwapInt32(&l.locker, 0, 1)
}

func (l *tryLocker) Unlock() {
	atomic.StoreInt32(&l.locker, 0)
}
