import createClient from "openapi-fetch";

import type { components, paths } from "@/api/generated/schema";
import { ApiError } from "@/shared/api/ApiError";
import { unwrap } from "@/shared/api/unwrap";

/** A fact as returned by `GET /api/v1/fact/{factID}`. */
export type Fact = components["schemas"]["fact"];

/** A single search hit: `id`, `createdAt` and `summary`, no content. */
export type FactListItem = components["schemas"]["searchResultItem"];

type SearchResponse = { readonly value: components["schemas"]["searchResult"] };
type FactResponse = { readonly value: Fact };
type AddFactResponse = { readonly value: components["schemas"]["factAddedOK"] };

/** Signature of a fetch implementation; injectable so tests never touch the network. */
export type FetchLike = (input: Request) => Promise<Response>;

export type FactApiConfig = {
  /**
   * Only for tests: Node's `Request` rejects relative URLs. In the browser the
   * client calls `/api/v1/...` against the current host, with no base URL.
   */
  readonly baseUrl?: string;
  readonly fetchImpl?: FetchLike;
};

export type RequestOptions = {
  readonly signal?: AbortSignal | undefined;
};

/**
 * The only boundary that knows about HTTP, JSON envelopes and error bodies.
 * Everything above it works with domain types and `ApiError`.
 */
export type FactApi = {
  readonly search: (
    query: string,
    options?: RequestOptions,
  ) => Promise<FactListItem[]>;
  readonly getFact: (id: string, options?: RequestOptions) => Promise<Fact>;
  readonly addFact: (fact: string, options?: RequestOptions) => Promise<string>;
  readonly deleteFact: (id: string, options?: RequestOptions) => Promise<void>;
};

const withSignal = (
  signal: AbortSignal | undefined,
): { signal?: AbortSignal } => (signal === undefined ? {} : { signal });

/**
 * No base URL on purpose: the UI and the API are always served by the same
 * binary on the same host, so requests go to the current origin by definition.
 */
export const createFactApi = (config: FactApiConfig = {}): FactApi => {
  const client = createClient<paths>({
    ...(config.baseUrl === undefined ? {} : { baseUrl: config.baseUrl }),
    ...(config.fetchImpl === undefined ? {} : { fetch: config.fetchImpl }),
  });

  // Transport failures (DNS, connection refused, abort) arrive as throws, while
  // documented failures arrive as `error` payloads. Normalise both to `ApiError`.
  const run = async <T>(request: () => Promise<T>): Promise<T> => {
    try {
      return await request();
    } catch (error) {
      throw ApiError.from(error);
    }
  };

  return {
    search: async (query, options) =>
      run(async () => {
        const body = unwrap<SearchResponse>(
          await client.POST("/api/v1/search", {
            body: { query },
            ...withSignal(options?.signal),
          }),
        );
        return body.value.found;
      }),

    getFact: async (id, options) =>
      run(async () => {
        const body = unwrap<FactResponse>(
          await client.GET("/api/v1/fact/{factID}", {
            params: { path: { factID: id } },
            ...withSignal(options?.signal),
          }),
        );
        return body.value;
      }),

    addFact: async (fact, options) =>
      run(async () => {
        const body = unwrap<AddFactResponse>(
          await client.POST("/api/v1/fact", {
            body: { value: { fact } },
            ...withSignal(options?.signal),
          }),
        );
        return body.value.summary;
      }),

    deleteFact: async (id, options) =>
      run(async () => {
        // Success is `200 { "value": {} }`: only the status matters here.
        unwrap<unknown>(
          await client.DELETE("/api/v1/fact/{factID}", {
            params: { path: { factID: id } },
            ...withSignal(options?.signal),
          }),
        );
      }),
  };
};
