import { describe, expect, it, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { ApiError } from "@/shared/api";
import {
  createFakeApi,
  renderWithProviders,
} from "@/shared/test/renderWithProviders";
import { AddFactForm } from "./AddFactForm";

const renderForm = (
  api = createFakeApi(),
  overrides: Partial<Parameters<typeof AddFactForm>[0]> = {},
): ReturnType<typeof renderWithProviders> =>
  renderWithProviders(
    <AddFactForm onSearchSimilar={vi.fn()} onGoHome={vi.fn()} {...overrides} />,
    api,
  );

describe("AddFactForm", () => {
  it("keeps the submit disabled for an empty fact", () => {
    renderForm();

    expect(screen.getByRole("button", { name: "Remember" })).toBeDisabled();
  });

  it("sends the trimmed text and shows the returned summary", async () => {
    const addFact = vi.fn(() => Promise.resolve("The user likes coffee"));
    renderForm(createFakeApi({ addFact }));

    await userEvent.type(
      screen.getByLabelText("New fact"),
      "  I like coffee  ",
    );
    await userEvent.click(screen.getByRole("button", { name: "Remember" }));

    expect(await screen.findByText("Fact added")).toBeInTheDocument();
    expect(addFact).toHaveBeenCalledWith("I like coffee");
    expect(screen.getByText(/The user likes coffee/)).toBeInTheDocument();
  });

  it("does not navigate to a fact page, because the backend returns no id", async () => {
    renderForm(createFakeApi({ addFact: () => Promise.resolve("s") }));

    await userEvent.type(screen.getByLabelText("New fact"), "fact");
    await userEvent.click(screen.getByRole("button", { name: "Remember" }));

    await screen.findByText("Fact added");
    expect(screen.queryByRole("link", { name: /Open fact/ })).toBeNull();
  });

  it("prevents a double submit while the request is in flight", async () => {
    const addFact = vi.fn(() => new Promise<string>(() => undefined));
    renderForm(createFakeApi({ addFact }));

    await userEvent.type(screen.getByLabelText("New fact"), "fact");
    const submit = screen.getByRole("button", { name: /Remember/ });
    await userEvent.click(submit);

    expect(submit).toBeDisabled();
    await userEvent.click(submit);
    expect(addFact).toHaveBeenCalledTimes(1);
  });

  it("shows a readable message when the fact already exists", async () => {
    renderForm(
      createFakeApi({
        addFact: () =>
          Promise.reject(new ApiError("ALREADY_HAS_FACT", { status: 400 })),
      }),
    );

    await userEvent.type(screen.getByLabelText("New fact"), "duplicate");
    await userEvent.click(screen.getByRole("button", { name: "Remember" }));

    expect(
      await screen.findByText("Could not add the fact"),
    ).toBeInTheDocument();
    expect(screen.getByText(/already in memory/)).toBeInTheDocument();
  });

  it("returns to the empty form from the success screen", async () => {
    renderForm(createFakeApi({ addFact: () => Promise.resolve("s") }));

    await userEvent.type(screen.getByLabelText("New fact"), "fact");
    await userEvent.click(screen.getByRole("button", { name: "Remember" }));
    await screen.findByText("Fact added");

    await userEvent.click(screen.getByRole("button", { name: "Add another" }));

    expect(await screen.findByLabelText("New fact")).toHaveValue("");
  });

  it('delegates "Find similar" and "Back to home" to the injected callbacks', async () => {
    const onSearchSimilar = vi.fn();
    const onGoHome = vi.fn();
    renderForm(createFakeApi({ addFact: () => Promise.resolve("a summary") }), {
      onSearchSimilar,
      onGoHome,
    });

    await userEvent.type(screen.getByLabelText("New fact"), "fact");
    await userEvent.click(screen.getByRole("button", { name: "Remember" }));
    await screen.findByText("Fact added");

    await userEvent.click(screen.getByRole("button", { name: "Find similar" }));
    expect(onSearchSimilar).toHaveBeenCalledWith("a summary");

    await userEvent.click(screen.getByRole("button", { name: "Back to home" }));
    expect(onGoHome).toHaveBeenCalledTimes(1);
  });
});
