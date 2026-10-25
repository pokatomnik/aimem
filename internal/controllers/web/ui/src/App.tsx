import type { ReactNode } from "react";
import { HashRouter } from "react-router";

import { ApiProvider } from "@/app/providers/ApiProvider";
import { AppRoutes } from "@/app/routes";
import type { FactApi } from "@/shared/api";

type AppProps = {
  /** Injected boundary: production passes the fetch client, tests pass a fake. */
  readonly api: FactApi;
};

/**
 * `HashRouter` is a requirement: the Go binary serves one HTML file at `/`, and
 * everything after `#` never reaches the server.
 */
export const App = ({ api }: AppProps): ReactNode => (
  <HashRouter>
    <ApiProvider api={api}>
      <AppRoutes />
    </ApiProvider>
  </HashRouter>
);
