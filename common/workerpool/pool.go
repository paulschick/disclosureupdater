package workerpool

import "sync"

type Pool struct {
	Tasks       []*Task
	concurrency int
	collector   chan *Task
	wg          sync.WaitGroup
}

// NewPool creates a new pool
func NewPool(tasks []*Task, concurrency, totalTasks int) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		collector:   make(chan *Task, totalTasks),
	}
}

func (p *Pool) Run() {
	for i := 1; i < p.concurrency; i++ {
		worker := NewWorker(p.collector, i)
		worker.Start(&p.wg)
	}
	for j := range p.Tasks {
		p.collector <- p.Tasks[j]
	}
	close(p.collector)
	p.wg.Wait()
}
