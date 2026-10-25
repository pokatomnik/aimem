import type { ReactNode } from "react";

import type { FactApi } from "@/shared/api";
import { ApiContext } from "./ApiContext";

type ApiProviderProps = {
  readonly api: FactApi;
  readonly children: ReactNode;
};

/**
 * Injects the HTTP boundary. Tests pass a fake, so no request ever leaves jsdom.
 */
export const ApiProvider = ({ api, children }: ApiProviderProps): ReactNode => (
  <ApiContext.Provider value={api}>{children}</ApiContext.Provider>
);
