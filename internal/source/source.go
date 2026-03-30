package source

import (
	"context"

	"github.com/pattynextdoor/toph/internal/data"
)

// Source is the pluggable interface for data providers.
// Each implementation (JSONL watcher, hook HTTP server, process scanner)
// emits normalized Events onto a shared channel.
type Source interface {
	// Name returns a human-readable identifier for this source.
	Name() string
	// Start begins emitting events. It should block until ctx is cancelled
	// or an unrecoverable error occurs.
	Start(ctx context.Context, events chan<- data.Event) error
	// Stop gracefully shuts down the source.
	Stop() error
}
