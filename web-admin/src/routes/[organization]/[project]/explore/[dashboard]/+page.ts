import { getExploreStates } from "@rilldata/web-common/features/explores/selectors";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";

export const load = async ({ url, parent, params }) => {
  const { exploreSpecPromise } = await parent();
  const { organization, project, dashboard: exploreName } = params;

  const searchParams = url.searchParams;
  const exploreStatePromise = exploreSpecPromise.then(
    ({ explore, metricsView, defaultExplorePreset }) => {
      const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};
      const exploreSpec = explore.explore?.state?.validSpec ?? {};

      return getExploreStates(
        exploreName,
        `${organization}__${project}__`,
        searchParams,
        metricsViewSpec,
        exploreSpec,
        defaultExplorePreset,
      );
    },
  );

  return {
    exploreName,
    exploreStatePromise,
  };
};
