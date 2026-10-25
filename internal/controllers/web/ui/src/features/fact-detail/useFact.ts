import { useQuery } from "@tanstack/react-query";

import { useApi } from "@/app/providers/useApi";
import { factKeys } from "@/entities/fact/api/keys";
import type { Fact } from "@/entities/fact/api/keys";

export type FactState = {
  readonly fact: Fact | undefined;
  readonly error: unknown;
  readonly isPending: boolean;
  readonly refetch: () => void;
};

export const useFact = (id: string): FactState => {
  const api = useApi();

  const result = useQuery({
    queryKey: factKeys.detail(id),
    queryFn: async ({ signal }) => api.getFact(id, { signal }),
  });

  return {
    fact: result.data,
    error: result.error,
    isPending: result.isPending,
    refetch: () => {
      void result.refetch();
    },
  };
};
