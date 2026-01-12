import { derived, get, readable, type Readable } from "svelte/store";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
import { getMetricsViewPickerOptions } from "@rilldata/web-common/features/chat/core/context/picker/data/metrics-views.ts";
import { getModelsPickerOptions } from "@rilldata/web-common/features/chat/core/context/picker/data/models.ts";
import { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";
import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/picker-tree.ts";

/**
 * Creates a store that contains a list of options for each valid metrics view and sources/models.
 * 1. Chooses top level options based on where this is run in rill developer or not.
 *    If rill dev, then sources/models are also included in the top level list along with metrics views.
 * 2. Bubbles up recently used and active top level options to the top of the list.
 *
 * The list contains parents immediately followed by their children.
 */
export function getPickerOptions(
  uiState: ContextPickerUIState,
): Readable<PickerItem[]> {
  const isRillDev = !get(featureFlags.adminServer);

  return derived(
    [
      getMetricsViewPickerOptions(),
      isRillDev ? getModelsPickerOptions(uiState) : readable(null),
      uiState.expandedParentsStore,
    ],
    ([metricsViewOptions, filesOption]) => {
      const recentlyUsed: PickerItem[] = [];
      const recentlyUsedIds = new Set<string>();
      const currentlyActive: PickerItem[] = [];
      const currentlyUsedIds = new Set<string>();

      const allOptions = [metricsViewOptions, filesOption]
        .filter(Boolean)
        .flat() as PickerItem[];

      // Mark items and all its children as recently used or currently active
      allOptions.forEach((o) => {
        const parentIsRecentlyUsed =
          o.parentId && recentlyUsedIds.has(o.parentId);
        if (o.recentlyUsed || parentIsRecentlyUsed) {
          recentlyUsedIds.add(o.id);
          recentlyUsed.push(o);
        }

        const parentIsActive = o.parentId && currentlyUsedIds.has(o.parentId);
        if (o.currentlyActive || parentIsActive) {
          currentlyUsedIds.add(o.id);
          currentlyActive.push(o);
        }
      });

      // Bubble up recently used and active items to the top of the list, including their children
      return [
        ...recentlyUsed,
        ...currentlyActive,
        ...allOptions.filter(
          (o) => !recentlyUsedIds.has(o.id) && !currentlyUsedIds.has(o.id),
        ),
      ];
    },
  );
}
