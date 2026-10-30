/**
 * Tokens end up in an HTTP header, and a header value is a byte string. `fetch`
 * throws a `TypeError` **before** sending anything when the value is not a valid
 * header value: a code point above U+00FF (so any Cyrillic or emoji) or a CR/LF
 * inside it. That throw looks exactly like a network failure at the transport
 * layer, so a token the browser cannot send would surface as "Server is
 * unreachable" — a message about the wrong thing entirely.
 *
 * Checking here keeps the failure where the user can act on it: the gate refuses
 * the token, and a token that reached storage anyway is not attached, which the
 * server answers with 401 and an honest message.
 */

/**
 * The rule the browser applies, stated once: every code point in Latin-1 range
 * and no line breaks. Deliberately not stricter than `fetch` — a token that can
 * be sent must be accepted, or the gate would reject a working credential.
 *
 * Iterating by code point (`for…of`) rather than by UTF-16 unit is what makes
 * the emoji case correct: an astral character is one code point above U+00FF,
 * while its halves are surrogates and would each pass a naive check.
 *
 * An empty string is not sendable: it is a missing token, and answering `false`
 * here means every caller gets the same answer for "nothing to send" as for
 * "cannot send it".
 */
export const isTokenSendable = (token: string): boolean => {
  if (token === "" || token.includes("\n") || token.includes("\r")) {
    return false;
  }

  for (const char of token) {
    if ((char.codePointAt(0) ?? 0x100) > 0xff) {
      return false;
    }
  }

  return true;
};
