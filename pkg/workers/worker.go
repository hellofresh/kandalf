package workers

// Worker is public interface for all worker services
type Worker interface {
	// Execute runs the service logic once in sync way
	Execute()
	// Go runs the service forever in async way in go-routine
	Go(interrupt chan bool)
	// Close closes worker resources
	Close() error
}
