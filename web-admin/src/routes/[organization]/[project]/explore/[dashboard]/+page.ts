import { convertBookmarkToUrlSearchParams } from "@rilldata/web-admin/features/bookmarks/selectors";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { getPartialExploreStateOrRedirect } from "@rilldata/web-common/features/explores/selectors";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ url, parent, params }) => {
  const { explore, metricsView, defaultExplorePreset, homeBookmark } =
    await parent();
  const { organization, project, dashboard: exploreName } = params;
  const metricsViewSpec = metricsView.metricsView?.state?.validSpec;
  const exploreSpec = explore.explore?.state?.validSpec;

  // On the first dashboard load, if there are no URL params, redirect to the "Home" bookmark.
  if (
    homeBookmark &&
    ![...url.searchParams.keys()].length &&
    !(exploreName in get(metricsExplorerStore).entities)
  ) {
    const newUrl = new URL(url);
    newUrl.search = convertBookmarkToUrlSearchParams(
      homeBookmark,
      metricsViewSpec,
      exploreSpec,
      {}, // TODO
      undefined,
      defaultExplorePreset,
    );

    if (newUrl.search !== url.search) {
      throw redirect(307, `${newUrl.pathname}${newUrl.search}`);
    }
  }

  const { partialExploreState, errors } = getPartialExploreStateOrRedirect(
    exploreName,
    metricsViewSpec,
    exploreSpec,
    defaultExplorePreset,
    `__${organization}__${project}`,
    url,
  );

  return {
    partialExploreState,
    errors,
  };
};
