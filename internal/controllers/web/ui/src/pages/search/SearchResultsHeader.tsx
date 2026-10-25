import type { ReactNode } from 'react';

import { formatResultCount } from '@/shared/lib/text';
import styles from './SearchResultsHeader.module.css';

export type SearchResultsHeaderProps = {
  readonly count: number;
};

/** Owns the "N results" / "No results" line above the list. */
export const SearchResultsHeader = ({ count }: SearchResultsHeaderProps): ReactNode => (
  <p className={styles.count}>{count === 0 ? 'No results' : formatResultCount(count)}</p>
);
