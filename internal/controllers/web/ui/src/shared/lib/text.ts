/**
 * Search queries shorter than this are rejected by the backend (`minLength: 3`),
 * so the client refuses to send them at all. The backend may relax this rule in
 * the future, never tighten it below three characters.
 */
export const SEARCH_QUERY_MIN_LENGTH = 3;

/** A fact must contain at least one non-whitespace character. */
export const FACT_MIN_LENGTH = 1;

/** Canonical form of user input: trimmed, ready to be used as a cache key or a request body. */
export const normalizeQuery = (value: string): string => value.trim();

export const isSearchQueryValid = (value: string): boolean =>
  normalizeQuery(value).length >= SEARCH_QUERY_MIN_LENGTH;

export const isFactValid = (value: string): boolean =>
  normalizeQuery(value).length >= FACT_MIN_LENGTH;

/** "1 result", "2 results" — the UI is English-only, so no plural forms. */
export const formatResultCount = (count: number): string =>
  count === 1 ? "1 result" : `${String(count)} results`;
