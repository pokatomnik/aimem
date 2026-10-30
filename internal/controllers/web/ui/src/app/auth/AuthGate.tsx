import type { ReactNode } from "react";

import { useAuth, needsToken } from "@/app/auth/AuthContext";
import { TokenGate } from "@/app/auth/TokenGate";

/**
 * Blocks the app until a token is stored, and nothing more.
 *
 * It sits inside the router but outside the routes, so a page never has to know
 * about authentication: either it renders normally (no auth, or a token present)
 * or it is replaced by the gate.
 */
export const AuthGate = ({
  children,
}: {
  readonly children: ReactNode;
}): ReactNode => {
  const session = useAuth();

  if (needsToken(session)) {
    return <TokenGate onSave={session.saveToken} />;
  }

  return children;
};
