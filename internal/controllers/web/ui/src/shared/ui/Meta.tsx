import type { ReactNode } from "react";

import styles from "./Meta.module.css";

type MetaProps = {
  readonly children: ReactNode;
  readonly mono?: boolean;
};

/** Secondary line under a title: timestamp, identifier, counter. */
export const Meta = ({ children, mono = false }: MetaProps): ReactNode => (
  <div
    className={[styles.meta, mono ? styles.mono : ""].filter(Boolean).join(" ")}
  >
    {children}
  </div>
);

/** Visual separator between meta items. */
export const MetaDot = (): ReactNode => (
  <span className={styles.dot} aria-hidden="true" />
);
