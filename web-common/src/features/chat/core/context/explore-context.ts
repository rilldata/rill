import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { useStableExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { isExpressionEmpty } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { createStableTimeControlStoreFromName } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import type { V1AnalystAgentContext } from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";

/**
 * Creates a store that contains the active explore context sent to the Complete API.
 * It returns the V1AnalystAgentContext that is passed to the API.
 */
export function getActiveExploreContext(): Readable<
  Partial<V1AnalystAgentContext>
> {
  const exploreNameStore = getExploreNameStore();

  const exploreState = useStableExploreState(exploreNameStore);
  const timeControlsStore =
    createStableTimeControlStoreFromName(exploreNameStore);

  return derived(
    [exploreNameStore, exploreState, timeControlsStore],
    ([exploreName, exploreState, timeControlsStore]) => {
      const context: V1AnalystAgentContext = {
        explore: exploreName,
      };

      if (timeControlsStore?.timeStart && timeControlsStore?.timeEnd) {
        context.timeStart = timeControlsStore.timeStart;
        context.timeEnd = timeControlsStore.timeEnd;
      }

      const filterIsAvailable = !isExpressionEmpty(exploreState?.whereFilter);
      if (filterIsAvailable) {
        context.where = exploreState?.whereFilter;
      }

      return context;
    },
  );
}
