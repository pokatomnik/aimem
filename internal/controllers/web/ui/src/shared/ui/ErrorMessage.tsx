import type { ReactNode } from "react";

import { ApiError } from "@/shared/api/ApiError";
import { getErrorMessage } from "@/shared/api/errorMessages";
import { Button } from "@/shared/ui/Button";
import styles from "./ErrorMessage.module.css";

type ErrorMessageProps = {
  readonly error: unknown;
  readonly onRetry?: (() => void) | undefined;
  /**
   * Escape hatch from a rejected credential. Passed in rather than derived here:
   * this component is shared UI and knows nothing about sessions, so the caller
   * decides when logging out makes sense.
   */
  readonly onLogout?: (() => void) | undefined;
  readonly title?: string | undefined;
};

/**
 * Turns any thrown value into a readable message. The mapping lives in
 * `errorMessages.ts`, so this component only decides the layout.
 */
export const ErrorMessage = ({
  error,
  onRetry,
  onLogout,
  title = "Request failed",
}: ErrorMessageProps): ReactNode => {
  const message =
    error instanceof ApiError ? error.message : getErrorMessage("UNKNOWN");
  const hasActions = onRetry !== undefined || onLogout !== undefined;

  return (
    <div className={styles.wrap} role="alert">
      <p className={styles.title}>{title}</p>
      <p className={styles.text}>{message}</p>
      {hasActions ? (
        <div className={styles.actions}>
          {onRetry !== undefined ? (
            <Button variant="secondary" onClick={onRetry}>
              Retry
            </Button>
          ) : null}
          {onLogout !== undefined ? (
            <Button variant="ghost" onClick={onLogout}>
              Logout
            </Button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
};
