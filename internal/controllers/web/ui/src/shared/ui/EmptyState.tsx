import type { ReactNode } from 'react';

import styles from './EmptyState.module.css';

export type EmptyStateProps = {
  readonly title: string;
  readonly description?: ReactNode;
  readonly action?: ReactNode;
  readonly centered?: boolean;
};

/**
 * Neutral "nothing here" presentation, shared by the search results and the
 * fact page, so neither has to know how an empty result looks.
 */
export const EmptyState = ({
  title,
  description,
  action,
  centered = false,
}: EmptyStateProps): ReactNode => (
  <div className={[styles.wrap, centered ? styles.centered : ''].filter(Boolean).join(' ')}>
    <p className={styles.title}>{title}</p>
    {description !== undefined ? <p className={styles.description}>{description}</p> : null}
    {action !== undefined ? <div className={styles.action}>{action}</div> : null}
  </div>
);
