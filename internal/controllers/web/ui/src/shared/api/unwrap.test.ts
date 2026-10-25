import { describe, expect, it } from "vitest";

import { ApiError, isAbortError } from "@/shared/api/ApiError";
import { unwrap } from "@/shared/api/unwrap";

const jsonResponse = (status: number): Response =>
  new Response(null, {
    status,
    headers: { "Content-Type": "application/json" },
  });

describe("unwrap", () => {
  it("returns the payload on success", () => {
    const result = { data: { value: "ok" }, response: jsonResponse(200) };
    expect(unwrap(result)).toEqual({ value: "ok" });
  });

  it("throws a typed error carrying the documented code and status", () => {
    const result = {
      error: { error: "FACT_DOES_NOT_EXIST", message: "missing" },
      response: jsonResponse(404),
    };

    try {
      unwrap(result);
      expect.unreachable("unwrap must throw");
    } catch (error) {
      expect(error).toBeInstanceOf(ApiError);
      const apiError = error as ApiError;
      expect(apiError.code).toBe("FACT_DOES_NOT_EXIST");
      expect(apiError.status).toBe(404);
      expect(apiError.message).toBe("Fact not found");
    }
  });

  it("degrades an unknown code to UNKNOWN but keeps the status", () => {
    const result = {
      error: { error: "NOT_IN_SPEC" },
      response: jsonResponse(500),
    };

    try {
      unwrap(result);
      expect.unreachable("unwrap must throw");
    } catch (error) {
      const apiError = error as ApiError;
      expect(apiError.code).toBe("UNKNOWN");
      expect(apiError.status).toBe(500);
    }
  });

  it("degrades a non-JSON error body to UNKNOWN", () => {
    const result = { error: undefined, response: jsonResponse(502) };

    try {
      unwrap(result);
      expect.unreachable("unwrap must throw");
    } catch (error) {
      expect((error as ApiError).code).toBe("UNKNOWN");
    }
  });
});

describe("ApiError.from", () => {
  it("passes an existing ApiError through unchanged", () => {
    const original = new ApiError("BAD_REQUEST", { status: 400 });
    expect(ApiError.from(original)).toBe(original);
  });

  it("maps a TypeError to a network failure", () => {
    expect(ApiError.from(new TypeError("Failed to fetch")).code).toBe(
      "NETWORK_ERROR",
    );
  });

  it("maps anything else to UNKNOWN", () => {
    expect(ApiError.from(new Error("boom")).code).toBe("UNKNOWN");
    expect(ApiError.from("string thrown").code).toBe("UNKNOWN");
  });
});

describe("isAbortError", () => {
  it("detects aborts by name, without needing a DOMException", () => {
    expect(isAbortError({ name: "AbortError" })).toBe(true);
    expect(isAbortError(new Error("AbortError"))).toBe(false);
    expect(isAbortError(null)).toBe(false);
    expect(isAbortError(undefined)).toBe(false);
  });
});
