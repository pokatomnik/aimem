import type { ReactNode } from "react";
import { Link } from "react-router";

import { Layout } from "@/app/layout/Layout";
import { SearchForm } from "@/features/search-facts";
import styles from "./HomePage.module.css";

export type HomePageProps = {
  readonly onSearch: (query: string) => void;
  readonly addFactHref: string;
  readonly homeHref: string;
};

/**
 * The add-fact link is a real anchor: it is a destination, not an action, so no
 * callback is needed.
 */
export const HomePage = ({
  onSearch,
  addFactHref,
  homeHref,
}: HomePageProps): ReactNode => (
  <Layout centered homeHref={homeHref}>
    <div className={styles.hero}>
      <h1 className={styles.title}>
        Your memory, <span className={styles.accent}>by meaning</span>
      </h1>
      <p className={styles.subtitle}>
        Ask in your own words — aimem finds the nearest facts in a local vector
        index.
      </p>
      <div className={styles.formWrap}>
        <SearchForm variant="hero" autoFocus onSearch={onSearch} />
      </div>
      <div className={styles.links}>
        <Link to={addFactHref} className={styles.link}>
          Add a fact
        </Link>
      </div>
    </div>
  </Layout>
);
