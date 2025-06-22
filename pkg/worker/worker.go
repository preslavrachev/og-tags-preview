package worker

import (
	"fmt"
	"log/slog"
	"sync"
)

var wg sync.WaitGroup

func Wait() {
	wg.Wait()
}

// Deathlock if fn call Wait()
func InvokeSafely(fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			pv := recover()
			if pv != nil {
				slog.Error(fmt.Sprintf("%v", pv))
			}
		}()
		fn()
	}()
}
