package xsync

import (
	"sync"
)

// OnceAtATime is similar to sync.Once, but is reusable.
// It restricts that the function passed to Do will not be run concurrently.
// If when such a function is running, another call to Do is done, then that
// other call will wait for the already running function to complete and then
// return, without running its own instance of the function.
//
// Intended for use with identical Do functions that are idempotent, for example
// database cleaning tasks or connection pool updating tasks.
type OnceAtATime struct {
	running bool
	mtx     sync.Mutex
	cond    *sync.Cond
}

// Do runs fn in case another Do is not already running a fn, otherwise wait for the other to complete.
// The instance that actually ran fn will also run the onDid functions before exiting.
func (o *OnceAtATime) Do(fn func(), onDid ...func()) {
	o.mtx.Lock()
	if o.cond == nil {
		o.cond = sync.NewCond(&o.mtx)
	}
	if o.running {
		// Wait for the runner, then return.
		o.cond.Wait()
		// No need to check if running == true in a loop, as recommended by the sync.Cond.Wait documentation,
		// because we only require that it has completed one run since this function was called,
		// it is possible and allowed that it has completed multiple runs and/or is currently running while leaving.
		o.mtx.Unlock()
	} else {
		o.running = true
		o.mtx.Unlock()

		// Run the function, which might take a long time, that is why we don't hold the mutex.
		fn()

		o.mtx.Lock()
		o.running = false
		for _, od := range onDid {
			od()
		}
		// Tell all waiters to stop waiting.
		o.cond.Broadcast()
		o.mtx.Unlock()
	}
}
