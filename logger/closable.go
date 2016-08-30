package logger

// Internal interface for hooks that must be closed (e.g. file hook)
type closable interface {
	Close() error
}
