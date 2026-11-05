import type { ReactNode } from "react";
import { Link } from "react-router";

import styles from "./Layout.module.css";

type LayoutProps = {
  readonly children: ReactNode;
  /** Landing page centers its content vertically; other pages start at the top. */
  readonly centered?: boolean;
  readonly headerExtra?: ReactNode;
  readonly homeHref: string;
};

export const Layout = ({
  children,
  centered = false,
  headerExtra,
  homeHref,
}: LayoutProps): ReactNode => (
  <div className={styles.shell}>
    <header className={styles.topbar}>
      <div className={styles.topbarInner}>
        <Link to={homeHref} className={styles.brand}>
          <span className={styles.brandMark} aria-hidden="true" />
          <span>
            aimem<span className={styles.brandMuted}>.web</span>
          </span>
        </Link>
        {headerExtra !== undefined ? (
          <>
            <span className={styles.spacer} />
            {headerExtra}
          </>
        ) : null}
      </div>
    </header>

    <main className={centered ? styles.mainCentered : styles.main}>
      {children}
    </main>

    <footer className={styles.footer}>memory · embeddings · HNSW</footer>
  </div>
);
