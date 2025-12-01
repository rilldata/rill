import { page } from "$app/stores";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { useStableExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { isExpressionEmpty } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { createStableTimeControlStoreFromName } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import type {
  RuntimeServiceCompleteBody,
  V1AnalystAgentContext,
} from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";

/**
 * Creates a store that contains the active explore context sent to the Complete API.
 * It returns RuntimeServiceCompleteBody with V1AnalystAgentContext that is passed to the API.
 */
export function getActiveExploreContext(): Readable<
  Partial<RuntimeServiceCompleteBody>
> {
  const exploreNameStore = getExploreNameStore();

  const exploreState = useStableExploreState(exploreNameStore);
  const timeControlsStore =
    createStableTimeControlStoreFromName(exploreNameStore);

  return derived(
    [exploreNameStore, exploreState, timeControlsStore],
    ([exploreName, exploreState, timeControlsStore]) => {
      const analystAgentContext: V1AnalystAgentContext = {
        explore: exploreName,
      };

      if (timeControlsStore?.timeStart && timeControlsStore?.timeEnd) {
        analystAgentContext.timeStart = timeControlsStore.timeStart;
        analystAgentContext.timeEnd = timeControlsStore.timeEnd;
      }

      const filterIsAvailable = !isExpressionEmpty(exploreState?.whereFilter);
      if (filterIsAvailable) {
        analystAgentContext.where = exploreState?.whereFilter;
      }

      return <RuntimeServiceCompleteBody>{
        analystAgentContext,
      };
    },
  );
}

export function getActiveFileContext(): Readable<
  Partial<RuntimeServiceCompleteBody>
> {
  return derived(page, (pageState) => {
    const filePath = pageState.params?.file;
    return <RuntimeServiceCompleteBody>{
      developerAgentContext: {
        currentFilePath: filePath,
      },
    };
  });
}
