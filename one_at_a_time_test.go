package xsync

import (
	"sync"
	"testing"
	"time"
)

func TestOneAtATime(t *testing.T) {
	poison := false
	count := 0
	var o OneAtATime

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				countBefore := count
				o.Do(
					func() {
						if poison {
							t.Fatal("poison")
						}
						poison = true
						time.Sleep(time.Microsecond)
						poison = false
					},
					func() {
						count++
					})
				countAfter := count
				if countAfter <= countBefore {
					t.Errorf("%v <= %v", countAfter, countBefore)
				}
			}
		}(i)
	}
	wg.Wait()
}
