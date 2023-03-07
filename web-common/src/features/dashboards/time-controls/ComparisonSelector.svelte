<script lang="ts">
  import { IconSpaceFixer } from "../../../components/button";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import { WithSelectMenu } from "../../../components/menu";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import {
    getComparisonTimeRange,
    prettyFormatTimeRange,
  } from "./time-range-utils";

  export let metricViewName;
  export let comparisonOptions;

  $: dashboardStore = useDashboardStore(metricViewName);

  // Comparison Menu
  $: options = comparisonOptions.map((key) => ({
    main: key,
    key,
  }));

  $: selectedTimeRange = $dashboardStore?.selectedTimeRange;
  $: selectedCompareRange =
    $dashboardStore?.comparisonTimeRange || options[0]?.key;

  $: datePrettyString = prettyFormatTimeRange(
    getComparisonTimeRange(selectedTimeRange, selectedCompareRange)
  );
  const onCompareRangeSelect = (comparisonRange) => {
    metricsExplorerStore.setSelectedComparisonRange(
      metricViewName,
      comparisonRange
    );
  };
</script>

<div class="flex gap-x-2 flex-row items-center pl-3">
  <div>Compare to</div>

  <WithSelectMenu
    distance={8}
    {options}
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
      <span>{datePrettyString}</span>
      <IconSpaceFixer pullRight>
        <div class="transition-transform" class:-rotate-180={active}>
          <CaretDownIcon size="16px" />
        </div>
      </IconSpaceFixer>
    </button>
  </WithSelectMenu>
</div>
