import type { ReactNode } from "react";
import { useState } from "react";

import styles from "./TokenGate.module.css";
import { isTokenSendable } from "@/shared/api/token";
import { Button } from "@/shared/ui/Button";
import { TextInput } from "@/shared/ui/TextInput";

export type TokenGateProps = {
  readonly onSave: (token: string) => void;
};

/**
 * The only thing a user sees until a token is stored. There is no login flow, so
 * it is deliberately a single field and a single button — the token is the whole
 * session, and it is one of the tokens the operator listed in the server's
 * `AIMEM_USERS_JSON`.
 *
 * Logging out returns here: dropping the stored token leaves the gate with no
 * token to show, which is exactly the state this component renders.
 */
export const TokenGate = ({ onSave }: TokenGateProps): ReactNode => {
  const [value, setValue] = useState("");
  const trimmed = value.trim();
  // Empty is "nothing typed yet"; unsendable is "typed, but the browser would
  // refuse to put it in a header". Both block Save, and only the second one needs
  // an explanation — the field is not wrong, the characters are.
  const sendable = isTokenSendable(trimmed);
  const valid = trimmed !== "" && sendable;

  return (
    <div className={styles.shell}>
      <header className={styles.topbar}>
        <div className={styles.topbarInner}>
          <span className={styles.brand}>
            <span className={styles.brandMark} aria-hidden="true" />
            <span>
              aimem<span className={styles.brandMuted}>.web</span>
            </span>
          </span>
        </div>
      </header>

      <main className={styles.main}>
        <div className={styles.card}>
          <h1 className={styles.title}>Token required</h1>
          <p className={styles.subtitle}>
            This server keeps a separate memory per user, so it needs to know
            who you are. Paste a token issued by the server operator.
          </p>

          <form
            className={styles.form}
            onSubmit={(event) => {
              event.preventDefault();
              if (valid) {
                onSave(trimmed);
              }
            }}
          >
            <TextInput
              label="Access token"
              placeholder="a-long-random-string"
              value={value}
              onChange={(event) => {
                setValue(event.target.value);
              }}
              hint={
                trimmed !== "" && !sendable
                  ? "A token goes into an HTTP header, so it can only use Latin letters, digits and symbols — no Cyrillic or emoji."
                  : null
              }
              hintError={trimmed !== "" && !sendable}
              autoFocus
              autoComplete="off"
              spellCheck={false}
            />
            <Button type="submit" disabled={!valid}>
              Save
            </Button>
          </form>

          <p className={styles.hint}>
            The token is kept in this browser&apos;s local storage and sent with
            every request.
          </p>
        </div>
      </main>

      <footer className={styles.footer}>memory · embeddings · HNSW</footer>
    </div>
  );
};
