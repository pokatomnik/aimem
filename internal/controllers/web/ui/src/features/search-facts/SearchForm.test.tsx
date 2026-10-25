import { describe, expect, it, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { SearchForm } from "./SearchForm";

describe("SearchForm", () => {
  it("does not search while typing", async () => {
    const onSearch = vi.fn();
    render(<SearchForm onSearch={onSearch} />);

    await userEvent.type(screen.getByRole("searchbox"), "what do I like");

    expect(onSearch).not.toHaveBeenCalled();
  });

  it("searches on button click with the trimmed query", async () => {
    const onSearch = vi.fn();
    render(<SearchForm onSearch={onSearch} />);

    await userEvent.type(screen.getByRole("searchbox"), "  like coffee  ");
    await userEvent.click(screen.getByRole("button", { name: "Search" }));

    expect(onSearch).toHaveBeenCalledWith("like coffee");
  });

  it("searches on Enter as well", async () => {
    const onSearch = vi.fn();
    render(<SearchForm onSearch={onSearch} />);

    await userEvent.type(screen.getByRole("searchbox"), "coffee{Enter}");

    expect(onSearch).toHaveBeenCalledWith("coffee");
  });

  it("blocks queries shorter than three characters", async () => {
    const onSearch = vi.fn();
    render(<SearchForm onSearch={onSearch} />);

    await userEvent.type(screen.getByRole("searchbox"), "ab{Enter}");

    expect(onSearch).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Search" })).toBeDisabled();
  });

  // A caption that appeared mid-typing would push everything below it down, so
  // it must be part of the initial layout.
  it("shows the length hint before anything is typed", () => {
    render(<SearchForm onSearch={vi.fn()} />);

    expect(screen.getByText("At least 3 characters")).toBeInTheDocument();
  });

  it("keeps the hint visible for a valid query", async () => {
    render(<SearchForm onSearch={vi.fn()} />);

    await userEvent.type(screen.getByRole("searchbox"), "coffee");

    expect(screen.getByText("At least 3 characters")).toBeInTheDocument();
  });

  it("starts with the query it was given", () => {
    render(<SearchForm initialQuery="previous query" onSearch={vi.fn()} />);

    expect(screen.getByRole("searchbox")).toHaveValue("previous query");
  });

  it("shows the magnifier emoji on the button", () => {
    render(<SearchForm onSearch={vi.fn()} />);

    expect(
      screen.getByRole("button", { name: "Search" }).textContent,
    ).toContain("🔍");
  });

  it("calls the handler once per submit, not per keystroke", async () => {
    const onSearch = vi.fn();
    render(<SearchForm onSearch={onSearch} />);

    await act(async () => {
      await userEvent.type(screen.getByRole("searchbox"), "query{Enter}");
    });

    expect(onSearch).toHaveBeenCalledTimes(1);
  });

  // The pill is a horizontal flex row, so a hint rendered inside it would sit
  // beside the button instead of under the field. This pins the structure.
  it("keeps the hint outside the search form", () => {
    render(<SearchForm onSearch={vi.fn()} />);

    const form = screen.getByRole("search");
    const hint = screen.getByText("At least 3 characters");

    expect(form).not.toContainElement(hint);
  });

  it("clears the query with the clear button", async () => {
    render(<SearchForm onSearch={vi.fn()} />);

    const input = screen.getByRole("searchbox");
    await userEvent.type(input, "coffee");

    await userEvent.click(screen.getByRole("button", { name: "Clear query" }));

    expect(input).toHaveValue("");
    expect(
      screen.queryByRole("button", { name: "Clear query" }),
    ).not.toBeInTheDocument();
  });
});
