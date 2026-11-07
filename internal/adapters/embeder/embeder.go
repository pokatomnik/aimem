package embeder

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/pokatomnik/gomem/internal/adapters/embeder/engine"
	"github.com/pokatomnik/gomem/internal/shared/logging"
)

type (
	Embedder struct {
		encoder     *engine.Encoder
		initialized bool
		encoderMx   sync.Mutex
		lastUsed    *time.Time

		logger logging.Logger
	}
)

func New(ctx context.Context, logger logging.Logger) *Embedder {
	embedder := &Embedder{logger: logger}

	go func() {
		for {
			time.Sleep(periodicCheck)

			if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return
			}

			embedder.encoderMx.Lock()

			now := time.Now()
			wasLoaded := embedder.encoder != nil

			if embedder.lastUsed != nil && now.Sub(*embedder.lastUsed) >= ttl {
				embedder.unloadEncoder()
			}

			nowUnloaded := embedder.encoder == nil

			embedder.encoderMx.Unlock()

			if wasLoaded && nowUnloaded {
				runtime.GC()
			}
		}
	}()

	return embedder
}
