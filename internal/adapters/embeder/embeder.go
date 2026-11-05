package embeder

import (
	"sync"

	"github.com/townsendmerino/goinfer/embeddinggemma2"
)

type (
	Embedder struct {
		modelDir    string
		encoder     *embeddinggemma2.Encoder
		initialized bool
		mx          sync.Mutex
	}
)

func New(modelDir string) *Embedder {
	return &Embedder{modelDir: modelDir}
}
