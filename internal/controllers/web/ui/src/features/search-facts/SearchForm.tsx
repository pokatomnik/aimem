import type { ReactNode } from "react";
import { useState } from "react";

import { isSearchQueryValid, SEARCH_QUERY_MIN_LENGTH } from "@/shared/lib/text";
import { Button } from "@/shared/ui/Button";
import styles from "./SearchForm.module.css";

export type SearchFormProps = {
  /** `hero` centers the field on the landing page, `compact` sits in the results header. */
  readonly variant?: "hero" | "compact";
  readonly initialQuery?: string;
  /** Called with the trimmed query; navigation is the caller's job. */
  readonly onSearch: (query: string) => void;
  readonly autoFocus?: boolean;
};

export const SearchForm = ({
  variant = "compact",
  initialQuery = "",
  onSearch,
  autoFocus = false,
}: SearchFormProps): ReactNode => {
  const [value, setValue] = useState(initialQuery);
  const valid = isSearchQueryValid(value);

  return (
    <form
      className={[
        styles.form,
        variant === "hero" ? styles.hero : styles.compact,
      ]
        .filter(Boolean)
        .join(" ")}
      role="search"
      onSubmit={(event) => {
        event.preventDefault();
        if (valid) {
          onSearch(value.trim());
        }
      }}
    >
      <input
        className={styles.input}
        type="search"
        name="q"
        value={value}
        onChange={(event) => {
          setValue(event.target.value);
        }}
        placeholder="Ask your memory…"
        aria-label="Search query"
        autoComplete="off"
        autoFocus={autoFocus}
      />
      <Button
        type="submit"
        variant="primary"
        className={styles.submit}
        disabled={!valid}
        icon="🔍"
        aria-label="Search"
      >
        Search
      </Button>
      {!valid && value.length > 0 ? (
        <p className={styles.hint}>
          At least {String(SEARCH_QUERY_MIN_LENGTH)} characters
        </p>
      ) : null}
    </form>
  );
};
