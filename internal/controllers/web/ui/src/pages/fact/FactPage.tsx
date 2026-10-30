import type { ReactNode } from "react";
import { useParams } from "react-router";

import { RequestError } from "@/app/auth/RequestError";
import { Layout } from "@/app/layout/Layout";
import { formatDate } from "@/entities/fact/model/formatDate";
import { MarkdownView } from "@/entities/fact/ui/MarkdownView";
import { DeleteFactButton } from "@/features/delete-fact";
import { useFact } from "@/features/fact-detail";
import { ApiError } from "@/shared/api/ApiError";
import { Button } from "@/shared/ui/Button";
import { EmptyState } from "@/shared/ui/EmptyState";
import { Meta, MetaDot } from "@/shared/ui/Meta";
import { Spinner } from "@/shared/ui/Spinner";
import styles from "./FactPage.module.css";

export type FactPageProps = {
  readonly onBack: () => void;
  readonly homeHref: string;
};

const isNotFound = (error: unknown): boolean =>
  error instanceof ApiError &&
  (error.code === "FACT_DOES_NOT_EXIST" ||
    error.code === "NOT_FOUND" ||
    error.status === 404);

/** The fact id is a route parameter: reading it is routing, not navigation. */
export const FactPage = ({ onBack, homeHref }: FactPageProps): ReactNode => {
  const { factID = "" } = useParams<{ factID: string }>();
  const { fact, error, isPending, refetch } = useFact(factID);

  return (
    <Layout homeHref={homeHref}>
      <div className={styles.actions}>
        <Button variant="secondary" onClick={onBack} icon="←">
          Back
        </Button>
        {fact !== undefined ? (
          <DeleteFactButton factId={fact.id} onDeleted={onBack} />
        ) : null}
      </div>

      {isPending ? <Spinner label="Loading the fact" /> : null}

      {error !== null && error !== undefined ? (
        isNotFound(error) ? (
          <EmptyState
            title="Fact not found"
            description="It may have been deleted."
            action={<a href={homeHref}>Back to home</a>}
          />
        ) : (
          <RequestError
            error={error}
            onRetry={refetch}
            title="Could not load the fact"
          />
        )
      ) : null}

      {fact !== undefined ? (
        <article className={styles.body}>
          <header className={styles.header}>
            <h1 className={styles.title}>{fact.summary}</h1>
            <Meta mono>
              <span>{fact.id}</span>
              <MetaDot />
              <span>{formatDate(fact.createdAt)}</span>
            </Meta>
          </header>
          <MarkdownView content={fact.content} />
        </article>
      ) : null}
    </Layout>
  );
};
