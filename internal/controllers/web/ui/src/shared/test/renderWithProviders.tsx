import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, renderHook } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import { AuthProvider } from "@/app/auth";
import { ApiProvider } from "@/app/providers/ApiProvider";
import type { FactApi } from "@/shared/api";
import { createMemoryTokenStorage } from "@/shared/api/tokenStorage";
import type { TokenStorage } from "@/shared/api/tokenStorage";

/** A fresh cache per test, with retries and garbage collection disabled. */
const createTestQueryClient = (): QueryClient =>
  new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0, staleTime: 0 },
      mutations: { retry: false, gcTime: 0 },
    },
  });

export type RenderOptions = {
  readonly route?: string;
  /** Defaults to no auth: most pages are tested without a session at all. */
  readonly authRequired?: boolean;
  /** Defaults to an empty in-memory store; pass one to seed a token. */
  readonly tokenStorage?: TokenStorage;
};

/**
 * Renders a component with the real providers and a fake API. Nothing here
 * touches the network, the filesystem or the environment.
 *
 * The auth provider is always mounted, even for pages that do not care: the
 * session is read by shared UI (`RequestError`), and a missing provider is a
 * programming error rather than a valid test setup.
 */
export const renderWithProviders = (
  ui: ReactNode,
  api: FactApi,
  options: RenderOptions = {},
): ReturnType<typeof render> => {
  const queryClient = createTestQueryClient();
  const route = options.route ?? "/";

  return render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider
        required={options.authRequired ?? false}
        storage={options.tokenStorage ?? createMemoryTokenStorage()}
      >
        <ApiProvider api={api}>
          <MemoryRouter initialEntries={[route]}>{ui}</MemoryRouter>
        </ApiProvider>
      </AuthProvider>
    </QueryClientProvider>,
  );
};

export type FakeApiOverrides = Partial<FactApi>;

/**
 * Builds a fake `FactApi` where every method is an explicitly provided stub.
 * The real client converts transport throws into `ApiError`; fakes that need an
 * error should reject with `new ApiError(...)` to mirror that contract.
 */
export const createFakeApi = (overrides: FakeApiOverrides = {}): FactApi => ({
  search: () => Promise.resolve([]),
  getFact: () =>
    Promise.resolve({
      id: "fact-1",
      createdAt: "2026-01-01T00:00:00Z",
      summary: "",
      content: "",
    }),
  addFact: () => Promise.resolve("summary"),
  deleteFact: () => Promise.resolve(),
  ...overrides,
});

/**
 * Hook variant of `renderWithProviders`. The wrapper is rebuilt per call, so
 * every test starts from an empty in-memory cache.
 */
export function renderHookWithProviders<TResult>(
  hook: () => TResult,
  api: FactApi,
): ReturnType<typeof renderHook<TResult, unknown>> {
  const queryClient = createTestQueryClient();

  return renderHook(hook, {
    wrapper: ({ children }: { readonly children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        <ApiProvider api={api}>{children}</ApiProvider>
      </QueryClientProvider>
    ),
  });
}

export const factFixture = {
  id: "a1b2c3",
  createdAt: "2026-03-14T09:26:53Z",
  summary: "I prefer a dark theme in editors",
  content: "## Preferences\n\n- dark theme\n- large font",
} as const;

export const factListItemFixture = {
  id: "a1b2c3",
  createdAt: "2026-03-14T09:26:53Z",
  summary: "I prefer a dark theme in editors",
} as const;
