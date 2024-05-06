package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func WalkDir(dir string, fileSize chan<- int64) {
	for _, entry := range dirents(dir) {
		if entry.IsDir() {
			subdir := filepath.Join(dir, entry.Name())
			WalkDir(subdir, fileSize)
		} else {
			fileInfo, err := entry.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "walkdir: %v\n", err)
			}
			fileSize <- fileInfo.Size()
		}

	}
}

func dirents(dir string) []os.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "walkdir: %v\n", err)
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
	go func() {
		for _, root := range roots {
			WalkDir(root, fileSizes)
		}
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
