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
});
