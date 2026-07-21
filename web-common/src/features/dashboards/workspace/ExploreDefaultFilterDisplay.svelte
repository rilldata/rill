<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    V1ExploreComparisonMode,
    type V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { get } from "svelte/store";
  import { parseDocument } from "yaml";
  import { getStateManagers } from "../state-managers/state-managers";
  import ExploreFilterChipsReadOnly from "../filters/ExploreFilterChipsReadOnly.svelte";
  import { getExploreFilterStateFromYAMLConfig } from "../stores/get-explore-state-from-yaml-config";

  let {
    fileArtifact,
    autoSave,
  }: {
    fileArtifact: FileArtifact;
    autoSave: boolean;
  } = $props();

  const { metricsViewName, validSpecStore } = getStateManagers();

  let exploreSpec = $derived($validSpecStore?.data?.explore);
  let defaultPreset = $derived(exploreSpec?.defaultPreset);
  let hasDefaults = $derived(
    !!(
      defaultPreset?.measures?.length ||
      defaultPreset?.dimensions?.length ||
      defaultPreset?.timeRange ||
      defaultPreset?.comparisonMode ||
      defaultPreset?.exploreSortBy ||
      defaultPreset?.where ||
      defaultPreset?.pinnedFilters?.length ||
      defaultPreset?.requiredFilters?.length
    ),
  );

  let filterState = $derived(
    exploreSpec ? getExploreFilterStateFromYAMLConfig(exploreSpec) : {},
  );

  let pinnedFilters = $derived(
    new Set([
      ...(defaultPreset?.pinnedFilters ?? []),
      ...(defaultPreset?.requiredFilters ?? []),
    ]),
  );

  let comparisonTimeRange = $derived(
    getComparisonTimeRange(
      defaultPreset?.comparisonMode,
      defaultPreset?.compareTimeRange,
    ),
  );

  function getComparisonTimeRange(
    mode: string | undefined,
    compareTimeRange: string | undefined,
  ): V1TimeRange | undefined {
    if (mode === V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME) {
      // Fall back to rill-PP when comparison_mode is "time" with no explicit range
      return { expression: compareTimeRange ?? "rill-PP" };
    }
    return undefined;
  }

  function clearDefaults() {
    const doc = parseDocument(get(fileArtifact.editorContent) ?? "");
    doc.delete("defaults");
    fileArtifact.updateEditorContent(doc.toString(), false, autoSave);
  }
</script>

<div class="flex-col flex h-full">
  {#if hasDefaults}
    <div class="page-param">
      <p class="text-fg-secondary mb-4">
        The filters listed below are saved as your default view and will
        automatically apply each time you open this dashboard in Rill Cloud.
      </p>

      <ExploreFilterChipsReadOnly
        metricsViewNames={[$metricsViewName]}
        filters={filterState.whereFilter}
        dimensionsWithInlistFilter={filterState.dimensionsWithInlistFilter ??
          []}
        dimensionThresholdFilters={filterState.dimensionThresholdFilters ?? []}
        displayTimeRange={{ expression: defaultPreset?.timeRange }}
        displayComparisonTimeRange={comparisonTimeRange}
        {pinnedFilters}
      />
    </div>

    <div class="mt-auto border-t w-full px-5 py-3">
      <Button type="secondary" wide onClick={clearDefaults}>
        <Trash />
        Clear default filters
      </Button>
    </div>
  {:else}
    <div class="page-param">
      <p class="text-fg-secondary mb-4">
        No default state configured. Use the "Save as default" button in the
        header to save the current dashboard state as the default view.
      </p>
    </div>
  {/if}
</div>

<style lang="postcss">
  .page-param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
