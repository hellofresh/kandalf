package runnable

import (
	"sync"
	"time"

	"kandalf/config"
)

type cbDoRun func(*sync.WaitGroup, chan bool)

type Runnable interface {
	Run(*sync.WaitGroup, chan bool)
	Reload()
}

type RunnableWorker struct {
	ChReload  chan bool
	IsWorking bool
	callback  cbDoRun
}

// Returns runnable worker instance
func NewRunnableWorker(callback cbDoRun) *RunnableWorker {
	return &RunnableWorker{
		ChReload:  make(chan bool),
		IsWorking: false,
		callback:  callback,
	}
}

// Main working cycle
func (w *RunnableWorker) Run(wgMain *sync.WaitGroup, dieMain chan bool) {
	defer wgMain.Done()

	w.IsWorking = true

	go w.callback(wgMain, dieMain)

	for {
		select {
		case <-dieMain:
			w.IsWorking = false
			return
		default:
		}

		// Prevent CPU overload
		time.Sleep(config.InfiniteCycleTimeout)
	}
}

// Reloads the worker
func (w *RunnableWorker) Reload() {
	w.ChReload <- true
}
