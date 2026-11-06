package embeder

import (
	"sync"

	"github.com/pokatomnik/gomem/internal/adapters/embeder/engine"
)

type (
	Embedder struct {
		encoder     *engine.Encoder
		initialized bool
		mx          sync.Mutex
	}
)

func New() *Embedder {
	return &Embedder{}
}
