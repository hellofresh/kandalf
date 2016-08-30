package workers

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"kandalf/config"
	"kandalf/logger"
	"kandalf/runnable"
)

type Worker struct {
	*runnable.RunnableWorker
}

type internalWorker interface {
	run(*sync.WaitGroup, chan bool)
}

// Returns new instance of worker
func NewWorker() *Worker {
	w := &Worker{}
	w.RunnableWorker = runnable.NewRunnableWorker(w.doRun)

	return w
}

// Launches the internal workers and executes them infinitely
func (w *Worker) doRun(wgMain *sync.WaitGroup, dieMain chan bool) {
	var (
		die     chan bool
		err     error
		wg      *sync.WaitGroup
		workers []internalWorker
	)

	for w.RunnableWorker.IsWorking {
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
				case <-w.RunnableWorker.ChReload:
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
		c *internalConsumer
		q *internalQueue
	)

	q, err = newInternalQueue()
	if err != nil {
		return nil, fmt.Errorf("An error occured while instantiating queue: %v", err)
	}

	for _, url := range config.Instance().UList("rabbitmq.urls") {
		c, err = newInternalConsumer(url.(string), q)
		if err != nil {
			logger.Instance().
				WithError(err).
				WithField("url", url).
				Warning("Unable to create consumer")
		} else {
			workers = append(workers, c)

			logger.Instance().
				WithError(err).
				WithField("url", url).
				Debug("Created a new consumer")
		}
	}

	if len(workers) == 0 {
		return nil, errors.New("Haven't found any consumer or all of them failed to connect")
	}

	workers = append(workers, q)

	return workers, nil
}
