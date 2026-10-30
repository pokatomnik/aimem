/**
 * Reads the flag the server injects into `index.html` before the entry module
 * runs. There is no endpoint for it: the answer is a property of the running
 * binary, so the server writes it into the document that boots the app.
 *
 * The declaration is on `window` because the script the server emits is a plain
 * assignment — nothing escapes, nothing to parse.
 */
declare global {
  interface Window {
    readonly requireAuth?: boolean;
  }
}

/**
 * `false` is the safe default for anything that is not literally `true`: a build
 * served without the injection (an old copy, a static host) must stay usable
 * rather than lock the user out of an API that needs no token.
 */
export const readAuthRequirement = (): boolean =>
  window.requireAuth === true;
