package workers

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hellofresh/kandalf/config"
	"github.com/hellofresh/kandalf/logger"
	"github.com/hellofresh/kandalf/pipes"
)

type Worker struct {
	die       chan bool
	reload    chan bool
	wg        *sync.WaitGroup
	isWorking bool
}

type internalWorker interface {
	run(*sync.WaitGroup, chan bool)
}

// Returns new instance of worker
func NewWorker() *Worker {
	return &Worker{
		die:    make(chan bool, 1),
		reload: make(chan bool),
		wg:     &sync.WaitGroup{},
	}
}

// Main working cycle
func (w *Worker) Run(wgMain *sync.WaitGroup, dieMain chan bool) {
	defer wgMain.Done()

	w.isWorking = true

	go w.doRun()

	for {
		select {
		case <-dieMain:
			w.isWorking = false
			return
		default:
		}

		// Prevent CPU overload
		time.Sleep(config.InfiniteCycleTimeout)
	}
}

// Reloads the worker
func (w *Worker) Reload() {
	w.reload <- true
}

// Launches the internal workers and executes them infinitely
func (w *Worker) doRun() {
	var (
		die     chan bool
		err     error
		wg      *sync.WaitGroup
		workers []internalWorker
	)

	for w.isWorking {
		wg = &sync.WaitGroup{}
		die = make(chan bool)
		workers, err = w.getWorkers()

		if err != nil {
			logger.Instance().
				WithError(err).
				Error("Unable to get list of the workers")

			return
		}

		go func() {
			for {
				select {
				case <-w.reload:
					logger.Instance().Info("Caught reload signal. Will stop all workers")

					close(die)

					return
				default:
				}

				// Prevent CPU overload
				time.Sleep(config.InfiniteCycleTimeout)
			}
		}()

		wg.Add(len(workers))
		for _, w := range workers {
			go w.run(wg, die)
		}
		wg.Wait()
	}
}

// Returns list of the internal workers
func (w *Worker) getWorkers() (workers []internalWorker, err error) {
	var (
		consumer *internalConsumer
		queue    *internalQueue
		rmqUrl   string
	)

	queue, err = newInternalQueue()
	if err != nil {
		return nil, fmt.Errorf("An error occured while instantiating queue: %v", err)
	}

	rmqUrl, err = config.Instance().String("rabbitmq.url")
	if err != nil {
		return nil, fmt.Errorf("Unable to get RabbitMQ connection URL: %v", err)
	}

	for _, pipe := range pipes.All() {
		consumer, err = newInternalConsumer(rmqUrl, queue, pipe)
		if err != nil {
			logger.Instance().
				WithError(err).
				Warning("Unable to create consumer")
		} else {
			workers = append(workers, consumer)

			logger.Instance().
				WithError(err).
				Debug("Created a new consumer")
		}
	}

	if len(workers) == 0 {
		return nil, errors.New("Haven't found any consumer or all of them failed to connect")
	}

	workers = append(workers, queue)

	return workers, nil
}
