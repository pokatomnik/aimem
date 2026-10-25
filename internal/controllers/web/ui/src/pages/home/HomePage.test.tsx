import { describe, expect, it, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { AppRoutes } from "@/app/routes";
import {
  createFakeApi,
  renderWithProviders,
} from "@/shared/test/renderWithProviders";

describe("HomePage", () => {
  it("shows the search field with the magnifier button", () => {
    renderWithProviders(<AppRoutes />, createFakeApi(), { route: "/" });

    expect(screen.getByRole("searchbox")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Search" }).textContent,
    ).toContain("🔍");
  });

  it("navigates to /search with a properly encoded query", async () => {
    const search = vi.fn(() => Promise.resolve([]));
    renderWithProviders(<AppRoutes />, createFakeApi({ search }), {
      route: "/",
    });

    await userEvent.type(
      screen.getByRole("searchbox"),
      "what do I like{Enter}",
    );

    await vi.waitFor(() => {
      expect(search).toHaveBeenCalledWith("what do I like", expect.anything());
    });
  });

  it("offers a link to the add-fact page", async () => {
    renderWithProviders(<AppRoutes />, createFakeApi(), { route: "/" });

    await userEvent.click(screen.getByRole("link", { name: "Add a fact" }));

    expect(await screen.findByLabelText("New fact")).toBeInTheDocument();
  });

  it("does not query the backend on the landing page", () => {
    const search = vi.fn(() => Promise.resolve([]));
    renderWithProviders(<AppRoutes />, createFakeApi({ search }), {
      route: "/",
    });

    expect(search).not.toHaveBeenCalled();
  });
});
