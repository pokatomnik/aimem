import { describe, expect, it } from "vitest";

import { formatDate, formatDateShort } from "./formatDate";

const CREATED_AT = "2026-03-14T09:26:53Z";

describe("formatDate", () => {
  it("renders a stable string regardless of the machine locale", () => {
    const formatted = formatDate(CREATED_AT);
    expect(formatted).toContain("2026");
    expect(formatted).not.toBe("—");
  });

  it("is deterministic for the same input", () => {
    expect(formatDate(CREATED_AT)).toBe(formatDate(CREATED_AT));
  });

  it("uses UTC, so the calendar day does not shift with the local zone", () => {
    // 23:30 UTC would already be the next day in a +02:00 zone.
    expect(formatDate("2026-03-14T23:30:00Z")).toContain("14");
  });
});

describe("formatDateShort", () => {
  it("renders a date without the time", () => {
    const formatted = formatDateShort(CREATED_AT);
    expect(formatted).toContain("2026");
    expect(formatted).not.toContain(":");
  });
});

describe("invalid input", () => {
  it("returns a dash instead of Invalid Date", () => {
    expect(formatDate("not-a-date")).toBe("—");
    expect(formatDateShort("")).toBe("—");
  });
});
