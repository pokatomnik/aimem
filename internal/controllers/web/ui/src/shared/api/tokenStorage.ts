/**
 * The token is the whole session: there is no login, no expiry and no logout.
 * Keeping it in one place means the storage key is never duplicated across the
 * client, the gate and the tests.
 */
const STORAGE_KEY = "aimem.token";

/**
 * Storage is injected so a test never touches `window.localStorage`. jsdom does
 * provide one, but relying on it would make every test share state.
 */
export type TokenStorage = {
  readonly get: () => string | undefined;
  readonly set: (token: string) => void;
};

/** `undefined` means "nothing stored"; an empty string is treated the same way. */
const normalize = (value: string | null): string | undefined =>
  value === null || value.trim() === "" ? undefined : value.trim();

export const createTokenStorage = (storage: Storage): TokenStorage => ({
  get: () => normalize(storage.getItem(STORAGE_KEY)),
  set: (token) => {
    storage.setItem(STORAGE_KEY, token.trim());
  },
});

/**
 * In-memory storage for tests and for an environment without `localStorage`.
 * It keeps the same contract, so the gate cannot tell the difference.
 */
export const createMemoryTokenStorage = (
  initial?: string,
): TokenStorage & { readonly current: () => string | undefined } => {
  let token = normalize(initial ?? null);

  return {
    get: () => token,
    set: (value) => {
      token = normalize(value);
    },
    current: () => token,
  };
};

/**
 * `window.localStorage` throws when storage is disabled (private mode, blocked
 * cookies). A token that cannot be persisted still works for the session, so the
 * fallback degrades to memory instead of failing the whole app.
 */
export const createDefaultTokenStorage = (): TokenStorage => {
  try {
    return createTokenStorage(window.localStorage);
  } catch {
    return createMemoryTokenStorage();
  }
};
