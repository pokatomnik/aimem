import type { ReactNode } from "react";

import { ApiError } from "@/shared/api/ApiError";
import { getErrorMessage } from "@/shared/api/errorMessages";
import { Button } from "@/shared/ui/Button";
import styles from "./ErrorMessage.module.css";

type ErrorMessageProps = {
  readonly error: unknown;
  readonly onRetry?: () => void;
  readonly title?: string;
};

/**
 * Turns any thrown value into a readable message. The mapping lives in
 * `errorMessages.ts`, so this component only decides the layout.
 */
export const ErrorMessage = ({
  error,
  onRetry,
  title = "Request failed",
}: ErrorMessageProps): ReactNode => {
  const message =
    error instanceof ApiError ? error.message : getErrorMessage("UNKNOWN");

  return (
    <div className={styles.wrap} role="alert">
      <p className={styles.title}>{title}</p>
      <p className={styles.text}>{message}</p>
      {onRetry !== undefined ? (
        <div className={styles.actions}>
          <Button variant="secondary" onClick={onRetry}>
            Retry
          </Button>
        </div>
      ) : null}
    </div>
  );
};
