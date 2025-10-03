<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import PivotEmpty from "@rilldata/web-common/features/dashboards/pivot/PivotEmpty.svelte";
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import PivotTable from "@rilldata/web-common/features/dashboards/pivot/PivotTable.svelte";
  import {
    type PivotDataStore,
    type PivotDataStoreConfig,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";

  import type { Readable, Writable } from "svelte/store";

  export let schema: {
    isValid: boolean;
    error?: string;
  };
  export let pivotDataStore: PivotDataStore | undefined;
  export let pivotConfig: Readable<PivotDataStoreConfig> | undefined;
  export let pivotState: Writable<PivotState>;
  export let hasHeader = false;

  $: pivotColumns = splitPivotChips($pivotState.columns);

  $: hasColumnAndNoMeasure =
    pivotColumns.dimension.length > 0 && pivotColumns.measure.length === 0;
</script>

<div
  class="size-full overflow-hidden"
  style:max-height="inherit"
  class:p-4={hasHeader}
  class:pt-1={hasHeader}
>
  {#if !schema.isValid}
    <ComponentError error={schema.error} />
  {:else if pivotDataStore && $pivotDataStore && pivotConfig && $pivotConfig}
    {#if $pivotDataStore?.error?.length}
      <PivotError errors={$pivotDataStore.error} />
    {:else if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
      <PivotEmpty
        assembled={$pivotDataStore.assembled}
        isFetching={$pivotDataStore.isFetching}
        {hasColumnAndNoMeasure}
      />
    {:else}
      <PivotTable
        border={hasHeader}
        rounded={hasHeader}
        {pivotDataStore}
        config={pivotConfig}
        {pivotState}
        setPivotExpanded={(expanded) => {
          pivotState.update((state) => ({
            ...state,
            expanded,
          }));
        }}
        setPivotSort={(sorting) => {
          pivotState.update((state) => ({
            ...state,
            sorting,
            rowPage: 1,
            expanded: {},
          }));
        }}
        setPivotRowPage={(page) => {
          pivotState.update((state) => ({
            ...state,
            rowPage: page,
          }));
        }}
      />
    {/if}
  {/if}
</div>
