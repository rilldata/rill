<script lang="ts">
  import SearchableFilterButton from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterButton.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import { createShowHideDimensionsStore } from "@rilldata/web-common/features/dashboards/show-hide-selectors";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { metricsExplorerStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { useMetricsView } from "../selectors";
  import { getStateManagers } from "../state-managers/state-managers";
  import * as Select from "@rilldata/web-common/components/select";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export let metricViewName: string;

  const {
    selectors: {
      measures: {
        filteredSimpleMeasures,
        leaderboardMeasureName,
        getMeasureByName,
      },
    },
    actions: {
      contextCol: { setContextColumn },
      setLeaderboardMeasureName,
    },
  } = getStateManagers();

  let active = false;

  $: metricsView = useMetricsView($runtime.instanceId, metricViewName);

  $: measures = $filteredSimpleMeasures();

  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  $: activeLeaderboardMeasure = $getMeasureByName($leaderboardMeasureName);

  $: validPercentOfTotal =
    activeLeaderboardMeasure?.validPercentOfTotal || false;

  // if the percent of total is currently being shown,
  // but it is not valid for this measure, then turn it off
  $: if (
    !validPercentOfTotal &&
    metricsExplorer?.leaderboardContextColumn ===
      LeaderboardContextColumn.PERCENT
  ) {
    setContextColumn(LeaderboardContextColumn.HIDDEN);
  }

  $: showHideDimensions = createShowHideDimensionsStore(
    metricViewName,
    metricsView,
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
  {#if measures.length && activeLeaderboardMeasure}
    <div
      class="flex flex-row items-center ui-copy-muted gap-x-0.5"
      style:max-width="450px"
    >
      <SearchableFilterButton
        selectableItems={$showHideDimensions.selectableItems}
        selectedItems={$showHideDimensions.selectedItems}
        on:item-clicked={toggleDimensionVisibility}
        on:deselect-all={setAllDimensionsNotVisible}
        on:select-all={setAllDimensionsVisible}
        label="Dimensions"
        tooltipText="Choose dimensions to display"
      />

      <Select.Root
        bind:open={active}
        items={measures.map((measure) => ({
          value: measure.name ?? "",
          label: measure.label ?? measure.name,
        }))}
        onSelectedChange={(newSelection) => {
          if (!newSelection) return;
          setLeaderboardMeasureName(newSelection.value);
        }}
      >
        <Select.Trigger class="outline-none border-none w-fit  px-0 gap-x-0.5">
          <Button type="text" label="Select a measure to filter by">
            <span class="truncate text-gray-700 hover:text-inherit">
              Showing <b>
                {activeLeaderboardMeasure?.label ??
                  activeLeaderboardMeasure.name}
              </b>
            </span>
          </Button>
        </Select.Trigger>

        <Select.Content
          sameWidth={false}
          align="start"
          class="max-h-80 overflow-y-auto"
        >
          {#each measures as measure (measure.name)}
            <Select.Item
              value={measure.name}
              label={measure.label ?? measure.name}
              class="text-[12px] flex flex-col items-start"
            >
              <div class:font-bold={$leaderboardMeasureName === measure.name}>
                {measure.label ?? measure.name}
              </div>

              <p class="ui-copy-muted" style:font-size="11px">
                {measure.description}
              </p>
            </Select.Item>
          {/each}
        </Select.Content>
      </Select.Root>
    </div>
  {/if}
</div>
