import { describe, expect, it } from "vitest";

import { isTokenSendable } from "@/shared/api/token";

/**
 * The rule the browser enforces on a header value. Every case here was checked
 * against `fetch` itself: each "sendable: false" token makes it throw a
 * `TypeError` before a request leaves, which is the bug this guard prevents.
 */
describe("isTokenSendable", () => {
  it("accepts a plain ASCII token", () => {
    expect(isTokenSendable("57aa8cba-0ba5-47c1-8a96-9791e9576241")).toBe(true);
  });

  it("accepts Latin-1 letters, which the header encoding still carries", () => {
    // The gate must not be stricter than the browser: a token it rejects but
    // `fetch` would send is a working credential refused for no reason.
    expect(isTokenSendable("café-Ünïcode-ß")).toBe(true);
  });

  it("rejects Cyrillic", () => {
    expect(isTokenSendable("токен")).toBe(false);
  });

  it("rejects emoji", () => {
    expect(isTokenSendable("token-🔑")).toBe(false);
  });

  it("rejects a line break, which would end the header value", () => {
    expect(isTokenSendable("abc\ndef")).toBe(false);
    expect(isTokenSendable("abc\rdef")).toBe(false);
  });

  it("treats an empty token as unsendable", () => {
    expect(isTokenSendable("")).toBe(false);
  });
});
