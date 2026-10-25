import type { ReactNode } from "react";

import { Layout } from "@/app/layout/Layout";
import { EmptyState } from "@/shared/ui/EmptyState";
import styles from "./NotFoundPage.module.css";

export type NotFoundPageProps = {
  readonly homeHref: string;
};

export const NotFoundPage = ({ homeHref }: NotFoundPageProps): ReactNode => (
  <Layout homeHref={homeHref}>
    <div className={styles.wrap}>
      <EmptyState
        title="Page not found"
        description="The link may be outdated."
        action={<a href={homeHref}>Back to home</a>}
      />
    </div>
  </Layout>
);
