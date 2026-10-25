import { describe, expect, it, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { AppRoutes } from "@/app/routes";
import { ApiError } from "@/shared/api";
import {
  createFakeApi,
  factFixture,
  renderWithProviders,
} from "@/shared/test/renderWithProviders";

const ROUTE = `/fact/${factFixture.id}`;

const renderFactRoute = (
  api = createFakeApi({ getFact: () => Promise.resolve(factFixture) }),
) => renderWithProviders(<AppRoutes />, api, { route: ROUTE });

describe("FactPage", () => {
  it("renders the summary, the id and the markdown content", async () => {
    renderFactRoute();

    expect(
      await screen.findByRole("heading", { name: factFixture.summary }),
    ).toBeInTheDocument();
    expect(screen.getByText(factFixture.id)).toBeInTheDocument();
    // Markdown is parsed: a `##` heading becomes a real heading element.
    expect(
      screen.getByRole("heading", { name: "Preferences" }),
    ).toBeInTheDocument();
    expect(screen.getByText("dark theme")).toBeInTheDocument();
  });

  it("does not render raw HTML from the content", async () => {
    renderFactRoute(
      createFakeApi({
        getFact: () =>
          Promise.resolve({
            ...factFixture,
            content: '<img src=x onerror="alert(1)">',
          }),
      }),
    );

    await screen.findByRole("heading", { name: factFixture.summary });
    expect(document.querySelector("img")).toBeNull();
  });

  it("navigates back to the home page through the injected callback", async () => {
    renderFactRoute();

    await screen.findByRole("heading", { name: factFixture.summary });
    await userEvent.click(screen.getByRole("button", { name: "Back" }));

    expect(await screen.findByText(/Your memory,/)).toBeInTheDocument();
  });

  it("shows a dedicated state for a missing fact", async () => {
    renderFactRoute(
      createFakeApi({
        getFact: () =>
          Promise.reject(new ApiError("FACT_DOES_NOT_EXIST", { status: 404 })),
      }),
    );

    expect(await screen.findByText("Fact not found")).toBeInTheDocument();
  });

  it("deletes after explicit confirmation", async () => {
    const deleteFact = vi.fn(() => Promise.resolve());
    renderFactRoute(
      createFakeApi({
        getFact: () => Promise.resolve(factFixture),
        deleteFact,
      }),
    );

    await screen.findByRole("heading", { name: factFixture.summary });

    // Nothing is deleted on the first click: confirmation is required.
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    expect(deleteFact).not.toHaveBeenCalled();

    await userEvent.click(screen.getByRole("button", { name: "Yes, delete" }));

    expect(deleteFact).toHaveBeenCalledWith(factFixture.id);
  });

  it("stays on the page and reports an error when deletion fails", async () => {
    renderFactRoute(
      createFakeApi({
        getFact: () => Promise.resolve(factFixture),
        deleteFact: () =>
          Promise.reject(new ApiError("FACT_DOES_NOT_EXIST", { status: 404 })),
      }),
    );

    await screen.findByRole("heading", { name: factFixture.summary });
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));
    await userEvent.click(screen.getByRole("button", { name: "Yes, delete" }));

    expect(
      await screen.findByText("Could not delete the fact"),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: factFixture.summary }),
    ).toBeInTheDocument();
  });
});
