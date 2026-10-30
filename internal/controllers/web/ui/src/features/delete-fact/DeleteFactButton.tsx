import type { ReactNode } from "react";
import { useState } from "react";

import { RequestError } from "@/app/auth/RequestError";
import { Button } from "@/shared/ui/Button";
import styles from "./DeleteFactButton.module.css";
import { useDeleteFact } from "./useDeleteFact";

export type DeleteFactButtonProps = {
  readonly factId: string;
  readonly onDeleted: () => void;
};

/**
 * Two-step inline confirmation instead of `window.confirm`: no blocking modal,
 * no native dialog, and the whole flow is testable with Testing Library.
 */
export const DeleteFactButton = ({
  factId,
  onDeleted,
}: DeleteFactButtonProps): ReactNode => {
  const [isConfirming, setIsConfirming] = useState(false);
  const { remove, isPending, error } = useDeleteFact(factId, onDeleted);

  return (
    <div>
      <div className={styles.actions}>
        {isConfirming ? (
          <>
            <span className={styles.confirmText}>
              Delete this fact permanently?
            </span>
            <Button variant="danger" loading={isPending} onClick={remove}>
              Yes, delete
            </Button>
            <Button
              variant="ghost"
              onClick={() => {
                setIsConfirming(false);
              }}
              disabled={isPending}
            >
              Cancel
            </Button>
          </>
        ) : (
          <Button
            variant="danger"
            onClick={() => {
              setIsConfirming(true);
            }}
          >
            Delete
          </Button>
        )}
      </div>
      {error !== null && error !== undefined ? (
        <div className={styles.error}>
          <RequestError error={error} title="Could not delete the fact" />
        </div>
      ) : null}
    </div>
  );
};
