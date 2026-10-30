import { useQueryClient } from "@tanstack/react-query";

import { useAuth } from "@/app/auth/AuthContext";
import { ApiError } from "@/shared/api/ApiError";

/**
 * Whether the failure is a rejected credential rather than a broken request.
 * Only then does offering "Logout" help: the token is stale, and dropping it
 * returns the user to the form where a fresh one can be entered.
 */
export const isUnauthorized = (error: unknown): boolean =>
  error instanceof ApiError && error.code === "UNAUTHORIZED";

/**
 * Logout offered next to a failed request, or `undefined` when it would change
 * nothing: with no auth the token is irrelevant, and with no token the gate is
 * already showing.
 *
 * Only a rejected credential (`UNAUTHORIZED`) earns the button. Every other
 * failure — Ollama down, a storage error — is answered by "Retry"; offering a
 * logout there would suggest the token is at fault when it is not.
 *
 * The cache is cleared on the way out because everything in it was fetched under
 * the identity being dropped: after a fresh token is entered, an answer cached
 * from the old one must not be shown.
 */
export const useLogout = (error: unknown): (() => void) | undefined => {
  const { required, token, clearToken } = useAuth();
  const queryClient = useQueryClient();

  if (!required || token === undefined || !isUnauthorized(error)) {
    return undefined;
  }

  return () => {
    clearToken();
    queryClient.clear();
  };
};
