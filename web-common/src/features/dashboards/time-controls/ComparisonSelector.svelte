<script lang="ts">
  import { IconSpaceFixer, Switch } from "../../../components/button";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import { WithSelectMenu } from "../../../components/menu";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import { getComparisonOptionsForTimeRange } from "./time-range-utils";

  export let metricViewName;

  $: dashboardStore = useDashboardStore(metricViewName);

  $: selectedTimeRange = $dashboardStore?.selectedTimeRange;

  let comparisonOptions = [];
  $: if (selectedTimeRange) {
    comparisonOptions = getComparisonOptionsForTimeRange(selectedTimeRange);
  }

  // Comparison Switch
  $: isComparisonEnabled = $dashboardStore?.showComparison;
  const toggleComparison = () => {
    metricsExplorerStore.toggleComparison(metricViewName);
  };

  // Comparison Menu
  $: options = comparisonOptions.map((key) => ({
    main: key,
    key,
  }));
  $: selectedCompareRange =
    $dashboardStore.comparisonTimeRange || options[0]?.key || "None";

  const onCompareRangeSelect = (comparisonRange) => {
    metricsExplorerStore.setSelectedComparisonRange(
      metricViewName,
      comparisonRange
    );
  };
</script>

<div class="flex gap-x-2 flex-row items-center pl-3">
  <Switch on:click={() => toggleComparison()} checked={isComparisonEnabled}>
    Compare {#if isComparisonEnabled} to {/if}
  </Switch>

  {#if isComparisonEnabled}
    <WithSelectMenu
      distance={8}
      {options}
      disabled={!isComparisonEnabled}
      selection={{
        main: selectedCompareRange,
        key: selectedCompareRange,
      }}
      on:select={(event) => onCompareRangeSelect(event.detail.key)}
      let:toggleMenu
      let:active
    >
      <button
        class="px-3 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600"
        on:click={toggleMenu}
      >
        <span class="font-bold">{selectedCompareRange}</span>
        <IconSpaceFixer pullRight>
          <div class="transition-transform" class:-rotate-180={active}>
            <CaretDownIcon size="16px" />
          </div>
        </IconSpaceFixer>
      </button>
    </WithSelectMenu>
  {/if}
</div>
