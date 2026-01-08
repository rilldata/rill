import { createQuery } from "@tanstack/svelte-query";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { derived, get, readable, type Readable } from "svelte/store";
import { InlineContextType } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { InlineContextPickerParentOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
import { getMetricsViewPickerOptions } from "@rilldata/web-common/features/chat/core/context/picker/data/metrics-views.ts";
import { getModelsPickerOptions } from "@rilldata/web-common/features/chat/core/context/picker/data/models.ts";
import { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";

/**
 * Creates a store that contains a 2-level list of options for each valid metrics view and sources/models.
 * 1. Chooses top level options based on where this is run in rill developer or not.
 *    If rill dev, then sources/models are also included in the top level list along with metrics views.
 * 2. Any asynchronous options for 2nd level lists are fetched based on the open status of the top level option.
 * 3. Active metrics view and active source/model are tracked separately and filled in the resolved options.
 */
export function getPickerOptions(
  uiState: ContextPickerUIState,
): Readable<InlineContextPickerParentOption[]> {
  const isRillDev = !get(featureFlags.adminServer);

  return derived(
    [
      getMetricsViewPickerOptions(),
      isRillDev ? getModelsPickerOptions(uiState) : readable(null),
    ],
    ([metricsViewOptions, filesOption], set) => {
      const allOptions = [metricsViewOptions, filesOption]
        .filter(Boolean)
        .flat() as InlineContextPickerParentOption[];

      // Create a list of stores for children queries.
      const subOptionStores = allOptions.map((o) =>
        o.childrenQueryOptions
          ? createQuery(o.childrenQueryOptions, queryClient)
          : readable(null),
      );

      // Convert queries for children to children from query results.
      const resolvedOptionsStore = derived(
        subOptionStores,
        (subOptionStoresResp) => {
          const resolvedOptions = new Array<InlineContextPickerParentOption>(
            subOptionStoresResp.length,
          );
          subOptionStoresResp.forEach((subOptionStore, i) => {
            resolvedOptions[i] = {
              ...allOptions[i],
              children: subOptionStore?.data ?? allOptions[i].children ?? [],
              childrenLoading: subOptionStore?.isPending ?? false,
            };
          });
          return resolvedOptions;
        },
      );

      return resolvedOptionsStore.subscribe((resolvedOptions) =>
        set(resolvedOptions),
      );
    },
  );
}

export const ParentPickerTypes = new Set([
  InlineContextType.MetricsView,
  InlineContextType.Model,
]);
