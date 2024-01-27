package workerpool

import (
	"github.com/paulschick/disclosureupdater/common/logger"
	"go.uber.org/zap"
	"sync"
)

// Worker
// source reference https://github.com/Joker666/goworkerpool/blob/main/workerpool/worker.go
type Worker struct {
	ID       int
	taskChan chan *Task
}

func NewWorker(channel chan *Task, ID int) *Worker {
	return &Worker{
		ID:       ID,
		taskChan: channel,
	}
}

func (wr *Worker) Start(wg *sync.WaitGroup) {
	logger.Logger.Info("Worker started", zap.Int("workerID", wr.ID))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range wr.taskChan {
			Process(wr.ID, task)
		}
	}()
}
