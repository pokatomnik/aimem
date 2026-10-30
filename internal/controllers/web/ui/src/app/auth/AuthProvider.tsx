import { useMemo, useState } from "react";
import type { ReactNode } from "react";

import { AuthContext } from "@/app/auth/AuthContext";
import type { AuthSession } from "@/app/auth/AuthContext";
import type { TokenStorage } from "@/shared/api/tokenStorage";

export type AuthProviderProps = {
  readonly required: boolean;
  readonly storage: TokenStorage;
  readonly children: ReactNode;
};

/**
 * The token is read once on mount and written on save. It is not re-read on
 * every render: the only writer is this provider, so storage and state cannot
 * drift apart.
 */
export const AuthProvider = ({
  required,
  storage,
  children,
}: AuthProviderProps): ReactNode => {
  const [token, setToken] = useState<string | undefined>(() => storage.get());

  const session = useMemo<AuthSession>(
    () => ({
      required,
      token,
      saveToken: (next: string) => {
        storage.set(next);
        setToken(storage.get());
      },
    }),
    [required, storage, token],
  );

  return (
    <AuthContext.Provider value={session}>{children}</AuthContext.Provider>
  );
};
