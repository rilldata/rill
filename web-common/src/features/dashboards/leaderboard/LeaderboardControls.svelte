<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { createShowHideDimensionsStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { MetricsViewMeasure } from "@rilldata/web-common/runtime-client";
  import { crossfade, fly } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";

  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";
  import { useMetaQuery } from "../selectors";
  import LeaderboardContextColumnMenu from "./LeaderboardContextColumnMenu.svelte";

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
  let activeLeaderboardMeasure: ReturnType<typeof formatForSelector>;
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

  $: validPercentOfTotal =
    activeLeaderboardMeasure?.validPercentOfTotal || false;

  // if the percent of total is currently being shown,
  // but it is not valid for this measure, then turn it off
  $: if (
    !validPercentOfTotal &&
    metricsExplorer?.leaderboardContextColumn ===
      LeaderboardContextColumn.PERCENT
  ) {
    metricsExplorerStore.setContextColumn(
      metricViewName,
      LeaderboardContextColumn.HIDDEN
    );
  }

  $: showHideDimensions = createShowHideDimensionsStore(
    metricViewName,
    metaQuery
  );

  const toggleDimensionVisibility = (e) => {
    showHideDimensions.toggleVisibility(e.detail.name);
  };
  const setAllDimensionsNotVisible = () => {
    showHideDimensions.setAllToNotVisible();
  };
  const setAllDimensionsVisible = () => {
    showHideDimensions.setAllToVisible();
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
        selectableItems={$showHideDimensions.selectableItems}
        selectedItems={$showHideDimensions.selectedItems}
        on:item-clicked={toggleDimensionVisibility}
        on:deselect-all={setAllDimensionsNotVisible}
        on:select-all={setAllDimensionsVisible}
        label="Dimensions"
        tooltipText="Choose dimensions to display"
      />

      <div class="whitespace-nowrap">showing</div>

      <SelectMenu
        paddingTop={2}
        paddingBottom={2}
        {options}
        {selection}
        alignment="end"
        on:select={handleMeasureUpdate}
      />

      <LeaderboardContextColumnMenu {metricViewName} {validPercentOfTotal} />
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
