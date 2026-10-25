import { useQuery } from "@tanstack/react-query";

import { useApi } from "@/app/providers/useApi";
import { factKeys } from "@/entities/fact/api/keys";
import type { FactListItem } from "@/entities/fact/api/keys";
import { isSearchQueryValid, normalizeQuery } from "@/shared/lib/text";

export type SearchState = {
  readonly items: FactListItem[] | undefined;
  readonly error: unknown;
  readonly isPending: boolean;
  readonly isFetching: boolean;
  readonly refetch: () => void;
};

/**
 * Semantic search. Disabled for queries the backend would reject, so a short
 * query never turns into a guaranteed 400.
 */
export const useSearch = (query: string): SearchState => {
  const api = useApi();
  const normalized = normalizeQuery(query);
  const enabled = isSearchQueryValid(normalized);

  const result = useQuery({
    queryKey: factKeys.search(normalized),
    queryFn: async ({ signal }) => api.search(normalized, { signal }),
    enabled,
  });

  return {
    items: result.data,
    error: result.error,
    isPending: result.isPending,
    isFetching: result.isFetching,
    refetch: () => {
      void result.refetch();
    },
  };
};
