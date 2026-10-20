package dump

import (
	"context"
	"io"
)

type (
	dumper interface {
		Dump(ctx context.Context, to io.Writer) error
	}

	Dump struct {
		d dumper
	}
)

func New(d dumper) *Dump {
	return &Dump{d}
}

func (d *Dump) Dump(ctx context.Context, to io.Writer) error {
	return d.d.Dump(ctx, to)
}
