import {
  ChatContextEntryType,
  ContextTypeData,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import { formatV1TimeRange } from "@rilldata/web-common/features/chat/core/context/formatters.ts";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { useStableExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { isExpressionEmpty } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { createStableTimeControlStoreFromName } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import type { RuntimeServiceCompleteBody } from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";

export function getExploreContext() {
  const exploreNameStore = getExploreNameStore();

  const exploreState = useStableExploreState(exploreNameStore);
  const timeControlsStore =
    createStableTimeControlStoreFromName(exploreNameStore);

  return derived(
    [exploreNameStore, exploreState, timeControlsStore],
    ([exploreName, exploreState, timeControlsStore]) => {
      let context: RuntimeServiceCompleteBody =
        ContextTypeData[ChatContextEntryType.Explore].serializer(exploreName);

      const timeRange = formatV1TimeRange({
        start: timeControlsStore?.timeStart,
        end: timeControlsStore?.timeEnd,
      });
      context = {
        ...context,
        ...ContextTypeData[ChatContextEntryType.TimeRange].serializer(
          timeRange,
        ),
      };

      const filterIsAvailable = !isExpressionEmpty(exploreState?.whereFilter);
      if (filterIsAvailable) {
        context = {
          ...context,
          ...ContextTypeData[ChatContextEntryType.Where].serializer(
            exploreState.whereFilter,
          ),
        };
      }

      return context;
    },
  );
}
