package core

import (
	"context"
	"sync"
)

type WorkerPool struct {
	runner     *Runner
	queue      *Queue
	concurrent int
	cancel     context.CancelFunc
	mu         sync.Mutex
	running    bool
	OnProgress func(Progress)
}

func NewWorkerPool(runner *Runner, queue *Queue, concurrent int) *WorkerPool {
	if concurrent < 1 {
		concurrent = 1
	}
	return &WorkerPool{
		runner:     runner,
		queue:      queue,
		concurrent: concurrent,
	}
}

func (wp *WorkerPool) Start() {
	wp.mu.Lock()
	if wp.running {
		wp.mu.Unlock()
		return
	}
	wp.running = true
	ctx, cancel := context.WithCancel(context.Background())
	wp.cancel = cancel
	wp.mu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < wp.concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				item := wp.queue.NextPending()
				if item == nil {
					return
				}

				presets := DefaultPresets()
				var preset *Preset
				for i := range presets {
					if presets[i].Name == item.PresetName {
						preset = &presets[i]
						break
					}
				}
				if preset == nil {
					preset = &presets[0]
				}

				err := wp.runner.Run(item.InputPath, item.OutputPath, *preset, func(p Progress) {
					p.File = item.InputPath
					if wp.OnProgress != nil {
						wp.OnProgress(p)
					}
				})

				wp.queue.mu.Lock()
				for idx := range wp.queue.Items {
					if &wp.queue.Items[idx] == item {
						if err != nil {
							wp.queue.Items[idx].Status = "failed"
							wp.queue.Items[idx].Error = err.Error()
						} else {
							wp.queue.Items[idx].Status = "done"
							wp.queue.Items[idx].Progress = 100
						}
						break
					}
				}
				wp.queue.mu.Unlock()

				if wp.OnProgress != nil {
					wp.OnProgress(Progress{File: item.InputPath, Percent: 100, Status: "done"})
				}
			}
		}()
	}
	wg.Wait()

	wp.mu.Lock()
	wp.running = false
	wp.mu.Unlock()
}

func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	if wp.cancel != nil {
		wp.cancel()
	}
	wp.running = false
}

func (wp *WorkerPool) IsRunning() bool {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.running
}
