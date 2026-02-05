<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { V1ExploreComparisonMode } from "@rilldata/web-common/runtime-client";
  import { get } from "svelte/store";
  import { parseDocument } from "yaml";
  import { getStateManagers } from "../state-managers/state-managers";
  import ExploreFilterChipsReadOnly from "../filters/ExploreFilterChipsReadOnly.svelte";
  import {
    getExploreFilterStateFromYAMLConfig,
    getExploreStateFromYAMLConfig,
  } from "../stores/get-explore-state-from-yaml-config";

  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;

  const { metricsViewName, validSpecStore } = getStateManagers();

  $: exploreSpec = $validSpecStore?.data?.explore;
  $: defaultPreset = exploreSpec?.defaultPreset;
  $: hasDefaults = !!(
    defaultPreset?.measures?.length ||
    defaultPreset?.dimensions?.length ||
    defaultPreset?.timeRange ||
    defaultPreset?.comparisonMode ||
    defaultPreset?.exploreSortBy ||
    defaultPreset?.filter?.expression ||
    defaultPreset?.pinned?.length
  );

  $: filterState = getExploreFilterStateFromYAMLConfig(exploreSpec!);

  $: comparisonLabel = getComparisonLabel(defaultPreset?.comparisonMode);

  function getComparisonLabel(mode: string | undefined): string | undefined {
    switch (mode) {
      case V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME:
        return "Time comparison";
      case V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION:
        return `Dimension comparison: ${defaultPreset?.comparisonDimension ?? ""}`;
      default:
        return undefined;
    }
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
        displayComparisonTimeRange={{
          expression: defaultPreset?.compareTimeRange,
        }}
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

  .default-item {
    @apply flex flex-col gap-y-0.5;
  }

  .label {
    @apply text-xs text-fg-muted font-medium;
  }

  .value {
    @apply text-sm text-fg-primary;
  }
</style>
