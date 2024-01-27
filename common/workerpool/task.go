package workerpool

import (
	"github.com/paulschick/disclosureupdater/common/logger"
	"go.uber.org/zap"
)

// Task
// source reference https://github.com/Joker666/goworkerpool/blob/main/workerpool/task.go
type Task struct {
	ID   int
	Err  error
	Data interface{}
	f    func(interface{}) error
}

func NewTask(f func(interface{}) error, data interface{}, ID int) *Task {
	return &Task{
		ID:   ID,
		f:    f,
		Data: data,
	}
}

func Process(workerID int, t *Task) {
	logger.Logger.Info("Worker started",
		zap.Int("workerID", workerID),
		zap.Int("taskID", t.ID))
	t.Err = t.f(t.Data)
}
