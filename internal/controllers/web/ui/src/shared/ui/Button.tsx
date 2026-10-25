import type { ButtonHTMLAttributes, ReactNode } from 'react';

import styles from './Button.module.css';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  readonly variant?: ButtonVariant;
  /** Renders a spinner and disables the button: one click per request. */
  readonly loading?: boolean;
  readonly fullWidth?: boolean;
  /** Emoji or short text shown before the label. */
  readonly icon?: ReactNode;
};

const VARIANT_CLASS: Record<ButtonVariant, string> = {
  primary: styles.primary ?? '',
  secondary: styles.secondary ?? '',
  ghost: styles.ghost ?? '',
  danger: styles.danger ?? '',
};

export const Button = ({
  variant = 'primary',
  loading = false,
  fullWidth = false,
  icon,
  disabled,
  children,
  type = 'button',
  className,
  ...rest
}: ButtonProps): ReactNode => (
  <button
    type={type}
    className={[styles.base, VARIANT_CLASS[variant], fullWidth ? styles.fullWidth : '', className ?? '']
      .filter(Boolean)
      .join(' ')}
    disabled={disabled === true || loading}
    aria-busy={loading}
    {...rest}
  >
    {loading ? <span className={styles.spinner} aria-hidden="true" /> : null}
    {!loading && icon !== undefined ? (
      <span className={styles.icon} aria-hidden="true">
        {icon}
      </span>
    ) : null}
    {children}
  </button>
);
