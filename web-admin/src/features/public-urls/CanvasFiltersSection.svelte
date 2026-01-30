<script lang="ts">
  import type { AdminServiceIssueMagicAuthTokenBody } from "@rilldata/web-admin/client";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import type {
    FilterManager,
    UIFilters,
  } from "@rilldata/web-common/features/canvas/stores/filter-manager";
  import CanvasFilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/CanvasFilterChipsReadOnly.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { derived, type Readable } from "svelte/store";
  import {
    getCanvasFilters,
    getCanvasStateUrl,
    hasCanvasFilters,
  } from "./canvas-form-utils";

  export let dashboard: string;
  export let instanceId: string;
  export let onFilterStateChange: (hasFilters: boolean) => void;
  export let onProvideFilters: (
    provider: () => Partial<AdminServiceIssueMagicAuthTokenBody>,
  ) => void;
  let canvasEntity: CanvasEntity | undefined;
  let canvasFilterManager: FilterManager | undefined;
  let canvasActiveUIFiltersStore: Readable<UIFilters> | undefined;
  let canvasAppliedFiltersStore: Readable<UIFilters> | undefined;

  try {
    const canvasStore = getCanvasStore(dashboard, instanceId);
    canvasEntity = canvasStore.canvasEntity;
    canvasFilterManager = canvasEntity.filterManager;
    canvasActiveUIFiltersStore = canvasFilterManager.activeUIFiltersStore;

    // Create a derived store that only includes filters with actual values (not just pinned ones)
    canvasAppliedFiltersStore = derived(
      canvasActiveUIFiltersStore,
      ($activeFilters) => {
        const appliedDimensionFilters = new Map();
        const appliedMeasureFilters = new Map();

        // Filter dimension filters that have selected values or input text
        $activeFilters.dimensionFilters.forEach((filter, key) => {
          const hasValues =
            (filter.selectedValues && filter.selectedValues.length > 0) ||
            (filter.inputText && filter.inputText.trim() !== "");
          if (hasValues) {
            appliedDimensionFilters.set(key, filter);
          }
        });

        // Filter measure filters that have a filter expression
        $activeFilters.measureFilters.forEach((filter, key) => {
          if (filter.filter) {
            appliedMeasureFilters.set(key, filter);
          }
        });

        return {
          dimensionFilters: appliedDimensionFilters,
          measureFilters: appliedMeasureFilters,
          complexFilters: $activeFilters.complexFilters,
          hasFilters:
            appliedDimensionFilters.size > 0 ||
            appliedMeasureFilters.size > 0 ||
            $activeFilters.complexFilters.length > 0,
          hasClearableFilters:
            appliedDimensionFilters.size > 0 || appliedMeasureFilters.size > 0,
        } as UIFilters;
      },
    );
  } catch (e) {
    console.error("Failed to get canvas store:", e);
  }

  $: canvasFilters = canvasEntity ? getCanvasFilters(canvasEntity) : undefined;
  $: canvasStateUrl = getCanvasStateUrl(new URL(window.location.href));
  $: hasSomeFilter = canvasEntity ? hasCanvasFilters(canvasEntity) : false;

  // Notify parent of filter state changes
  $: onFilterStateChange(hasSomeFilter);

  // Provide filter data to parent
  $: onProvideFilters(() => ({
    resourceType: ResourceKind.Canvas as string,
    metricsViewFilters: canvasFilters,
    fields: undefined, // Grant full access to all fields
    state: canvasStateUrl || undefined,
  }));
</script>

{#if hasSomeFilter}
  <hr class="mt-4 mb-4" />

  <div class="flex flex-col gap-y-1">
    <p class="text-xs text-fg-primary font-normal">
      The following filters will be locked:
    </p>
    {#if canvasAppliedFiltersStore}
      <div class="flex flex-col gap-2 my-2">
        <CanvasFilterChipsReadOnly
          uiFilters={$canvasAppliedFiltersStore}
          col={false}
        />
      </div>
    {/if}
  </div>
{/if}
