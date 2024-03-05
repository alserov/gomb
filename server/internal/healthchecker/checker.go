package healthchecker

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

func PrintStat(log *slog.Logger) {
	t := time.NewTicker(time.Minute)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	for range t.C {
		log.Info(fmt.Sprintf("Goroutines running: %d | CPUs usable: %d | TotalAlloc: %d MiB | Alloc: %d MiB | Bytes in stack: %d MiB | NumGC: %d",
			runtime.NumGoroutine(), runtime.NumCPU(), m.TotalAlloc/1024/1024, m.Alloc/1024/1024, m.StackInuse/1024/1024, m.NumGC))
	}
}
