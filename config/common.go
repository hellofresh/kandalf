package config

import "time"

var (
	// The value of pause time to prevent CPU overload
	InfiniteCycleTimeout time.Duration = 2 * time.Second
)
