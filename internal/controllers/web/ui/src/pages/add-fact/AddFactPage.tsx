import type { ReactNode } from "react";

import { Layout } from "@/app/layout/Layout";
import { AddFactForm } from "@/features/add-fact";
import { Button } from "@/shared/ui/Button";
import styles from "./AddFactPage.module.css";

export type AddFactPageProps = {
  readonly onBack: () => void;
  readonly onSearchSimilar: (summary: string) => void;
  readonly homeHref: string;
};

export const AddFactPage = ({
  onBack,
  onSearchSimilar,
  homeHref,
}: AddFactPageProps): ReactNode => (
  <Layout homeHref={homeHref}>
    <Button variant="ghost" className={styles.back} icon="←" onClick={onBack}>
      Back to home
    </Button>

    <header className={styles.header}>
      <h1 className={styles.title}>New fact</h1>
      <p className={styles.subtitle}>
        The text will be summarized and added to memory.
      </p>
    </header>

    <AddFactForm onSearchSimilar={onSearchSimilar} onGoHome={onBack} />
  </Layout>
);
