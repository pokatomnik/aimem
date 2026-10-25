import { createPath, createSearchParams } from "react-router";

import { normalizeQuery } from "@/shared/lib/text";

/** Route templates. `fact` is a pattern; `factPath` builds a concrete URL. */
export const ROUTES = {
  home: "/",
  search: "/search",
  factAdd: "/fact/add",
  fact: "/fact/:factID",
} as const;

export const factPath = (id: string): string =>
  `/fact/${encodeURIComponent(id)}`;

/**
 * Builds `/search?q=...` in the exact shape the task requires: a space becomes
 * `+` and every other reserved character is percent-encoded by `URLSearchParams`.
 */
export const buildSearchPath = (query: string): string =>
  createPath({
    pathname: ROUTES.search,
    search: createSearchParams({ q: query }).toString(),
  });

/** The parts of a router location this module needs. */
export type SearchLocation = {
  readonly pathname: string;
  readonly search: string;
};

/** Reads `?q=` and reports `''` when it is absent. */
export const readSearchQuery = (search: string): string =>
  new URLSearchParams(search).get("q") ?? "";

/**
 * Tells whether `query` is the search the page is already showing.
 *
 * Navigating there again would push a second entry with an identical URL onto
 * the history stack, so Back would appear to do nothing. Any other route is not
 * a search page, which is why going to a search from `/` or `/fact/add` must
 * always navigate.
 */
export const isSameSearch = (
  location: SearchLocation,
  query: string,
): boolean => {
  if (location.pathname !== ROUTES.search) {
    return false;
  }

  return (
    normalizeQuery(readSearchQuery(location.search)) === normalizeQuery(query)
  );
};
