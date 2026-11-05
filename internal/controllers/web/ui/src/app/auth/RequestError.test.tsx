import { describe, expect, it } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { AuthGate, AuthProvider, RequestError } from "@/app/auth";
import { ApiError } from "@/shared/api";
import { createMemoryTokenStorage } from "@/shared/api/tokenStorage";

/**
 * The whole session, without the app: a gate over a failing request. This is the
 * shape the search page has when the server rejects the stored token — gate
 * passed, request failed — and logout has to undo the first to recover from the
 * second.
 */
const renderFailedSession = (
  error: ApiError,
  authRequired = true,
  storage = createMemoryTokenStorage("stale-token"),
): {
  readonly storage: ReturnType<typeof createMemoryTokenStorage>;
  readonly queryClient: QueryClient;
} => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });

  render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider required={authRequired} storage={storage}>
        <AuthGate>
          <RequestError
            error={error}
            onRetry={() => undefined}
            title="Search failed"
          />
        </AuthGate>
      </AuthProvider>
    </QueryClientProvider>,
  );

  return { storage, queryClient };
};

const unauthorized = (): ApiError =>
  new ApiError("UNAUTHORIZED", { status: 401 });

describe("RequestError", () => {
  it("shows the message and keeps Retry", () => {
    renderFailedSession(unauthorized());

    expect(screen.getByText("Search failed")).toBeInTheDocument();
    expect(
      screen.getByText(/Not authorized. Check the token/),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });

  it("offers Logout next to Retry when the token was rejected", () => {
    renderFailedSession(unauthorized());

    expect(screen.getByRole("button", { name: "Logout" })).toBeInTheDocument();
  });

  it("does not offer Logout for a failure that is not about credentials", () => {
    // A failed model API is answered by Retry: a logout button there would send
    // the user to the token form to fix something the token did not cause.
    renderFailedSession(new ApiError("NETWORK_ERROR"));

    expect(
      screen.queryByRole("button", { name: "Logout" }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });

  it("never offers Logout when the deployment needs no token", () => {
    renderFailedSession(unauthorized(), false, createMemoryTokenStorage());

    expect(
      screen.queryByRole("button", { name: "Logout" }),
    ).not.toBeInTheDocument();
  });

  it("drops the token and brings the form back on logout", async () => {
    const user = userEvent.setup();
    const { storage } = renderFailedSession(unauthorized());

    await user.click(screen.getByRole("button", { name: "Logout" }));

    await waitFor(() => {
      expect(screen.getByLabelText("Access token")).toBeInTheDocument();
    });
    expect(screen.queryByText("Search failed")).not.toBeInTheDocument();
    expect(storage.get()).toBeUndefined();
  });

  it("clears cached answers fetched under the dropped identity", async () => {
    const user = userEvent.setup();
    const { queryClient } = renderFailedSession(unauthorized());

    queryClient.setQueryData(["fact", "search", "coffee"], [{ id: "leaked" }]);
    expect(queryClient.getQueryCache().getAll()).not.toHaveLength(0);

    await user.click(screen.getByRole("button", { name: "Logout" }));

    // A cached hit belongs to whoever asked for it; after a fresh token is
    // entered it must not be shown as if it were the new user's memory.
    await waitFor(() => {
      expect(queryClient.getQueryCache().getAll()).toHaveLength(0);
    });
  });
});
