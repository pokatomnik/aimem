import { ApiError } from "@/shared/api/ApiError";

/**
 * Shape shared by every `openapi-fetch` call: either `data` or `error`, always
 * a `Response`. Declared structurally so tests can feed plain objects.
 */
export type ApiResult<TData> = {
  readonly data?: TData | undefined;
  readonly error?: unknown;
  readonly response: Response;
};

/** Returns the successful payload or throws a typed `ApiError`. */
export const unwrap = <TData>(result: ApiResult<TData>): TData => {
  if (result.data !== undefined) {
    return result.data;
  }
  throw ApiError.fromBody(result.error, result.response.status);
};
