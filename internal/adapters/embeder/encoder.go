package embeder

import "github.com/pokatomnik/gomem/internal/adapters/embeder/engine"

func (e *Embedder) getEncoder() (*engine.Encoder, error) {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.initialized {
		return e.encoder, nil
	}

	encoder, err := engine.LoadEncoder(gemmaDir, gemmaDirName)
	if err != nil {
		e.logger.Error("load encoder failed", "checkpoint", gemmaDirName, "error", err)

		return nil, err
	}

	e.initialized = true
	e.encoder = encoder

	return encoder, nil
}
