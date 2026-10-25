import { useMutation, useQueryClient } from "@tanstack/react-query";

import { useApi } from "@/app/providers/useApi";
import { factKeys } from "@/entities/fact/api/keys";

export type AddFactState = {
  /** Summary produced by the backend; present only after a successful add. */
  readonly summary: string | undefined;
  readonly error: unknown;
  readonly isPending: boolean;
  readonly submit: (fact: string) => void;
  readonly reset: () => void;
};

/**
 * `reset()` puts the mutation back to idle, which is how the form returns to
 * its initial state after "Add another".
 */
export const useAddFact = (): AddFactState => {
  const api = useApi();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (fact: string) => api.addFact(fact),
    onSuccess: async () => {
      // Existing search results are now stale: the new fact may belong to them.
      await queryClient.invalidateQueries({ queryKey: factKeys.searchRoot });
    },
  });

  return {
    summary: mutation.data,
    error: mutation.error,
    isPending: mutation.isPending,
    submit: (fact) => {
      mutation.mutate(fact);
    },
    reset: () => {
      mutation.reset();
    },
  };
};
