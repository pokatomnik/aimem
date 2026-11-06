package embeder

import (
	"sync"

	"github.com/pokatomnik/gomem/internal/adapters/embeder/engine"
	"github.com/pokatomnik/gomem/internal/shared/logging"
)

type (
	Embedder struct {
		encoder     *engine.Encoder
		initialized bool
		mx          sync.Mutex
		logger      logging.Logger
	}
)

func New(logger logging.Logger) *Embedder {
	return &Embedder{logger: logger}
}
