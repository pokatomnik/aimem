# The Auth Flag Lives in the Served HTML

## Decision

The token prompt is baked into the served HTML instead of being exposed as an
endpoint. `newSPAHandler` takes `authRequired`, and `injectAuthFlag` writes
`<script>window.requireAuth=…</script>` into `index.html`'s head, before the entry
module runs. Only the entry document is rewritten; hashed assets are served
untouched.

There is no `GET /api/v1/auth` endpoint. The token lives in `localStorage` under
`aimem.token`. Logging out is client-side only: the error block drops the token and
shows the form again.

## Why

The answer comes from the same `authRequired` value the handlers enforce, so the
page and the API cannot disagree. An endpoint would let the two drift, and the
server already knows the answer when it builds the page.

Rewriting only the entry document keeps immutable caching honest: a hashed asset is
unchanged byte for byte, so serving it untouched is not a lie to the browser.

Logout stays client-side because the credential is a static token from
`AIMEM_USERS_JSON`, parsed once and cached. Revoking it for real means removing it
from that list and restarting the process; a logout endpoint would imply a
server-side session that this design does not have.
