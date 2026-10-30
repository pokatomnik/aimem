import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClientProvider } from "@tanstack/react-query";

import { App } from "@/App";
import { createDefaultTokenStorage } from "@/shared/api/tokenStorage";
import { readAuthRequirement } from "@/shared/api/authRequirement";
import { createFactApi } from "@/shared/api";
import { createQueryClient } from "@/shared/api/queryClient";
import "@/app/styles/reset.css";
import "@/app/styles/tokens.css";
import "@/app/styles/global.css";

const container = document.getElementById("root");
if (container === null) {
  throw new Error("#root element is missing in index.html");
}

const tokenStorage = createDefaultTokenStorage();

// No base URL: the SPA and the API are served by the same binary on the same
// host, so every request is relative to the current origin. The token is read
// through the same storage on every request, so saving one takes effect at once.
const api = createFactApi({ getToken: tokenStorage.get });
const queryClient = createQueryClient();

// The server decides whether a token is needed and bakes the answer into this
// page, so the very first render already knows — no request, no flash of the
// gate on a deployment that does not use one.
const authRequired = readAuthRequirement();

createRoot(container).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <App api={api} authRequired={authRequired} tokenStorage={tokenStorage} />
    </QueryClientProvider>
  </StrictMode>,
);
