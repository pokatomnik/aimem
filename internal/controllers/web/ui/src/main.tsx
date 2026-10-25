import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClientProvider } from "@tanstack/react-query";

import { App } from "@/App";
import { createFactApi } from "@/shared/api";
import { createQueryClient } from "@/shared/api/queryClient";
import "@/app/styles/reset.css";
import "@/app/styles/tokens.css";
import "@/app/styles/global.css";

const container = document.getElementById("root");
if (container === null) {
  throw new Error("#root element is missing in index.html");
}

// No base URL: the SPA and the API are served by the same binary on the same
// host, so every request is relative to the current origin.
const api = createFactApi();
const queryClient = createQueryClient();

createRoot(container).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <App api={api} />
    </QueryClientProvider>
  </StrictMode>,
);
