package workers

import "sync"

type Worker interface {
	Run(*sync.WaitGroup, chan bool)
}
