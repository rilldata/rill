<script lang="ts">
  import type { AdminServiceIssueMagicAuthTokenBody } from "@rilldata/web-admin/client";
  import ExploreFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/ExploreFilterChipsReadOnly.svelte";
  import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    getExploreFields,
    getSanitizedExploreStateParam,
    hasDashboardDimensionThresholdFilter,
    hasDashboardWhereFilter,
  } from "./form-utils";

  export let onFilterStateChange: (hasFilters: boolean) => void;
  export let onProvideFilters: (
    provider: () => Partial<AdminServiceIssueMagicAuthTokenBody>,
  ) => void;

  const StateManagers = getStateManagers();
  const dashboardStore = StateManagers.dashboardStore;
  const metricsViewName = StateManagers.metricsViewName;
  const visibleMeasures = StateManagers.selectors.measures.visibleMeasures;
  const visibleDimensions =
    StateManagers.selectors.dimensions.visibleDimensions;
  const validSpecStore = StateManagers.validSpecStore;
  const timeControlStore = useTimeControlStore(StateManagers);

  $: exploreFields = getExploreFields(
    $dashboardStore,
    $visibleDimensions,
    $visibleMeasures,
  );

  $: sanitizedState = getSanitizedExploreStateParam(
    $dashboardStore,
    exploreFields,
    $validSpecStore.data?.explore,
  );

  $: hasWhereFilter = !!hasDashboardWhereFilter($dashboardStore);
  $: hasDimensionThresholdFilter =
    !!hasDashboardDimensionThresholdFilter($dashboardStore);
  $: hasSomeFilter = hasWhereFilter || hasDimensionThresholdFilter;

  // Notify parent of filter state changes
  $: onFilterStateChange(hasSomeFilter);

  // Provide filter data to parent
  $: onProvideFilters(() => {
    const filter = hasSomeFilter
      ? mergeDimensionAndMeasureFilters(
          $dashboardStore.whereFilter,
          $dashboardStore.dimensionThresholdFilters,
        )
      : undefined;
    const metricsViewFilters = filter
      ? { [$metricsViewName]: filter }
      : undefined;

    return {
      resourceType: ResourceKind.Explore as string,
      metricsViewFilters,
      fields: exploreFields,
      state: sanitizedState ? sanitizedState : undefined,
    };
  });
</script>

{#if hasSomeFilter}
  <hr class="mt-4 mb-4" />

  <div class="flex flex-col gap-y-1">
    <p class="text-xs text-fg-primary font-normal">
      The following filters will be locked and hidden:
    </p>
    <div class="flex flex-row gap-1 mt-2">
      <ExploreFilterChipsReadOnly
        metricsViewNames={[$metricsViewName]}
        filters={$dashboardStore.whereFilter}
        dimensionsWithInlistFilter={$dashboardStore.dimensionsWithInlistFilter}
        dimensionThresholdFilters={$dashboardStore.dimensionThresholdFilters}
        queryTimeStart={$timeControlStore.timeStart}
        queryTimeEnd={$timeControlStore.timeEnd}
      />
    </div>
  </div>

  <p class="text-xs text-fg-primary font-normal mt-4 mb-4">
    Measures and dimensions will be limited to current visible set.
  </p>
{/if}
