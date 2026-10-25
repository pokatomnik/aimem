import { createPath, createSearchParams } from 'react-router';

/** Route templates. `fact` is a pattern; `factPath` builds a concrete URL. */
export const ROUTES = {
  home: '/',
  search: '/search',
  factAdd: '/fact/add',
  fact: '/fact/:factID',
} as const;

export const factPath = (id: string): string => `/fact/${encodeURIComponent(id)}`;

/**
 * Builds `/search?q=...` in the exact shape the task requires: a space becomes
 * `+` and every other reserved character is percent-encoded by `URLSearchParams`.
 */
export const buildSearchPath = (query: string): string =>
  createPath({
    pathname: ROUTES.search,
    search: createSearchParams({ q: query }).toString(),
  });
