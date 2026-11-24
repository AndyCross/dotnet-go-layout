// runner starts a command n times, one after another, and reports what the
// kernel knows about each run: wall time from spawn to exit, and peak RSS. It
// is built once as a static Linux binary and mounted into every image, so the
// Go and .NET builds of chalk are measured by the same few lines of Go.
//
//	runner <n> <command> [args...]
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"syscall"
	"time"
)

type report struct {
	Runs        int     `json:"runs"`
	MedianMs    float64 `json:"median_ms"`
	MinMs       float64 `json:"min_ms"`
	P90Ms       float64 `json:"p90_ms"`
	MedianRSSKB int64   `json:"median_rss_kb"`
	Stdout      string  `json:"stdout"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: runner <n> <command> [args...]")
		os.Exit(2)
	}
	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 1 {
		fmt.Fprintln(os.Stderr, "runner: n must be a positive number")
		os.Exit(2)
	}

	// Five runs first that don't count, so the binary and its libraries are
	// in the page cache for every run that does, as they would be on a
	// machine that runs it more than once.
	var first string
	for range 5 {
		first, _, _ = once()
	}

	walls := make([]float64, 0, n)
	rss := make([]int64, 0, n)
	for range n {
		out, ms, kb := once()
		if out != first {
			fmt.Fprintln(os.Stderr, "runner: output changed between runs")
			os.Exit(1)
		}
		walls = append(walls, ms)
		rss = append(rss, kb)
	}
	slices.Sort(walls)
	slices.Sort(rss)
	json.NewEncoder(os.Stdout).Encode(report{
		Runs:        n,
		MedianMs:    walls[n/2],
		MinMs:       walls[0],
		P90Ms:       walls[n*9/10],
		MedianRSSKB: rss[n/2],
		Stdout:      first,
	})
}

func once() (stdout string, wallMs float64, maxRSSKB int64) {
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	start := time.Now()
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "runner:", err)
		os.Exit(1)
	}
	wall := time.Since(start)
	if ru, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
		maxRSSKB = ru.Maxrss // kilobytes on Linux
	}
	return out.String(), float64(wall.Microseconds()) / 1000, maxRSSKB
}
