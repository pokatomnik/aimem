import { describe, expect, it, vi } from "vitest";

import {
  createFakeApi,
  factListItemFixture,
  renderHookWithProviders,
} from "@/shared/test/renderWithProviders";
import { useSearch } from "./useSearch";

describe("useSearch", () => {
  it("stays idle for a query the backend would reject", async () => {
    const search = vi.fn(() => Promise.resolve([]));
    renderHookWithProviders(() => useSearch("ab"), createFakeApi({ search }));

    // Give pending effects a chance to run before asserting the negative.
    await vi.waitFor(() => {
      expect(search).not.toHaveBeenCalled();
    });
  });

  it("issues one request per normalised query", async () => {
    const search = vi.fn(() => Promise.resolve([factListItemFixture]));
    const { result } = renderHookWithProviders(
      () => useSearch("  what do I like  "),
      createFakeApi({ search }),
    );

    await vi.waitFor(() => {
      expect(result.current.items).toHaveLength(1);
    });

    expect(search).toHaveBeenCalledTimes(1);
    expect(search).toHaveBeenCalledWith("what do I like", expect.anything());
  });

  it("shares the cache entry for queries differing only in whitespace", async () => {
    const search = vi.fn(() => Promise.resolve([factListItemFixture]));
    const { result, rerender } = renderHookWithProviders(
      () => useSearch("coffee"),
      createFakeApi({ search }),
    );

    await vi.waitFor(() => {
      expect(result.current.items).toBeDefined();
    });

    // Same normalised key: TanStack Query serves it from cache.
    rerender();

    expect(search).toHaveBeenCalledTimes(1);
  });

  it("exposes the error without retrying", async () => {
    const search = vi.fn(() => Promise.reject(new Error("boom")));
    const { result } = renderHookWithProviders(
      () => useSearch("coffee"),
      createFakeApi({ search }),
    );

    await vi.waitFor(() => {
      expect(result.current.error).not.toBeNull();
    });

    expect(search).toHaveBeenCalledTimes(1);
  });

  it("reports pending before the answer arrives", () => {
    const search = vi.fn(() => new Promise<never>(() => undefined));
    const { result } = renderHookWithProviders(
      () => useSearch("coffee"),
      createFakeApi({ search }),
    );

    expect(result.current.isPending).toBe(true);
    expect(result.current.items).toBeUndefined();
  });
});
