package runnable

import "sync"

type Runnable interface {
	Run(*sync.WaitGroup, chan bool)
	Reload()
}
