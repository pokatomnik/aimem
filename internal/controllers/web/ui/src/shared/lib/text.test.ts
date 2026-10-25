import { describe, expect, it } from "vitest";

import {
  FACT_MIN_LENGTH,
  formatResultCount,
  isFactValid,
  isSearchQueryValid,
  normalizeQuery,
  SEARCH_QUERY_MIN_LENGTH,
} from "./text";

describe("normalizeQuery", () => {
  it("trims leading and trailing whitespace", () => {
    expect(normalizeQuery("  hello  ")).toBe("hello");
  });

  it("leaves inner whitespace alone", () => {
    expect(normalizeQuery("hello   world")).toBe("hello   world");
  });
});

describe("isSearchQueryValid", () => {
  it("requires at least three characters after trimming", () => {
    expect(SEARCH_QUERY_MIN_LENGTH).toBe(3);
    expect(isSearchQueryValid("")).toBe(false);
    expect(isSearchQueryValid("ab")).toBe(false);
    expect(isSearchQueryValid("  ab  ")).toBe(false);
    expect(isSearchQueryValid("abc")).toBe(true);
  });

  it("accepts multi-word queries", () => {
    expect(isSearchQueryValid("what do I like")).toBe(true);
  });
});

describe("isFactValid", () => {
  it("rejects whitespace-only input", () => {
    expect(FACT_MIN_LENGTH).toBe(1);
    expect(isFactValid("")).toBe(false);
    expect(isFactValid("   \n\t ")).toBe(false);
    expect(isFactValid("a")).toBe(true);
  });
});

describe("formatResultCount", () => {
  it("uses the singular for exactly one result", () => {
    expect(formatResultCount(1)).toBe("1 result");
  });

  it("uses the plural for zero and many", () => {
    expect(formatResultCount(0)).toBe("0 results");
    expect(formatResultCount(2)).toBe("2 results");
    expect(formatResultCount(100)).toBe("100 results");
  });
});
