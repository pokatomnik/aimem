import type { ReactNode } from "react";
import { useId, useState } from "react";

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
  const hintId = useId();

  return (
    // The hint lives outside the form: the form is a horizontal row, so a child
    // paragraph would become a flex item beside the button instead of a caption
    // underneath the pill.
    <div className={styles.wrap}>
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
          aria-describedby={hintId}
          autoComplete="off"
          autoFocus={autoFocus}
        />
        {value.length > 0 ? (
          // The native clear control is hidden by the appearance reset, so the
          // affordance is rebuilt here and keeps the same accessible name.
          <button
            type="button"
            className={styles.clear}
            aria-label="Clear query"
            onClick={() => {
              setValue("");
            }}
          >
            ×
          </button>
        ) : null}
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
      </form>
      {/* Always rendered rather than shown on a short query: a caption that
          appears mid-typing would push the page below it every time. */}
      <p className={styles.hint} id={hintId}>
        At least {String(SEARCH_QUERY_MIN_LENGTH)} characters
      </p>
    </div>
  );
};
