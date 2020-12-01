package stats

import (
	"bufio"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"time"
)

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
)

// EnableMemoryStatistics enables go routine that periodically prints memory
// usage of the go process.
func EnableMemoryStatistics(ctx context.Context, interval time.Duration) {

	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				PrintMemoryStatistics()
				PrintNumOfRoutines()
			case <-ctx.Done():
				err := DumpPrometheusDefaults()
				if err != nil {
					fmt.Println(err)
				}
				return
			}
		}
	}()
}

// toGigabytes returns given memory in bytes to gigabytes.
func toGigabytes(bytes uint64) float64 {
	return float64(bytes) / GIGABYTE
}

// PrintMemoryStatistics prints memory statistics using go runtime library.
func PrintMemoryStatistics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	bytesTotalAllocated := memStats.TotalAlloc
	bytesHeapAllocated := memStats.HeapAlloc
	countMalloc := memStats.Mallocs
	countFrees := memStats.Frees

	log.Infof(
		"Total allocated: %.3fGB, Heap allocated: %.3fGB, "+
			"Allocated objects count: %v, Freed objects count: %v",
		toGigabytes(bytesTotalAllocated),
		toGigabytes(bytesHeapAllocated),
		countMalloc,
		countFrees,
	)
}

// DumpPrometheusDefaults write default Prometheus metrics to a file
func DumpPrometheusDefaults() error {
	file, err := os.OpenFile(
		"stats",
		os.O_APPEND|os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)

	metricFamily, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return err
	}
	for _, v := range metricFamily {
		_, err := writer.WriteString(v.String() + "\n")
		if err != nil {
			return err
		}
	}

	writer.Flush()
	file.Close()

	return nil
}

// PrintNumOfRoutines prints number of go routines currently running
func PrintNumOfRoutines() {
	log.Infof("Num of go routines: %v\n", runtime.NumGoroutine())
}
