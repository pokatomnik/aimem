package server

import (
	"cmp"
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

const defaultShutdownTimeout = 5 * time.Second

type (
	listener interface {
		ListenAndServe() error
		Shutdown(ctx context.Context) error
	}

	HTTPServerExt struct {
		server listener
	}

	ListenAndServeOptions struct {
		ShutdownTimeout time.Duration
	}
)

func NewHTTPServerExt(orig *http.Server) *HTTPServerExt {
	return &HTTPServerExt{server: orig}
}

func (s *HTTPServerExt) ListenAndServeContext(ctx context.Context, options *ListenAndServeOptions) error {
	shutdownTimeout := cmp.Or(options.ShutdownTimeout, defaultShutdownTimeout)
	eg, ctx := errgroup.WithContext(ctx)
	lch := make(chan error, 1)
	sch := make(chan error, 1)

	eg.Go(func() error {
		err := s.server.ListenAndServe()
		lch <- err
		return err
	})

	eg.Go(func() error {
		<-ctx.Done()
		stopCtx, cancel := context.WithTimeout(
			context.Background(),
			shutdownTimeout,
		)
		defer cancel()

		err := s.server.Shutdown(stopCtx)
		sch <- err
		return err
	})

	eg.Wait()

	return errors.Join(<-lch, <-sch)
}
