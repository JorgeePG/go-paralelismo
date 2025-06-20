package worker

import (
	"strings"
	"sync"
)

type Worker struct{}

func (w *Worker) Process(lines []string) int {
	var totalWords int
	var wg sync.WaitGroup
	mu := &sync.Mutex{}

	for _, line := range lines {
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			wordCount := len(strings.Fields(l))
			mu.Lock()
			totalWords += wordCount
			mu.Unlock()
		}(line)
	}

	wg.Wait()
	return totalWords
}