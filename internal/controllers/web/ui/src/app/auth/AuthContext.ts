import { createContext, useContext } from "react";

/**
 * Session state: whether this deployment needs a token, and the token itself.
 *
 * `required` and `token` are independent on purpose. With the embedded backend
 * `required` is false and a token, if one is stored, is irrelevant; with Qdrant
 * `required` is true and `token === undefined` is what makes the gate show.
 */
export type AuthSession = {
  readonly required: boolean;
  readonly token: string | undefined;
  readonly saveToken: (token: string) => void;
};

/** `null` means "no provider above": `useAuth` turns that into a loud error. */
export const AuthContext = createContext<AuthSession | null>(null);

export const useAuth = (): AuthSession => {
  const session = useContext(AuthContext);
  if (session === null) {
    throw new Error("useAuth must be used within <AuthProvider>");
  }
  return session;
};

/** Whether the gate should replace the app: a token is required and missing. */
export const needsToken = (session: AuthSession): boolean =>
  session.required && session.token === undefined;
