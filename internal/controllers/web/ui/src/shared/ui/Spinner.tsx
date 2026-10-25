import type { ReactNode } from "react";

import styles from "./Spinner.module.css";

/** Waiting state. `role="status"` so screen readers announce the change. */
export const Spinner = ({
  label = "Loading",
}: {
  readonly label?: string;
}): ReactNode => (
  <div className={styles.wrap} role="status">
    <span className={styles.dot} aria-hidden="true" />
    <span className={styles.dot} aria-hidden="true" />
    <span className={styles.dot} aria-hidden="true" />
    <span className={styles.visuallyHidden}>{label}</span>
  </div>
);
