import type { ReactNode } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

import styles from "./MarkdownView.module.css";

/**
 * Renders fact content as Markdown. No `rehype-raw` on purpose: raw HTML stays
 * escaped, so stored content cannot inject markup into the app.
 */
export const MarkdownView = ({
  content,
}: {
  readonly content: string;
}): ReactNode => (
  <div className={styles.markdown}>
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        a: ({ href, children }) => (
          <a href={href} target="_blank" rel="noopener noreferrer">
            {children}
          </a>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  </div>
);
