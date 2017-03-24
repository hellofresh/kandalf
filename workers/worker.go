package workers

import (
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/stats-go"
)

type Worker struct {
	die          chan bool
	wg           *sync.WaitGroup
	statsClient  stats.StatsClient
	isWorking    bool
	pipes        []config.Pipe
	globalConfig config.GlobalConfig
}

type internalWorker interface {
	run(*sync.WaitGroup, chan bool)
}

// Returns new instance of worker
func NewWorker(globalConfig config.GlobalConfig, pipes []config.Pipe, statsClient stats.StatsClient) *Worker {
	return &Worker{
		die:          make(chan bool, 1),
		wg:           &sync.WaitGroup{},
		statsClient:  statsClient,
		pipes:        pipes,
		globalConfig: globalConfig,
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
			// Prevent CPU overload
			time.Sleep(w.globalConfig.InfiniteCycleTimeout.Duration)
		}
	}
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
			log.WithError(err).Error("Unable to get list of the workers")

			return
		}

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
	)

	queue, err = newInternalQueue(w.globalConfig, w.statsClient)
	if err != nil {
		return nil, fmt.Errorf("An error occured while instantiating queue: %v", err)
	}

	for _, pipe := range w.pipes {
		consumer, err = newInternalConsumer(w.globalConfig.RabbitDSN, queue, pipe, w.globalConfig.InfiniteCycleTimeout.Duration)
		if err != nil {
			log.WithError(err).Warning("Unable to create consumer")
		} else {
			workers = append(workers, consumer)

			log.WithError(err).Debug("Created a new consumer")
		}
	}

	if len(workers) == 0 {
		return nil, errors.New("Haven't found any consumer or all of them failed to connect")
	}

	workers = append(workers, queue)

	return workers, nil
}
