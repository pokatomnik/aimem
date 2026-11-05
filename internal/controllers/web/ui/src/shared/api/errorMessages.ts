import type { components } from "@/api/generated/schema";

/** Error codes declared by `openapi.yaml`. */
export type ContractErrorCode = components["schemas"]["errorCode"];

/** Failures that happen before the backend can answer with a code. */
export type LocalErrorCode = "NETWORK_ERROR" | "UNKNOWN";

export type ErrorCode = ContractErrorCode | LocalErrorCode;

/**
 * `satisfies` makes the compiler demand a message for every declared code:
 * adding an enum value to `openapi.yaml` breaks the build until it is handled.
 */
export const ERROR_MESSAGES = {
  DB_CONVERT_ERROR: "Storage error. Please try again",
  ALREADY_HAS_FACT: "This fact is already in memory",
  CANNOT_GET_MODELS: "Cannot reach the model API — check its configuration",
  GET_EMBEDDINGS: "Failed to compute embeddings. Try again",
  SUMMARIZE: "Failed to summarize the fact",
  EMBEDDINGS_QUANTITY_NOT_EQUAL_TO_FACTS_QUANTITY:
    "Storage returned inconsistent data. Try again",
  EMPTY_EMBEDDINGS: "Failed to compute embeddings for the query",
  FACT_DOES_NOT_EXIST: "Fact not found",
  BAD_REQUEST: "Invalid request",
  INTERNAL_ERROR: "Internal server error",
  NOT_FOUND: "Not found",
  UNAUTHORIZED: "Not authorized. Check the token in the `Authorization` header",
  NETWORK_ERROR: "Server is unreachable. Make sure `aimem web` is running",
  UNKNOWN: "Something went wrong",
} as const satisfies Record<ErrorCode, string>;

const KNOWN_CODES: ReadonlySet<string> = new Set(Object.keys(ERROR_MESSAGES));

export const isErrorCode = (value: unknown): value is ErrorCode =>
  typeof value === "string" && KNOWN_CODES.has(value);

export const getErrorMessage = (code: ErrorCode): string =>
  ERROR_MESSAGES[code];
