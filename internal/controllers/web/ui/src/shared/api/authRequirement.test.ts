import { afterEach, describe, expect, it } from "vitest";

import { readAuthRequirement } from "@/shared/api/authRequirement";

/**
 * The flag is written by the Go server into index.html, so the only thing the UI
 * has to get right is how it interprets the global.
 */
describe("readAuthRequirement", () => {
  afterEach(() => {
    Reflect.deleteProperty(window, "requireAuth");
  });

  const set = (value: unknown): void => {
    Object.defineProperty(window, "requireAuth", {
      value,
      configurable: true,
      writable: true,
    });
  };

  it("reads true when the server injected true", () => {
    set(true);

    expect(readAuthRequirement()).toBe(true);
  });

  it("reads false when the server injected false", () => {
    set(false);

    expect(readAuthRequirement()).toBe(false);
  });

  it("defaults to false when nothing was injected", () => {
    // A page served without the script — an older build, a static host — must stay
    // usable rather than lock the user out of an API that needs no token.
    expect(readAuthRequirement()).toBe(false);
  });

  it("defaults to false for anything that is not literally true", () => {
    // Only the boolean true opens the gate path; a string "true" from a hand-
    // written script must not silently mean "auth on".
    set("true");

    expect(readAuthRequirement()).toBe(false);
  });
});
