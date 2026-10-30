import { describe, expect, it, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { ApiError } from "@/shared/api";
import { createMemoryTokenStorage } from "@/shared/api/tokenStorage";
import {
  createFakeApi,
  factListItemFixture,
  renderWithProviders,
} from "@/shared/test/renderWithProviders";
import { SearchPage } from "./SearchPage";

const ROUTE = "/search?q=coffee";

const renderSearchPage = (
  api = createFakeApi(),
  overrides: Partial<Parameters<typeof SearchPage>[0]> = {},
  options: Parameters<typeof renderWithProviders>[2] = {},
): ReturnType<typeof renderWithProviders> =>
  renderWithProviders(
    <SearchPage
      onSearch={vi.fn()}
      factHref={(id) => `/fact/${id}`}
      homeHref="/"
      addFactHref="/fact/add"
      {...overrides}
    />,
    api,
    { route: ROUTE, ...options },
  );

describe("SearchPage", () => {
  it("renders a hit with its id and summary only", async () => {
    const api = createFakeApi({
      search: () => Promise.resolve([factListItemFixture]),
    });
    renderSearchPage(api);

    expect(
      await screen.findByText(factListItemFixture.summary),
    ).toBeInTheDocument();
    expect(screen.getByText(factListItemFixture.id)).toBeInTheDocument();
    expect(screen.getByText("1 result")).toBeInTheDocument();
  });

  it("reads the query from the URL, not from the form state", async () => {
    const search = vi.fn(() => Promise.resolve([factListItemFixture]));
    renderSearchPage(createFakeApi({ search }));

    await vi.waitFor(() => {
      expect(search).toHaveBeenCalledWith("coffee", expect.anything());
    });
  });

  it("shows an empty state when nothing matched", async () => {
    renderSearchPage(createFakeApi({ search: () => Promise.resolve([]) }));

    expect(
      await screen.findByText("No results for “coffee”"),
    ).toBeInTheDocument();
    expect(screen.getByText("No results")).toBeInTheDocument();
  });

  it("shows a readable message for a failed search", async () => {
    // The real client converts transport throws into `ApiError`, so the fake
    // reproduces that behaviour instead of leaking a raw TypeError into the UI.
    renderSearchPage(
      createFakeApi({
        search: () => Promise.reject(new ApiError("NETWORK_ERROR")),
      }),
    );

    expect(await screen.findByText("Search failed")).toBeInTheDocument();
    expect(screen.getByText(/Server is unreachable/)).toBeInTheDocument();
  });

  it("prompts for a query when q is missing or too short", () => {
    const search = vi.fn(() => Promise.resolve([]));
    renderWithProviders(
      <SearchPage
        onSearch={vi.fn()}
        factHref={(id) => `/fact/${id}`}
        homeHref="/"
        addFactHref="/fact/add"
      />,
      createFakeApi({ search }),
      { route: "/search?q=ab" },
    );

    expect(screen.getByText("Enter a search query")).toBeInTheDocument();
    expect(search).not.toHaveBeenCalled();
  });

  it("links each hit using the href the page was given", async () => {
    const api = createFakeApi({
      search: () => Promise.resolve([factListItemFixture]),
    });
    renderSearchPage(api, { factHref: (id) => `/custom/${id}` });

    const link = await screen.findByRole("link", {
      name: new RegExp(factListItemFixture.summary),
    });
    expect(link).toHaveAttribute("href", `/custom/${factListItemFixture.id}`);
  });

  it("offers Logout next to Retry when the token was rejected", async () => {
    renderSearchPage(
      createFakeApi({
        search: () =>
          Promise.reject(new ApiError("UNAUTHORIZED", { status: 401 })),
      }),
      {},
      { authRequired: true, tokenStorage: createMemoryTokenStorage("stale") },
    );

    expect(
      await screen.findByRole("button", { name: "Retry" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Logout" })).toBeInTheDocument();
  });

  it("keeps only Retry when the failure is not about the token", async () => {
    renderSearchPage(
      createFakeApi({
        search: () => Promise.reject(new ApiError("NETWORK_ERROR")),
      }),
      {},
      { authRequired: true, tokenStorage: createMemoryTokenStorage("ok") },
    );

    expect(
      await screen.findByRole("button", { name: "Retry" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Logout" }),
    ).not.toBeInTheDocument();
  });

  it("delegates a new search to the injected callback", async () => {
    const onSearch = vi.fn();
    renderSearchPage(createFakeApi(), { onSearch });

    // The field is pre-filled with the current query from the URL.
    await userEvent.clear(screen.getByRole("searchbox"));
    await userEvent.type(screen.getByRole("searchbox"), "tea{Enter}");

    expect(onSearch).toHaveBeenCalledWith("tea");
  });
});
