import { describe, expect, it } from "vitest";

import { ApiError } from "@/shared/api/ApiError";
import { createFactApi } from "@/shared/api/client";
import type { FetchLike } from "@/shared/api/client";

type RecordedRequest = {
  readonly method: string;
  readonly url: string;
  readonly body: string | undefined;
  readonly signal: AbortSignal;
};

/**
 * Tests run in Node, where `Request` rejects relative URLs. The production
 * client deliberately passes none (same-origin), so the seam is an explicit
 * base URL used only by these tests.
 */
const BASE_URL = "http://api.test";

/**
 * A fake transport: no sockets, no global `fetch`. It records the request the
 * client built and answers with an in-memory `Response`. The body must be read
 * from a clone, because `Request.body` is a stream.
 */
const createRecordingFetch = (
  response: () => Response,
): {
  readonly calls: RecordedRequest[];
  readonly fetchImpl: FetchLike;
} => {
  const calls: RecordedRequest[] = [];

  const fetchImpl: FetchLike = async (request) => {
    const body = await request.clone().text();
    calls.push({
      method: request.method,
      url: request.url,
      body: body === "" ? undefined : body,
      signal: request.signal,
    });
    return response();
  };

  return { calls, fetchImpl };
};

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });

describe("search", () => {
  it("POSTs the query to /api/v1/search and unwraps the envelope", async () => {
    const { calls, fetchImpl } = createRecordingFetch(() =>
      jsonResponse(200, {
        value: {
          found: [{ id: "1", createdAt: "2026-01-01T00:00:00Z", summary: "s" }],
        },
      }),
    );

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });
    const found = await api.search("what do I like");

    expect(found).toHaveLength(1);
    expect(found[0]?.summary).toBe("s");
    expect(calls[0]?.method).toBe("POST");
    expect(calls[0]?.url).toBe(`${BASE_URL}/api/v1/search`);
    expect(calls[0]?.body).toBe(JSON.stringify({ query: "what do I like" }));
  });

  it("throws a typed error on a documented failure", async () => {
    const { fetchImpl } = createRecordingFetch(() =>
      jsonResponse(500, { error: "EMPTY_EMBEDDINGS" }),
    );

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });

    await expect(api.search("abc")).rejects.toMatchObject({
      code: "EMPTY_EMBEDDINGS",
      status: 500,
    });
  });

  it("converts a transport throw into NETWORK_ERROR", async () => {
    const api = createFactApi({
      baseUrl: BASE_URL,
      fetchImpl: () => Promise.reject(new TypeError("Failed to fetch")),
    });

    const error = await api.search("abc").catch((caught: unknown) => caught);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ code: "NETWORK_ERROR" });
  });

  it("forwards an abort signal to the request", async () => {
    const { calls, fetchImpl } = createRecordingFetch(() =>
      jsonResponse(200, { value: { found: [] } }),
    );
    const controller = new AbortController();

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });
    await api.search("abc", { signal: controller.signal });

    expect(calls[0]?.signal).toBeInstanceOf(AbortSignal);
  });
});

describe("getFact", () => {
  it("puts the id into the path and unwraps {value}", async () => {
    const { calls, fetchImpl } = createRecordingFetch(() =>
      jsonResponse(200, {
        value: {
          id: "a1",
          createdAt: "2026-03-14T09:26:53Z",
          summary: "s",
          content: "c",
        },
      }),
    );

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });
    const fact = await api.getFact("a1");

    expect(fact.id).toBe("a1");
    expect(fact.content).toBe("c");
    expect(calls[0]?.method).toBe("GET");
    expect(calls[0]?.url).toBe(`${BASE_URL}/api/v1/fact/a1`);
  });

  it("encodes ids that contain slashes", async () => {
    const { calls, fetchImpl } = createRecordingFetch(() =>
      jsonResponse(200, {
        value: { id: "a/b", createdAt: "", summary: "", content: "" },
      }),
    );

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });
    await api.getFact("a/b");

    expect(calls[0]?.url).toBe(`${BASE_URL}/api/v1/fact/a%2Fb`);
  });

  it("surfaces a missing fact as a typed error", async () => {
    const { fetchImpl } = createRecordingFetch(() =>
      jsonResponse(404, { error: "FACT_DOES_NOT_EXIST" }),
    );

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });

    await expect(api.getFact("gone")).rejects.toMatchObject({
      code: "FACT_DOES_NOT_EXIST",
      status: 404,
    });
  });
});

describe("addFact", () => {
  it("sends the double-nested body the contract requires", async () => {
    const { calls, fetchImpl } = createRecordingFetch(() =>
      jsonResponse(200, { summary: "short" }),
    );

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });
    const summary = await api.addFact("I like coffee");

    expect(summary).toBe("short");
    expect(calls[0]?.url).toBe(`${BASE_URL}/api/v1/fact`);
    expect(calls[0]?.body).toBe(
      JSON.stringify({ value: { fact: "I like coffee" } }),
    );
  });

  it("surfaces ALREADY_HAS_FACT", async () => {
    const { fetchImpl } = createRecordingFetch(() =>
      jsonResponse(400, { error: "ALREADY_HAS_FACT" }),
    );

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });

    await expect(api.addFact("dup")).rejects.toMatchObject({
      code: "ALREADY_HAS_FACT",
    });
  });
});

describe("deleteFact", () => {
  it("sends DELETE and resolves on the empty success envelope", async () => {
    const { calls, fetchImpl } = createRecordingFetch(() =>
      jsonResponse(200, { value: {} }),
    );

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });
    await expect(api.deleteFact("a1")).resolves.toBeUndefined();

    expect(calls[0]?.method).toBe("DELETE");
    expect(calls[0]?.url).toBe(`${BASE_URL}/api/v1/fact/a1`);
    expect(calls[0]?.body).toBeUndefined();
  });

  it("rejects with a typed error when the backend refuses", async () => {
    const { fetchImpl } = createRecordingFetch(() =>
      jsonResponse(404, { error: "FACT_DOES_NOT_EXIST" }),
    );

    const api = createFactApi({ baseUrl: BASE_URL, fetchImpl });

    await expect(api.deleteFact("gone")).rejects.toMatchObject({
      code: "FACT_DOES_NOT_EXIST",
      status: 404,
    });
  });
});
