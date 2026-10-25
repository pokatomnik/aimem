import { QueryClient } from "@tanstack/react-query";

/**
 * Shared defaults. Retries are off on purpose: a failed search already costs an
 * embedding pass plus a summarisation on the backend, and three automatic
 * retries would multiply that by four. The user gets an explicit "Retry".
 */
export const createQueryClient = (): QueryClient =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        refetchOnWindowFocus: false,
        staleTime: 30_000,
        gcTime: 5 * 60_000,
      },
    },
  });
