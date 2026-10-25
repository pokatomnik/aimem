import { useContext } from "react";

import type { FactApi } from "@/shared/api";
import { ApiContext } from "./ApiContext";

export const useApi = (): FactApi => {
  const api = useContext(ApiContext);
  if (api === null) {
    throw new Error("useApi must be used within <ApiProvider>");
  }
  return api;
};
