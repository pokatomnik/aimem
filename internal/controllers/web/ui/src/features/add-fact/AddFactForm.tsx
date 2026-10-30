import type { ReactNode } from "react";
import { useState } from "react";

import { RequestError } from "@/app/auth/RequestError";
import { isFactValid, normalizeQuery } from "@/shared/lib/text";
import { Button } from "@/shared/ui/Button";
import { TextArea } from "@/shared/ui/TextInput";
import styles from "./AddFactForm.module.css";
import { useAddFact } from "./useAddFact";

export type AddFactFormProps = {
  /** Navigation is injected: the form only says *what* the user wants. */
  readonly onSearchSimilar: (summary: string) => void;
  readonly onGoHome: () => void;
};

/**
 * Plain textarea on purpose: a Markdown editor is explicitly out of scope, and
 * the backend treats the input as text anyway.
 */
export const AddFactForm = ({
  onSearchSimilar,
  onGoHome,
}: AddFactFormProps): ReactNode => {
  const [value, setValue] = useState("");
  const { summary, error, isPending, submit, reset } = useAddFact();
  const valid = isFactValid(value);

  // The backend returns no id, so the success screen offers a summary-based
  // search instead of a link to a fact page that could not be built.
  if (summary !== undefined) {
    return (
      <div className={styles.success} role="status">
        <p className={styles.successTitle}>Fact added</p>
        <p className={styles.summary}>Summary: “{summary}”</p>
        <div className={styles.row}>
          <Button
            variant="primary"
            onClick={() => {
              setValue("");
              reset();
            }}
          >
            Add another
          </Button>
          <Button
            variant="secondary"
            onClick={() => {
              onSearchSimilar(summary);
            }}
          >
            Find similar
          </Button>
          <Button variant="ghost" onClick={onGoHome}>
            Back to home
          </Button>
        </div>
      </div>
    );
  }

  return (
    <form
      className={styles.form}
      onSubmit={(event) => {
        event.preventDefault();
        if (valid && !isPending) {
          submit(normalizeQuery(value));
        }
      }}
    >
      <TextArea
        label="New fact"
        placeholder={
          "For example: “I prefer a dark theme in editors”.\n\nMarkdown is supported."
        }
        value={value}
        onChange={(event) => {
          setValue(event.target.value);
        }}
        rows={10}
        disabled={isPending}
      />
      <div className={styles.row}>
        <Button type="submit" loading={isPending} disabled={!valid}>
          Remember
        </Button>
        <span className={styles.counter}>{value.trim().length} characters</span>
      </div>
      {error !== null && error !== undefined ? (
        <RequestError error={error} title="Could not add the fact" />
      ) : null}
    </form>
  );
};
