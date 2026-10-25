import type { ReactNode } from "react";
import { Route, Routes } from "react-router";

import { useAppNavigation } from "@/app/navigation";
import { AddFactPage } from "@/pages/add-fact/AddFactPage";
import { FactPage } from "@/pages/fact/FactPage";
import { HomePage } from "@/pages/home/HomePage";
import { NotFoundPage } from "@/pages/not-found/NotFoundPage";
import { SearchPage } from "@/pages/search/SearchPage";

/**
 * The composition root for navigation: `useAppNavigation` is called once here
 * and every page receives plain callbacks and hrefs through props.
 *
 * `/fact/add` is declared before `/fact/:factID` on purpose: React Router ranks
 * static segments higher anyway, but the explicit order documents the intent.
 */
export const AppRoutes = (): ReactNode => {
  const nav = useAppNavigation();

  return (
    <Routes>
      <Route
        path={nav.homeHref}
        element={
          <HomePage
            onSearch={nav.goToSearch}
            addFactHref={nav.addFactHref}
            homeHref={nav.homeHref}
          />
        }
      />
      <Route
        path={nav.searchHref}
        element={
          <SearchPage
            onSearch={nav.goToSearch}
            factHref={nav.factHref}
            homeHref={nav.homeHref}
            addFactHref={nav.addFactHref}
          />
        }
      />
      <Route
        path={nav.addFactHref}
        element={
          <AddFactPage
            onBack={nav.goHome}
            onSearchSimilar={nav.goToSearch}
            homeHref={nav.homeHref}
          />
        }
      />
      <Route
        path={nav.factHrefPattern}
        element={<FactPage onBack={nav.goHome} homeHref={nav.homeHref} />}
      />
      <Route path="*" element={<NotFoundPage homeHref={nav.homeHref} />} />
    </Routes>
  );
};
