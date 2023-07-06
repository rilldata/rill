<script lang="ts">
  import { SelectMenu } from "@rilldata/web-common/components/menu";
  import SeachableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SeachableFilterButton.svelte";
  import { createShowHideDimensionsStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { MetricsViewMeasure } from "@rilldata/web-common/runtime-client";
  import { crossfade, fly } from "svelte/transition";
  import { runtime } from "../../../runtime-client/runtime-store";
  import Spinner from "../../entity-management/Spinner.svelte";
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import PieChart from "@rilldata/web-common/components/icons/PieChart.svelte";
  import {
    ButtonGroup,
    SubButton,
  } from "@rilldata/web-common/components/button-group";

  import { useModelHasTimeSeries } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../dashboard-stores";
  import { useMetaQuery } from "../selectors";

  export let metricViewName;

  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);

  $: measures = $metaQuery.data?.measures;

  $: hasTimeSeriesQuery = useModelHasTimeSeries(
    $runtime.instanceId,
    metricViewName
  );
  $: hasTimeSeries = $hasTimeSeriesQuery?.data;

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

  let disabledButtons: ("delta" | "pie")[] = [];
  $: {
    disabledButtons = [];
    if (!hasTimeSeries) disabledButtons.push("delta");
    if (activeLeaderboardMeasure?.validPercentOfTotal !== true)
      disabledButtons.push("pie");
  }

  let selectedButton: "delta" | "pie" | null = null;
  // NOTE: time comparison takes precedence over percent of total
  $: selectedButton = metricsExplorer?.showComparison
    ? "delta"
    : metricsExplorer?.showPercentOfTotal
    ? "pie"
    : null;

  const handleContextValueButtonGroupClick = (evt) => {
    // console.log("handleContextValueButtonGroupClick", evt.detail);
    const value = evt.detail;
    if (value === "delta" && selectedButton == "delta") {
      metricsExplorerStore.displayComparison(metricViewName, false);
    } else if (value === "delta" && selectedButton != "delta") {
      metricsExplorerStore.displayComparison(metricViewName, true);
    } else if (value === "pie" && selectedButton == "pie") {
      metricsExplorerStore.displayPercentOfTotal(metricViewName, false);
    } else if (value === "pie" && selectedButton != "pie") {
      metricsExplorerStore.displayPercentOfTotal(metricViewName, true);
    }
  };

  $: selectedButtons = selectedButton === null ? [] : [selectedButton];

  const pieTooltips = {
    selected: "Hide percent of total",
    unselected: "Show percent of total",
    disabled: "To show percent of total, show top values by a summable metric",
  };

  const deltaTooltips = {
    selected: "Hide percent change",
    unselected: "Show percent change",
    disabled: "To show percent change, select a comparison period above",
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

      <ButtonGroup
        selected={selectedButtons}
        disabled={disabledButtons}
        on:subbutton-click={handleContextValueButtonGroupClick}
      >
        <SubButton value={"delta"} tooltips={deltaTooltips}>
          <Delta />%
        </SubButton>
        <SubButton value={"pie"} tooltips={pieTooltips}>
          <PieChart />%
        </SubButton>
      </ButtonGroup>
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
