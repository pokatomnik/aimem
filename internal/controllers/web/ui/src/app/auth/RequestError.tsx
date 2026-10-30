import type { ReactNode } from "react";

import { useLogout } from "@/app/auth/useLogout";
import { ErrorMessage } from "@/shared/ui/ErrorMessage";

export type RequestErrorProps = {
  readonly error: unknown;
  readonly onRetry?: () => void;
  readonly title?: string;
};

/**
 * `ErrorMessage` for a failed request: the same layout, plus a "Logout" button
 * when the server rejected the token.
 *
 * The policy lives here, once, so no page has to ask "was this a 401?". A page
 * passes its error and the decision follows; a failure that has nothing to do
 * with credentials still gets only "Retry".
 */
export const RequestError = ({
  error,
  onRetry,
  title,
}: RequestErrorProps): ReactNode => {
  const onLogout = useLogout(error);

  return (
    <ErrorMessage error={error} onRetry={onRetry} onLogout={onLogout} title={title} />
  );
};
