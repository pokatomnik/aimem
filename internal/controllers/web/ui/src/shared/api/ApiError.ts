import { isRecord } from "@/shared/lib/guards";
import {
  getErrorMessage,
  isErrorCode,
  type ErrorCode,
} from "@/shared/api/errorMessages";

export type ApiErrorOptions = {
  readonly status?: number;
  readonly cause?: unknown;
};

/** Abort is control flow, not a failure: callers rethrow it untouched. */
export const isAbortError = (error: unknown): boolean =>
  isRecord(error) && error.name === "AbortError";

/**
 * A single failure type for the whole UI. The presentation layer only needs
 * `message`; `code` and `status` are available for special cases such as 404.
 */
export class ApiError extends Error {
  readonly code: ErrorCode;
  readonly status: number | undefined;

  constructor(code: ErrorCode, options: ApiErrorOptions = {}) {
    super(
      getErrorMessage(code),
      options.cause === undefined ? undefined : { cause: options.cause },
    );
    this.name = "ApiError";
    this.code = code;
    this.status = options.status;
  }

  /** Build from an `appError` body. Anything unparseable degrades to `UNKNOWN`. */
  static fromBody(body: unknown, status: number): ApiError {
    if (!isRecord(body)) {
      return new ApiError("UNKNOWN", { status });
    }

    const rawCode = body.error;
    const code: ErrorCode = isErrorCode(rawCode) ? rawCode : "UNKNOWN";
    return new ApiError(code, { status });
  }

  /** Normalise an unexpected throwable coming from the transport layer. */
  static from(cause: unknown): ApiError {
    if (cause instanceof ApiError) {
      return cause;
    }
    return new ApiError(
      cause instanceof TypeError ? "NETWORK_ERROR" : "UNKNOWN",
      { cause },
    );
  }
}
