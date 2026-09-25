<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import type { LeaderboardSpec } from "@rilldata/web-common/features/canvas/components/leaderboard";
  import DimensionFiltersInput from "@rilldata/web-common/features/canvas/inspector/filters/DimensionFiltersInput.svelte";
  import TimeFiltersInput from "@rilldata/web-common/features/canvas/inspector/filters/TimeFiltersInput.svelte";
  import type { BaseCanvasComponent } from "../../components/BaseCanvasComponent";
  import { resolveTimeFilters } from "../../components/time-filters";
  import type {
    ComponentFilterProperties,
    ComponentSpec,
  } from "../../components/types";
  import type { AllKeys, FilterInputParam } from "../types";

  let { component }: { component: BaseCanvasComponent } = $props();

  let {
    specStore,
    type,
    localExpressionFilters,
    localTimeFilters,
    metricsViewsProvider,
    yamlConfigProvider,
  } = $derived(component);

  let localParamValues = $derived($specStore);

  let inputParams = $derived(component.inputParams().filter);

  let metricsView = $derived(
    "metrics_view" in localParamValues
      ? (localParamValues.metrics_view ?? null)
      : null,
  );

  let excludedDimensions = $derived(
    type === "leaderboard"
      ? Object.fromEntries(
          (localParamValues as LeaderboardSpec).dimensions.map((d) => [
            d,
            true,
          ]),
        )
      : {},
  );

  let entries = $derived(
    Object.entries(inputParams) as [AllKeys<ComponentSpec>, FilterInputParam][],
  );

  let { hasTimeSeries } = $derived(localTimeFilters);

  // Mirrors the chip visibility check in ComponentHeader so the switch only
  // appears when the component would actually show local filter chips.
  let filterProperties = $derived(
    localParamValues as ComponentFilterProperties,
  );
  let { hasLocalTimeRange, comparison } = $derived(
    resolveTimeFilters(filterProperties.time_filters),
  );
  let hasLocalFilter = $derived(
    Boolean(filterProperties.dimension_filters) ||
      hasLocalTimeRange ||
      comparison.mode === "local",
  );

  let hideLocalFilters = $derived(Boolean(filterProperties.hide_local_filters));
</script>

<div>
  {#each entries as [key, config] (key)}
    <div class="component-param">
      {#if config.type === "time_filters"}
        {#if hasTimeSeries}
          <TimeFiltersInput
            id={key}
            {localTimeFilters}
            {metricsViewsProvider}
            {yamlConfigProvider}
            showComparison={config?.meta?.hasComparison}
            updateLocalTimeFilterString={(newString) => {
              component.updateProperty("time_filters", newString);
            }}
          />
        {/if}
      {:else if config.type == "dimension_filters" && metricsView}
        <DimensionFiltersInput
          {localExpressionFilters}
          updateLocalFilterString={(newString) => {
            component.updateProperty("dimension_filters", newString);
          }}
          {excludedDimensions}
          id={key}
        />
      {/if}
    </div>
  {/each}
  {#if hasLocalFilter}
    <div class="component-param flex flex-col gap-y-2">
      <div class="flex justify-between">
        <InputLabel
          capitalize={false}
          small
          label={m.canvas_show_local_filters_label()}
          id="hide_local_filters"
          faint={hideLocalFilters}
        />
        <Switch
          checked={!hideLocalFilters}
          onCheckedChange={(next) => {
            component.updateProperty(
              "hide_local_filters",
              next ? undefined : true,
            );
          }}
          small
        />
      </div>
      <div class="text-fg-secondary">
        {m.canvas_show_local_filters_hint()}
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  .component-param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
