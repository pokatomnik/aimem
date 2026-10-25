import { useMutation, useQueryClient } from "@tanstack/react-query";

import { useApi } from "@/app/providers/useApi";
import { factKeys } from "@/entities/fact/api/keys";

export type DeleteFactState = {
  readonly error: unknown;
  readonly isPending: boolean;
  readonly remove: () => void;
};

/**
 * Deletes a fact and drops every cached copy of it, so navigating back to the
 * detail page cannot resurrect the removed entry from the cache.
 */
export const useDeleteFact = (
  id: string,
  onDeleted: () => void,
): DeleteFactState => {
  const api = useApi();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async () => api.deleteFact(id),
    onSuccess: async () => {
      queryClient.removeQueries({ queryKey: factKeys.detail(id) });
      await queryClient.invalidateQueries({ queryKey: factKeys.searchRoot });
      onDeleted();
    },
  });

  return {
    error: mutation.error,
    isPending: mutation.isPending,
    remove: () => {
      mutation.mutate();
    },
  };
};
