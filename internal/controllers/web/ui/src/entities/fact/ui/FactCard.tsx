import type { ReactNode } from "react";
import { Link } from "react-router";

import type { FactListItem } from "@/entities/fact/api/keys";
import { formatDateShort } from "@/entities/fact/model/formatDate";
import { Meta, MetaDot } from "@/shared/ui/Meta";
import styles from "./FactCard.module.css";

type FactCardProps = {
  readonly fact: FactListItem;
  /** Resolved by the caller, so this component never knows the route table. */
  readonly href: string;
};

/**
 * A search hit. The task is explicit: only the id and the summary are shown,
 * because `/api/v1/search` does not return the content.
 */
export const FactCard = ({ fact, href }: FactCardProps): ReactNode => (
  <Link to={href} className={styles.card}>
    <span className={styles.summary}>{fact.summary}</span>
    <Meta mono>
      <span>{fact.id}</span>
      <MetaDot />
      <span>{formatDateShort(fact.createdAt)}</span>
    </Meta>
  </Link>
);
