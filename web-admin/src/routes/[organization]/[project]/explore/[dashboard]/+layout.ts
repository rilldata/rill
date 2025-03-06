import {
  fetchBookmarks,
  isHomeBookmark,
} from "@rilldata/web-admin/features/bookmarks/selectors";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors";

export const load = async ({ params, depends, parent }) => {
  const { user, project, runtime } = await parent();

  const { dashboard: exploreName } = params;

  depends(`explore:${exploreName}`);

  const exploreSpecPromise = fetchExploreSpec(runtime?.instanceId, exploreName);
  const homeBookmarkExploreStatePromise = user
    ? exploreSpecPromise.then(async ({ explore, metricsView, schema }) => {
        const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
        const exploreSpec = explore.explore?.state?.validSpec ?? {};

        const bookmarks = await fetchBookmarks(project.id, exploreName);
        const homeBookmark = bookmarks.find(isHomeBookmark);
        if (!homeBookmark) return <Partial<MetricsExplorerEntity>>{};
        return getDashboardStateFromUrl(
          homeBookmark.data ?? "",
          metricsViewSpec,
          exploreSpec,
          schema,
        );
      })
    : Promise.resolve(<Partial<MetricsExplorerEntity>>{});

  return {
    exploreSpecPromise,
    homeBookmarkExploreStatePromise,
  };
};
