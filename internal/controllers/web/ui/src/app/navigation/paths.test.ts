import { describe, expect, it } from "vitest";

import {
  buildSearchPath,
  factPath,
  isSameSearch,
  readSearchQuery,
  ROUTES,
} from "./paths";

describe("ROUTES", () => {
  it("describes the SPA contract", () => {
    expect(ROUTES.home).toBe("/");
    expect(ROUTES.search).toBe("/search");
    expect(ROUTES.factAdd).toBe("/fact/add");
    expect(ROUTES.fact).toBe("/fact/:factID");
  });
});

describe("buildSearchPath", () => {
  it("encodes a space as plus, matching the required ?q=word+word shape", () => {
    expect(buildSearchPath("a b")).toBe("/search?q=a+b");
  });

  it("percent-encodes characters that would break the query string", () => {
    expect(buildSearchPath("a&b")).toBe("/search?q=a%26b");
    expect(buildSearchPath("a#b")).toBe("/search?q=a%23b");
    expect(buildSearchPath("a?b")).toBe("/search?q=a%3Fb");
    expect(buildSearchPath("a=b")).toBe("/search?q=a%3Db");
  });

  it("keeps a plain single word readable", () => {
    expect(buildSearchPath("hello")).toBe("/search?q=hello");
  });

  it("round-trips through URLSearchParams", () => {
    const path = buildSearchPath("meaning & context");
    const params = new URLSearchParams(path.slice(path.indexOf("?")));
    expect(params.get("q")).toBe("meaning & context");
  });
});

describe("factPath", () => {
  it("builds a concrete fact route", () => {
    expect(factPath("a1b2c3")).toBe("/fact/a1b2c3");
  });

  it("encodes unsafe id characters", () => {
    expect(factPath("a/b c")).toBe("/fact/a%2Fb%20c");
  });
});

describe("readSearchQuery", () => {
  it("decodes the plus-separated form the app generates", () => {
    expect(readSearchQuery("?q=a+b")).toBe("a b");
  });

  it("reports an empty string when q is absent", () => {
    expect(readSearchQuery("")).toBe("");
    expect(readSearchQuery("?foo=bar")).toBe("");
  });
});

describe("isSameSearch", () => {
  const onSearchPage = (
    query: string,
  ): { pathname: string; search: string } => ({
    pathname: ROUTES.search,
    search: `?q=${encodeURIComponent(query).replace(/%20/g, "+")}`,
  });

  it("is true for the query the page is already showing", () => {
    expect(isSameSearch(onSearchPage("coffee"), "coffee")).toBe(true);
  });

  it("ignores surrounding whitespace, so a stray space is not a new search", () => {
    expect(isSameSearch(onSearchPage("coffee"), "  coffee  ")).toBe(true);
  });

  it("is false for a different query", () => {
    expect(isSameSearch(onSearchPage("coffee"), "tea")).toBe(false);
  });

  // Coming from the landing page or the add-fact page always has to navigate.
  it("is false on any other route", () => {
    expect(isSameSearch({ pathname: ROUTES.home, search: "" }, "coffee")).toBe(
      false,
    );
    expect(
      isSameSearch({ pathname: ROUTES.factAdd, search: "?q=coffee" }, "coffee"),
    ).toBe(false);
  });

  it("treats a missing q as the empty query", () => {
    expect(isSameSearch({ pathname: ROUTES.search, search: "" }, "")).toBe(
      true,
    );
  });
});
