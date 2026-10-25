import { useNavigate } from "react-router";

import { buildSearchPath, factPath, ROUTES } from "./paths";

/**
 * Every navigation the app is allowed to perform, in one place. Pages receive
 * these callbacks through props instead of reaching for `useNavigate`, which
 * keeps the set of reachable destinations small and explicit.
 */
export type AppNavigation = {
  readonly goHome: () => void;
  readonly goToAddFact: () => void;
  readonly goToSearch: (query: string) => void;
  readonly goToFact: (id: string) => void;
  /** `true` when the current location is the landing page. */
  readonly homeHref: string;
  readonly addFactHref: string;
  readonly searchHref: string;
  readonly factHrefPattern: string;
  readonly factHref: (id: string) => string;
  readonly searchHrefFor: (query: string) => string;
};

/**
 * The only place `useNavigate` is allowed. `Routes` calls it once and passes
 * the resulting object down as `props`.
 */
export const useAppNavigation = (): AppNavigation => {
  const navigate = useNavigate();

  const goHome = (): void => {
    void navigate(ROUTES.home);
  };

  return {
    goHome,
    goToAddFact: () => {
      void navigate(ROUTES.factAdd);
    },
    goToSearch: (query) => {
      void navigate(buildSearchPath(query));
    },
    goToFact: (id) => {
      void navigate(factPath(id));
    },
    homeHref: ROUTES.home,
    addFactHref: ROUTES.factAdd,
    searchHref: ROUTES.search,
    factHrefPattern: ROUTES.fact,
    factHref: factPath,
    searchHrefFor: buildSearchPath,
  };
};
