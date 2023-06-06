<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { MetricsViewMeasure } from "@rilldata/web-common/runtime-client";
  import { crossfade, fly } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";
  import {
    selectBestDimensionStrings,
    selectDimensionKeys,
    useMetaQuery,
  } from "../selectors";

  export let metricViewName;

  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);

  $: measures = $metaQuery.data?.measures;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  function handleMeasureUpdate(event: CustomEvent) {
    metricsExplorerStore.setLeaderboardMeasureName(
      metricViewName,
      event.detail.key
    );
  }

  function formatForSelector(measure: MetricsViewMeasure) {
    if (!measure) return undefined;
    return {
      ...measure,
      key: measure.name,
      main: measure.label?.length ? measure.label : measure.expression,
    };
  }

  let [send, receive] = crossfade({ fallback: fly });

  /** this should be a single element */
  // reset selections based on the active leaderboard measure
  let activeLeaderboardMeasure;
  $: activeLeaderboardMeasure =
    measures?.length &&
    metricsExplorer?.leaderboardMeasureName &&
    formatForSelector(
      measures.find(
        (measure) => measure.name === metricsExplorer?.leaderboardMeasureName
      ) ?? undefined
    );

  /** this controls the animation direction */

  $: options =
    measures?.map((measure) => {
      let main = measure.label?.length ? measure.label : measure.expression;
      return {
        ...measure,
        key: measure.name,
        main,
      };
    }) || [];

  /** set the selection only if measures is not undefined */
  $: selection = measures ? activeLeaderboardMeasure : [];

  $: availableDimensionLabels = selectBestDimensionStrings($metaQuery);
  $: availableDimensionKeys = selectDimensionKeys($metaQuery);
  $: visibleDimensionKeys = metricsExplorer?.visibleDimensionKeys;
  $: visibleDimensionsBitmask = availableDimensionKeys.map((k) =>
    visibleDimensionKeys.has(k)
  );

  const toggleDimensionVisibility = (e) => {
    metricsExplorerStore.toggleDimensionVisibilityByKey(
      metricViewName,
      availableDimensionKeys[e.detail.index]
    );
  };
  const setAllDimensionsNotVisible = () => {
    metricsExplorerStore.hideAllDimensions(metricViewName);
  };
  const setAllDimensionsVisible = () => {
    metricsExplorerStore.setMultipleDimensionsVisible(
      metricViewName,
      availableDimensionKeys
    );
  };
</script>

<div>
  {#if measures && options.length && selection}
    <div
      class="flex flex-row items-center ui-copy-muted"
      style:padding-left="22px"
      style:grid-column-gap=".4rem"
      in:send={{ key: "leaderboard-metric" }}
      style:max-width="450px"
    >
      <SeachableFilterButton
        selectableItems={availableDimensionLabels}
        selectedItems={visibleDimensionsBitmask}
        on:item-clicked={toggleDimensionVisibility}
        on:deselect-all={setAllDimensionsNotVisible}
        on:select-all={setAllDimensionsVisible}
        label="Dimensions"
        tooltipText="Choose dimensions to display"
      />

      <div class="whitespace-nowrap">showing top values by</div>

      <SelectMenu
        paddingTop={2}
        paddingBottom={2}
        {options}
        {selection}
        tailwindClasses="overflow-hidden"
        alignment="end"
        on:select={handleMeasureUpdate}
      >
        <span class="font-bold truncate">{selection?.main}</span>
      </SelectMenu>
    </div>
  {:else}
    <div
      class="flex flex-row items-center"
      style:grid-column-gap=".4rem"
      in:receive={{ key: "loading-leaderboard-metric" }}
    >
      pulling leaderboards <Spinner status={EntityStatus.Running} />
    </div>
  {/if}
</div>
