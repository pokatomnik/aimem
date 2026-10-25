import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import { MemoryRouter, useLocation } from "react-router";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { useAppNavigation } from "./useAppNavigation";

/**
 * React Router assigns a fresh `key` to every history entry, including one whose
 * URL is identical to the current entry. Comparing keys is therefore the direct
 * way to prove that a second submit did not push anything.
 */
const Probe = (): ReactNode => {
  const nav = useAppNavigation();
  const location = useLocation();

  return (
    <div>
      <span data-testid="key">{location.key}</span>
      <span data-testid="path">
        {location.pathname}
        {location.search}
      </span>
      <button
        type="button"
        onClick={() => {
          nav.goToSearch("coffee");
        }}
      >
        search coffee
      </button>
      <button
        type="button"
        onClick={() => {
          nav.goToSearch("tea");
        }}
      >
        search tea
      </button>
    </div>
  );
};

const renderProbe = (route: string): void => {
  render(
    <MemoryRouter initialEntries={[route]}>
      <Probe />
    </MemoryRouter>,
  );
};

describe("useAppNavigation.goToSearch", () => {
  it("does not push a history entry for the query already shown", async () => {
    renderProbe("/search?q=coffee");

    const keyBefore = screen.getByTestId("key").textContent;
    expect(keyBefore).not.toBe("");

    await userEvent.click(screen.getByRole("button", { name: "search coffee" }));

    expect(screen.getByTestId("key").textContent).toBe(keyBefore);
    expect(screen.getByTestId("path").textContent).toBe("/search?q=coffee");
  });

  it("still navigates for a different query", async () => {
    renderProbe("/search?q=coffee");

    const keyBefore = screen.getByTestId("key").textContent;

    await userEvent.click(screen.getByRole("button", { name: "search tea" }));

    expect(screen.getByTestId("key").textContent).not.toBe(keyBefore);
    expect(screen.getByTestId("path").textContent).toBe("/search?q=tea");
  });

  // The guard must be scoped to the search page: from the landing page the same
  // text is not "already shown", it is a brand new search.
  it("navigates from the landing page even though nothing is shown yet", async () => {
    renderProbe("/");

    await userEvent.click(screen.getByRole("button", { name: "search coffee" }));

    expect(screen.getByTestId("path").textContent).toBe("/search?q=coffee");
  });
});
