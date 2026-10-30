import type { ReactNode } from "react";
import { Link, useSearchParams } from "react-router";

import { RequestError } from "@/app/auth/RequestError";
import { Layout } from "@/app/layout/Layout";
import { FactCard } from "@/entities/fact/ui/FactCard";
import { SearchForm, useSearch } from "@/features/search-facts";
import { isSearchQueryValid, normalizeQuery } from "@/shared/lib/text";
import { EmptyState } from "@/shared/ui/EmptyState";
import { Spinner } from "@/shared/ui/Spinner";
import styles from "./SearchPage.module.css";
import { SearchResultsHeader } from "./SearchResultsHeader";

export type SearchPageProps = {
  readonly onSearch: (query: string) => void;
  readonly factHref: (id: string) => string;
  readonly homeHref: string;
  readonly addFactHref: string;
};

/** `q` lives in the URL: it is the only state this page owns. */
export const SearchPage = ({
  onSearch,
  factHref,
  homeHref,
  addFactHref,
}: SearchPageProps): ReactNode => {
  const [searchParams] = useSearchParams();
  const query = normalizeQuery(searchParams.get("q") ?? "");

  const { items, error, isPending, isFetching, refetch } = useSearch(query);

  if (!isSearchQueryValid(query)) {
    return (
      <Layout homeHref={homeHref}>
        <SearchForm
          variant="compact"
          initialQuery={query}
          onSearch={onSearch}
        />
        <EmptyState
          title="Enter a search query"
          description="Describe what you are looking for in your own words."
          action={<Link to={homeHref}>Back to home</Link>}
        />
      </Layout>
    );
  }

  return (
    <Layout homeHref={homeHref}>
      <div className={styles.header}>
        <SearchForm
          variant="compact"
          initialQuery={query}
          onSearch={onSearch}
        />
      </div>

      <div
        className={styles.container}
        data-fetching={isFetching && !isPending}
      >
        {isPending ? <Spinner label="Searching by meaning" /> : null}

        {error !== null && error !== undefined ? (
          <RequestError error={error} onRetry={refetch} title="Search failed" />
        ) : null}

        {items !== undefined ? (
          <>
            <SearchResultsHeader count={items.length} />
            {items.length === 0 ? (
              <EmptyState
                title={`No results for “${query}”`}
                description={
                  <>
                    Try rephrasing: search matches meaning, not exact words. Or{" "}
                    <Link to={addFactHref}>add a fact</Link>.
                  </>
                }
              />
            ) : (
              <ul className={styles.results}>
                {items.map((item) => (
                  <li key={item.id}>
                    <FactCard fact={item} href={factHref(item.id)} />
                  </li>
                ))}
              </ul>
            )}
          </>
        ) : null}
      </div>
    </Layout>
  );
};
