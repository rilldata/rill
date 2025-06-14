<script lang="ts">
  import type { LeaderboardSpec } from "@rilldata/web-common/features/canvas/components/leaderboard";
  import DimensionFiltersInput from "@rilldata/web-common/features/canvas/inspector/filters/DimensionFiltersInput.svelte";
  import TimeFiltersInput from "@rilldata/web-common/features/canvas/inspector/filters/TimeFiltersInput.svelte";
  import type { BaseCanvasComponent } from "../../components/BaseCanvasComponent";
  import type { ComponentSpec } from "../../components/types";
  import type { AllKeys, FilterInputParam } from "../types";

  export let component: BaseCanvasComponent;

  $: ({
    specStore,
    type,
    state: componentStore,
    parent: { name: canvasName },
  } = component);

  $: localParamValues = $specStore;

  $: inputParams = component.inputParams().filter;

  $: metricsView =
    "metrics_view" in localParamValues ? localParamValues.metrics_view : null;

  $: excludedDimensions =
    type === "leaderboard"
      ? (localParamValues as LeaderboardSpec).dimensions
      : [];

  $: entries = Object.entries(inputParams) as [
    AllKeys<ComponentSpec>,
    FilterInputParam,
  ][];
</script>

<div>
  {#each entries as [key, config] (key)}
    <div class="component-param">
      {#if config.type === "time_filters"}
        <TimeFiltersInput
          {canvasName}
          id={key}
          {componentStore}
          timeFilter={localParamValues[key]}
          showComparison={config?.meta?.hasComparison}
          showGrain={config?.meta?.hasGrain}
          onChange={async (filter) => {
            localParamValues[key] = filter;
            component.updateProperty(key, localParamValues[key]);
          }}
        />
      {:else if config.type == "dimension_filters" && metricsView}
        <DimensionFiltersInput
          {canvasName}
          {metricsView}
          {componentStore}
          {excludedDimensions}
          id={key}
          filter={localParamValues[key]}
          onChange={async (filter) => {
            localParamValues[key] = filter;
            component.updateProperty(key, localParamValues[key]);
          }}
        />
      {/if}
    </div>
  {/each}
</div>

<style lang="postcss">
  .component-param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
