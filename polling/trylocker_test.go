package polling

import (
	"testing"

	"github.com/googollee/go-assert"
)

func TestTryLocker(t *testing.T) {
	l := tryLocker{}

	ok := l.TryLock()
	assert.Equal(t, ok, true)

	ok = l.TryLock()
	assert.Equal(t, ok, false)

	l.Unlock()

	ok = l.TryLock()
	assert.Equal(t, ok, true)
}
