import { describe, expect, it } from "vitest";

import { buildSearchPath, factPath, ROUTES } from "./paths";

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
