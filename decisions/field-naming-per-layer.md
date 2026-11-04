# Field Naming Per Layer

## Decision

One value has three names, and each layer uses its own:

| Layer                                                                                | Spelling     | Example                     |
| ------------------------------------------------------------------------------------ | ------------ | --------------------------- |
| Go code                                                                              | `CamelCase`  | `CreatedAt`, `FactFullView` |
| External contract — `openapi.yaml`, generated server and UI client, MCP tool schemas | `camelCase`  | `createdAt`                 |
| Storage — `kv` fields and the Qdrant payload                                         | `snake_case` | `created_at`                |

The spelling changes **between** layers — in `methods.go`, which reads `CreatedAt`
from the domain type and fills the generated one, and in the repositories, which
read `created_at` out of a payload — never within one layer.

Both generated sides of the contract come from one `openapi.yaml`, so they are
regenerated together, always:

```sh
make api-generate   # ogen → server/  and  openapi-typescript → schema.d.ts
```

Editing either generated tree by hand, or running only one generator, is the bug
this rule exists to prevent.

## Why

The contract is what a client sees, so it follows the convention every JSON API
uses. The storage schema is a different audience — data already written to
`~/.aimem` and to Qdrant — so it keeps `snake_case` and stays readable without a
migration.

The two generated sides can only agree if they are produced from one spec in one
step. A single-generator run leaves the UI typed against shapes the server no
longer returns, and the mismatch surfaces much later, as a `tsc` failure on a
stale `schema.d.ts`.
