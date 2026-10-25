import { describe, expect, it } from "vitest";

import { getErrorMessage, isErrorCode } from "@/shared/api/errorMessages";
import type { ContractErrorCode } from "@/shared/api/errorMessages";

/**
 * Every code declared in `openapi.yaml` must have a human-readable message.
 * If the spec gains a code, `ERROR_MESSAGES` fails to compile — this test then
 * guards against a code being added with an empty or placeholder string.
 */
const SPEC_CODES: readonly ContractErrorCode[] = [
  "DB_CONVERT_ERROR",
  "ALREADY_HAS_FACT",
  "CANNOT_GET_MODELS",
  "GET_EMBEDDINGS",
  "SUMMARIZE",
  "EMBEDDINGS_QUANTITY_NOT_EQUAL_TO_FACTS_QUANTITY",
  "EMPTY_EMBEDDINGS",
  "FACT_DOES_NOT_EXIST",
  "BAD_REQUEST",
  "INTERNAL_ERROR",
  "NOT_FOUND",
];

describe("error messages", () => {
  it("has a non-empty message for every declared code", () => {
    for (const code of SPEC_CODES) {
      expect(isErrorCode(code)).toBe(true);
      expect(getErrorMessage(code).length).toBeGreaterThan(0);
    }
  });

  it("covers local-only codes", () => {
    expect(getErrorMessage("NETWORK_ERROR")).toContain("unreachable");
    expect(getErrorMessage("UNKNOWN").length).toBeGreaterThan(0);
  });

  it("recognises only known codes", () => {
    expect(isErrorCode("ALREADY_HAS_FACT")).toBe(true);
    expect(isErrorCode("SOMETHING_ELSE")).toBe(false);
    expect(isErrorCode("")).toBe(false);
    expect(isErrorCode(42)).toBe(false);
    expect(isErrorCode(null)).toBe(false);
    expect(isErrorCode(undefined)).toBe(false);
  });

  it("rejects inherited object keys", () => {
    expect(isErrorCode("toString")).toBe(false);
    expect(isErrorCode("constructor")).toBe(false);
  });
});
