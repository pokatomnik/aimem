import type { ReactNode } from "react";
import { HashRouter } from "react-router";

import { AuthGate, AuthProvider } from "@/app/auth";
import { ApiProvider } from "@/app/providers/ApiProvider";
import { AppRoutes } from "@/app/routes";
import type { FactApi } from "@/shared/api";
import type { TokenStorage } from "@/shared/api/tokenStorage";

type AppProps = {
  /** Injected boundary: production passes the fetch client, tests pass a fake. */
  readonly api: FactApi;
  /** Whether this deployment needs a token; comes from the public auth endpoint. */
  readonly authRequired: boolean;
  /** Injected so a test never touches `window.localStorage`. */
  readonly tokenStorage: TokenStorage;
};

/**
 * `HashRouter` is a requirement: the Go binary serves one HTML file at `/`, and
 * everything after `#` never reaches the server.
 *
 * The gate sits inside the provider but outside the routes: once a token is
 * saved it re-renders the routes, and no page needs to know it exists.
 */
export const App = ({
  api,
  authRequired,
  tokenStorage,
}: AppProps): ReactNode => (
  <HashRouter>
    <AuthProvider required={authRequired} storage={tokenStorage}>
      <ApiProvider api={api}>
        <AuthGate>
          <AppRoutes />
        </AuthGate>
      </ApiProvider>
    </AuthProvider>
  </HashRouter>
);
