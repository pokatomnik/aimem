# aimem Web UI

SPA over the `aimem web` HTTP API. Sources build into `dist/`, which the Go
binary embeds with `go:embed`.

## Stack

React 19, TypeScript (strict), Vite, Vitest, React Router (`HashRouter`),
TanStack Query, `openapi-typescript` + `openapi-fetch` (transport: `fetch`).

## Commands

```sh
npm ci              # install from lockfile
npm run dev         # dev server on :5173, /api proxied to 127.0.0.1:8888
npm run build       # tsc --noEmit && vite build → dist/
npm run test        # vitest run (hermetic: no network, no filesystem)
npm run lint        # eslint, 0 warnings
npm run typecheck   # tsc --noEmit
npm run api:generate
```

`npm run api:generate` reads `../openapi.yaml` (the file stays where it is) and
writes types to `src/api/generated/schema.d.ts`. The generated file is committed,
so building does not depend on the code generator.

From the repository root: `make ui-build`, `make ui-test`, `make ui-lint`.
`make build` builds the UI first, then the Go binary — the order matters for
`go:embed`.

## Architecture

Feature-Sliced: `pages → features → entities → shared`, dependencies point down
only. Imports use the `@/` alias (see `tsconfig.json` `paths` and the Vite
`resolve.alias`), so a module is reached as `@/shared/api` from anywhere.

- `app/` — routing and composition: `routes.tsx`, `navigation/` (route table and
  the only `useNavigate` call), `providers/`, `layout/`, `styles/`.
- `shared/api` — the only place that knows about HTTP: the client, the
  `{value: …}` envelope unwrapping, `ApiError`, and error-code messages.
- `entities/fact` — domain types, cache keys, date formatting, `FactCard` and
  `MarkdownView`.
- `features/*` — request and mutation hooks plus composed widgets.
- `pages/*` — composition; the search query lives in the URL (`?q=`).

### Navigation is not a page concern

`useNavigate` exists in exactly one module, `app/navigation/useAppNavigation.ts`.
`routes.tsx` calls it once and passes plain callbacks and hrefs down as props, so
a page can only ask for a destination the app has already defined:

```tsx
<FactPage onBack={nav.goHome} homeHref={nav.homeHref} />
```

Pages may read route state (`useParams`, `useSearchParams`) and render `Link` for
plain destinations, but they never decide where the app goes.

### API calls are same-origin

There is no base URL anywhere: the client issues `fetch('/api/v1/…')` against the
current host, because the SPA and the API are always served by the same binary.
`createFactApi` accepts an optional `baseUrl` that only tests use — Node's
`Request` rejects relative URLs, the browser does not.

## Contract notes

- Search is `POST /api/v1/search` with a `{query}` body; `?q=` in the address is
  navigation state only.
- Queries shorter than 3 characters are never sent (client-side validation).
- `POST /api/v1/fact` returns no `id`, so after adding, the UI shows the
  `summary` instead of navigating to a fact page.
- `DELETE /api/v1/fact/{factID}` answers `200 { "value": {} }`; only the status
  matters on the client.
- `rehype-raw` is deliberately not installed: raw HTML in a fact stays escaped.
- All UI copy is in English. There is no i18n layer.
