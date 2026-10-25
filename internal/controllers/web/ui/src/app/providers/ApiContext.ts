import { createContext } from "react";

import type { FactApi } from "@/shared/api";

/** `null` means "no provider above": `useApi` turns that into a loud error. */
export const ApiContext = createContext<FactApi | null>(null);
