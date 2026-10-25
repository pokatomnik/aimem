import type { Fact, FactListItem } from "@/shared/api";
import { normalizeQuery } from "@/shared/lib/text";

/**
 * Query keys. The normalised query is part of the key so that `' foo'` and
 * `'foo'` share one cache entry instead of diverging into two requests.
 */
export const factKeys = {
  all: ["fact"] as const,
  searchRoot: ["fact", "search"] as const,
  search: (query: string) => ["fact", "search", normalizeQuery(query)] as const,
  detailRoot: ["fact", "detail"] as const,
  detail: (id: string) => ["fact", "detail", id] as const,
};

export type { Fact, FactListItem };
