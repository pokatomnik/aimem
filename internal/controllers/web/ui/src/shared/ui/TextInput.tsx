import type {
  InputHTMLAttributes,
  ReactNode,
  TextareaHTMLAttributes,
} from "react";

import styles from "./TextInput.module.css";

type Common = {
  readonly label: string;
  readonly hint?: string;
  readonly hintError?: boolean;
};

export type TextInputProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "size"
> &
  Common;

export const TextInput = ({
  label,
  hint,
  hintError = false,
  id,
  className,
  ...rest
}: TextInputProps): ReactNode => {
  const inputId = id ?? `input-${label.replace(/\s+/g, "-").toLowerCase()}`;
  return (
    <div className={styles.wrap}>
      <input
        id={inputId}
        aria-label={label}
        aria-invalid={hintError}
        className={[styles.field, className ?? ""].filter(Boolean).join(" ")}
        {...rest}
      />
      {hint !== undefined ? (
        <p
          className={[styles.hint, hintError ? styles.hintError : ""]
            .filter(Boolean)
            .join(" ")}
        >
          {hint}
        </p>
      ) : null}
    </div>
  );
};

export type TextAreaProps = TextareaHTMLAttributes<HTMLTextAreaElement> &
  Common;

export const TextArea = ({
  label,
  hint,
  hintError = false,
  id,
  className,
  ...rest
}: TextAreaProps): ReactNode => {
  const areaId = id ?? `textarea-${label.replace(/\s+/g, "-").toLowerCase()}`;
  return (
    <div className={styles.wrap}>
      <textarea
        id={areaId}
        aria-label={label}
        aria-invalid={hintError}
        className={[styles.field, styles.textarea, className ?? ""]
          .filter(Boolean)
          .join(" ")}
        {...rest}
      />
      {hint !== undefined ? (
        <p
          className={[styles.hint, hintError ? styles.hintError : ""]
            .filter(Boolean)
            .join(" ")}
        >
          {hint}
        </p>
      ) : null}
    </div>
  );
};
