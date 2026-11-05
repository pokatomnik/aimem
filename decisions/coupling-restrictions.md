# Coupling Restrictions

## Decision

Layers meet through interfaces, never through implementations. A package depends
on the behaviour it needs, declared as a narrow unexported interface next to the
consumer (`type embedder interface`, `type collection interface`), and receives a
concrete value injected from above. It never names a concrete type from another
layer — not in a signature, not in a struct field, not in a `new(...)` call, not
in a type assertion.

`cmd/aimem/scaffold.go` is the only composition root and the only place that
imports implementations and composes them. Nothing else may do that work.

These are bugs, fixed at once:

- `internal/controllers/**` importing `internal/adapters/**`, or the reverse.
- One adapter importing another, or one controller importing another, instead of
  both depending on an interface in `internal/entities` or on a locally declared
  one.
- A use case importing a repository package. Use cases declare the interfaces
  they need (`internal/use_cases/recall` declares `ai` and its store), and
  `scaffold` satisfies them.
- A package returning a concrete type from another layer, or accepting one as an
  argument, where an interface would do.
- A `WithAuth(bool)`-style parameter threaded so that the composition root does
  not have to decide. The decision belongs in `scaffold`, expressed as two
  constructors — see `authentication-mode-selection.md`.

Exceptions are narrow and always about data, not behaviour: shared value types in
`internal/entities/**` (`fact.FactFullView`, `user.User`, the `apperrors`
sentinels) may be imported everywhere, `internal/shared/**` helpers are importable
everywhere by design, and `main` in `cmd/aimem` may import anything because it is
the composition root's entry. A package may import its own subpackages
(`servers/mcp` → `servers/mcp/tools`); that is not a cross-layer edge.

When in doubt, ask which side owns the behaviour. If the answer is "both", the
interface belongs in `internal/entities` or `internal/shared`; if it is "the
caller", declare it next to the caller and inject.

## Why

Every violation compiles and passes the tests, so a review catches nothing and a
green build is no evidence at all. The shortcut is also always the shortest path
to a working build — a controller importing a repository "just for the type", an
adapter importing a controller "just for a constructor", two adapters sharing a
helper — which is why an agent is more likely to take it than a human.

Taking it welds two layers together invisibly, and the cost shows up much later:
the package can no longer be tested without the other layer, and a change to one
side silently changes what the other can do. Injecting a narrow interface instead
keeps both sides replaceable and leaves the one decision in one file.
