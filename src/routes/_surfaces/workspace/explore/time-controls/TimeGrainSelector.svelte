<script lang="ts">
  import type { TimeGrain } from "$common/database-service/DatabaseTimeSeriesActions";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "$lib/components/menu/wrappers/WithSelectMenu.svelte";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { Readable } from "svelte/store";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { selectTimeGrainApi } from "$lib/redux-store/explore/explore-apis";
  import { store } from "$lib/redux-store/store-root";
  import { prettyTimeGrain, TimeGrainOption } from "./time-range-utils";

  export let metricsDefId: string;

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  let selectableTimeGrains: TimeGrainOption[];
  $: selectableTimeGrains = $metricsExplorer?.selectableTimeGrains ?? [];

  let selectedTimeGrain: TimeGrain;
  $: selectedTimeGrain = $metricsExplorer?.selectedTimeGrain;

  $: options = selectableTimeGrains
    ? selectableTimeGrains.map(({ timeGrain, enabled }) => ({
        main: prettyTimeGrain(timeGrain),
        disabled: !enabled,
        key: timeGrain,
        description: !enabled ? "not valid for this time range" : undefined,
      }))
    : undefined;

  const onTimeGrainSelect = (timeGrain: TimeGrain) => {
    store.dispatch(selectTimeGrainApi({ metricsDefId, timeGrain }));
  };
</script>

{#if selectedTimeGrain && selectableTimeGrains}
  <WithSelectMenu
    {options}
    selection={{
      main: prettyTimeGrain(selectedTimeGrain),
      key: selectedTimeGrain,
    }}
    on:select={(event) => onTimeGrainSelect(event.detail.key)}
    let:toggleMenu
    let:active
  >
    <button
      class="px-4 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 transition-tranform duration-100"
      on:click={toggleMenu}
    >
      <span class="font-bold"
        >by {prettyTimeGrain(selectedTimeGrain)} increments</span
      >
      <span class="transition-transform" class:-rotate-180={active}>
        <CaretDownIcon size="16px" />
      </span>
    </button>
  </WithSelectMenu>
{/if}
