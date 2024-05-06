package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func WalkDir(dir string, fileSize chan<- int64, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, entry := range dirents(dir) {
		if entry.IsDir() {
			wg.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go WalkDir(subdir, fileSize, wg)
		} else {
			fileInfo, err := entry.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			fileSize <- fileInfo.Size()
		}

	}
}

func dirents(dir string) []os.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil
	}
	return entries
}

func PrintDiskUsage(nFiles, nBytes int64) {
	fmt.Printf("%d files, %.1f GB \n", nFiles, float64(nBytes)/1e9)
}

var verbose = flag.Bool("v", false, "show verbose progress")

func main() {
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	fileSizes := make(chan int64)

	var wg sync.WaitGroup
	for _, root := range roots {
		wg.Add(1)
		go WalkDir(root, fileSizes, &wg)
	}
	go func() {
		wg.Wait()
		close(fileSizes)
	}()

	var tick <-chan time.Time
	if *verbose {
		tick = time.Tick(500 * time.Millisecond)
	}

	var nFiles, nBytes int64
loop:
	for {
		select {
		case size, ok := <-fileSizes:
			if !ok {
				break loop
			}
			nFiles++
			nBytes += size
		case <-tick:
			PrintDiskUsage(nFiles, nBytes)
		}
		PrintDiskUsage(nFiles, nBytes)
	}

}
