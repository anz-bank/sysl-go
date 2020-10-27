package core

// StoppableServer offers control over the lifecycle of a server that
// can be started at most once and then stopped.
type StoppableServer interface {
	// Start() directs the server to attempt to start serving. This call will block
	// until a failure occurs, or the server is stopped. If the server is not stopped
	// and no failure occurs, this call may block indefinitely.
	// On a successful stop() or gracefulstop(), no error is returned.
	Start() error

	// Stop() directs the server to immediately and un-gracefully stop, which may
	// interrupt any in-flight requests/calls being processed.
	Stop() error

	// GracefulStop() directs the server to stop accepting new requests/calls and
	// wait for any in-flight requests/calls to complete. Implementations of
	// GracefulStop() should ensure they bring the server to a stop after some
	// bounded period of time, even if this involves ungracefully dropping some
	// laggard requests/calls.
	GracefulStop() error
}
