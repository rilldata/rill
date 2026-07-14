<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getCanvasStore } from "../state-managers/state-managers";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import CanvasFilterChipsReadOnly from "../../dashboards/filters/CanvasFilterChipsReadOnly.svelte";

  export let canvasName: string;

  const runtimeClient = useRuntimeClient();

  $: ({ instanceId } = runtimeClient);

  $: ({
    canvasEntity: {
      clearDefaultFilters,
      filterManager: { defaultUIFiltersStore },
      timeManager: {
        state: { interval: _interval },
        defaultTimeRangeStore,
        defaultComparisonRangeStore,
      },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: interval = $_interval;

  $: defaultTimeRange = $defaultTimeRangeStore;

  $: defaultComparisonRange = $defaultComparisonRangeStore;
</script>

<div class="flex-col flex h-full">
  <div class="page-param">
    <p class="text-fg-secondary mb-4">
      The filters listed below are saved as your default view and will
      automatically apply each time you open this dashboard in Rill Cloud.
    </p>

    <CanvasFilterChipsReadOnly
      uiFilters={$defaultUIFiltersStore}
      timeRangeString={defaultTimeRange}
      comparisonRange={defaultComparisonRange}
      timeStart={interval?.start?.toISO()}
      timeEnd={interval?.end?.toISO()}
    />
  </div>

  <div class="mt-auto border-t w-full px-5 py-3">
    <Button type="secondary" wide onClick={clearDefaultFilters}>
      <Trash />
      Clear default filters
    </Button>
  </div>
</div>

<style lang="postcss">
  .page-param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
